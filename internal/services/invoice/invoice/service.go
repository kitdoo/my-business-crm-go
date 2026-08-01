// Package invoice implements the invoice.Service interface: renders a Sale
// as an invoice PDF (seller letterhead + buyer + line items + totals +
// commercial terms, matching the layout of a typical TOM STUDIO invoice)
// and optionally emails it.
package invoice

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/go-pdf/fpdf"

	"github.com/altessa-s/go-atlas/domain/normalizer"
	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	"github.com/kitdoo/my-business-crm-go/internal/pkg/appconfig"
	"github.com/kitdoo/my-business-crm-go/internal/pkg/mailer"
	clientsvc "github.com/kitdoo/my-business-crm-go/internal/services/client"
	invoicesvc "github.com/kitdoo/my-business-crm-go/internal/services/invoice"
	partnersvc "github.com/kitdoo/my-business-crm-go/internal/services/partner"
	productsvc "github.com/kitdoo/my-business-crm-go/internal/services/product"
	skusvc "github.com/kitdoo/my-business-crm-go/internal/services/productsku"
	variantsvc "github.com/kitdoo/my-business-crm-go/internal/services/productvariant"
	salesvc "github.com/kitdoo/my-business-crm-go/internal/services/sale"
)

//go:embed assets/DejaVuSansCondensed.ttf
var regularFontBytes []byte

//go:embed assets/DejaVuSansCondensed-Bold.ttf
var boldFontBytes []byte

var _ invoicesvc.Service = (*Service)(nil)

// Service is the invoice.Service implementation. clients/partners/skus/
// variants/products are the respective entities' Service, not their
// Storage — see SERVICE_DEVELOPMENT_STANDARD.md's "A service controls only
// its own storage" rule. There is no storage of its own: Generate is
// stateless.
type Service struct {
	sales    salesvc.Service
	clients  clientsvc.Service
	partners partnersvc.Service
	skus     skusvc.Service
	variants variantsvc.Service
	products productsvc.Service
	mailer   mailer.Service
	cfg      *appconfig.Config
	logger   *slog.Logger
}

// New builds a Service.
func New(
	sales salesvc.Service,
	clients clientsvc.Service,
	partners partnersvc.Service,
	skus skusvc.Service,
	variants variantsvc.Service,
	products productsvc.Service,
	mailerSvc mailer.Service,
	cfg *appconfig.Config,
) *Service {
	return &Service{
		sales:    sales,
		clients:  clients,
		partners: partners,
		skus:     skus,
		variants: variants,
		products: products,
		mailer:   mailerSvc,
		cfg:      cfg,
		logger:   slog.Default().With(slogx.Module("service:invoice")),
	}
}

// buyer is the resolved, invoice-ready projection of either a Client or a
// Partner — Generate never needs to know which one it came from past this
// point.
type buyer struct {
	Name               string
	Address            string
	Email              string
	TaxID              string
	RegistrationNumber string
	Code               string
}

// line is one invoice row, already converted from wire (basis
// points/hundredths) to plain display amounts.
type line struct {
	Code               string
	Description        string
	Unit               string
	Quantity           float64
	UnitPrice          float64
	DiscountPercentage int32
	NetAmount          float64 // after discount, before VAT
	VATAmount          float64 // 0 when VAT is not included
	GrossAmount        float64
}

func (s *Service) Generate(ctx context.Context, in *entities.InvoiceGenerate) ([]byte, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	invCfg := s.cfg.CRM.Invoice
	if invCfg == nil {
		return nil, errs.ErrInvoiceNotConfigured
	}

	docType := in.DocumentType
	if docType == entities.InvoiceDocumentTypeUnspecified {
		docType = entities.InvoiceDocumentTypeOffer
	}

	sl, err := s.sales.Get(ctx, in.SaleID)
	if err != nil {
		s.logger.DebugContext(ctx, "get sale failed", slog.String("id", in.SaleID), slog.String("error", err.Error()))
		return nil, err
	}

	// A Receipt only makes sense once the sale has actually been paid —
	// Draft is the only status Create can produce, so this is the one gate
	// that matters (every other status implies payment or is terminal for
	// other reasons).
	if docType == entities.InvoiceDocumentTypeReceipt && sl.Status == entities.SaleStatusDraft {
		return nil, errs.ErrInvoiceReceiptRequiresPayment
	}

	b, err := s.resolveBuyer(ctx, sl, in)
	if err != nil {
		return nil, err
	}

	lines, err := s.buildLines(ctx, sl.Items, in.IncludeVAT, invCfg.VAT.Percentage)
	if err != nil {
		return nil, err
	}

	number := invoiceNumber(docType, sl, invCfg)
	data, err := s.render(docType, number, sl, b, lines, in.IncludeVAT, invCfg)
	if err != nil {
		return nil, err
	}

	if in.SendEmail {
		s.sendEmail(ctx, docType, number, sl.ID, b, invCfg.Company, in.EmailOverride, data)
	}

	return data, nil
}

// resolveBuyer loads the Sale's Client (or, for a partner-only sale, its
// Partner), merges any non-empty override from in onto it via that
// entity's own Update (so the legal fields are remembered for next time —
// TD decision: invoice-create persists them rather than treating them as
// one-off input), and returns the invoice-ready projection.
func (s *Service) resolveBuyer(ctx context.Context, sl *entities.Sale, in *entities.InvoiceGenerate) (buyer, error) {
	hasOverride := in.BuyerTaxID != "" || in.BuyerRegistrationNumber != "" || in.BuyerCode != "" || in.BuyerAddress != nil

	if sl.ClientID != "" {
		c, err := s.clients.Get(ctx, sl.ClientID)
		if err != nil {
			return buyer{}, err
		}
		if hasOverride {
			update := &entities.ClientUpdate{ID: c.ID, Etag: &c.Etag}
			if in.BuyerTaxID != "" {
				update.TaxID = &in.BuyerTaxID
			}
			if in.BuyerRegistrationNumber != "" {
				update.RegistrationNumber = &in.BuyerRegistrationNumber
			}
			if in.BuyerCode != "" {
				update.Code = &in.BuyerCode
			}
			if in.BuyerAddress != nil {
				update.Address = in.BuyerAddress
			}
			updated, err := s.clients.Update(ctx, update)
			if err != nil {
				s.logger.WarnContext(ctx, "persist buyer legal fields onto client failed; invoice still uses the submitted values",
					slog.String("clientId", c.ID), slog.String("error", err.Error()))
			} else {
				c = updated
			}
		}
		return buyer{
			Name: c.Name, Address: c.Address, Email: c.Email,
			TaxID: c.TaxID, RegistrationNumber: c.RegistrationNumber, Code: c.Code,
		}, nil
	}

	if sl.PartnerID != nil {
		p, err := s.partners.Get(ctx, *sl.PartnerID)
		if err != nil {
			return buyer{}, err
		}
		if hasOverride {
			update := &entities.PartnerUpdate{ID: p.ID, Etag: &p.Etag}
			if in.BuyerTaxID != "" {
				update.TaxID = &in.BuyerTaxID
			}
			if in.BuyerRegistrationNumber != "" {
				update.RegistrationNumber = &in.BuyerRegistrationNumber
			}
			if in.BuyerCode != "" {
				update.Code = &in.BuyerCode
			}
			if in.BuyerAddress != nil {
				update.Address = in.BuyerAddress
			}
			updated, err := s.partners.Update(ctx, update)
			if err != nil {
				s.logger.WarnContext(ctx, "persist buyer legal fields onto partner failed; invoice still uses the submitted values",
					slog.String("partnerId", p.ID), slog.String("error", err.Error()))
			} else {
				p = updated
			}
		}
		return buyer{
			Name: p.Name, Address: p.Address, Email: p.Email,
			TaxID: p.TaxID, RegistrationNumber: p.RegistrationNumber, Code: p.Code,
		}, nil
	}

	return buyer{}, errs.ErrInvoiceSaleHasNoBuyer
}

// buildLines resolves each Sale item's SKU -> Variant -> Product to build
// its description/unit, and recomputes the line's net amount the same way
// sale.Service.buildItems did at Create time (PriceAmount and Quantity are
// both wire-scale — basis points / hundredths — hence the double /100).
func (s *Service) buildLines(ctx context.Context, items []entities.SaleItem, includeVAT bool, vatPercent int) ([]line, error) {
	locale := s.cfg.CRM.DefaultLocale
	out := make([]line, 0, len(items))
	for _, item := range items {
		sku, err := s.skus.Get(ctx, item.SKUID)
		if err != nil {
			return nil, err
		}
		variant, err := s.variants.Get(ctx, sku.VariantID)
		if err != nil {
			return nil, err
		}
		product, err := s.products.Get(ctx, variant.ProductID)
		if err != nil {
			return nil, err
		}

		netBasis := item.PriceAmount * item.Quantity * int64(100-item.DiscountPercentage) / 100 / 100
		net := float64(netBasis) / 100
		var vatAmount float64
		if includeVAT {
			vatAmount = net * float64(vatPercent) / 100
		}

		out = append(out, line{
			Code:               sku.SKU,
			Description:        buildDescription(locale, product, variant, sku),
			Unit:               unitLabel(product.PriceUnit),
			Quantity:           float64(item.Quantity) / 100,
			UnitPrice:          float64(item.PriceAmount) / 100,
			DiscountPercentage: item.DiscountPercentage,
			NetAmount:          net,
			VATAmount:          vatAmount,
			GrossAmount:        net + vatAmount,
		})
	}
	return out, nil
}

// buildDescription composes "<product name>, <variant attrs>; <sku attrs>"
// (e.g. "Rockface Stone, Andes White (3.0); 1200/600/4"), attribute values
// joined in a stable (sorted-by-characteristic-name) order since
// map[string]LocalizedString carries no ordering of its own.
func buildDescription(locale string, product *entities.Product, variant *entities.ProductVariant, sku *entities.ProductSKU) string {
	desc := product.Name[locale]
	if v := sortedLocalizedValues(variant.Attributes, locale); len(v) > 0 {
		desc += ", " + strings.Join(v, ", ")
	}
	if v := sortedLocalizedValues(sku.Attributes, locale); len(v) > 0 {
		desc += "; " + strings.Join(v, ", ")
	}
	return desc
}

func sortedLocalizedValues(m map[string]entities.LocalizedString, locale string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		if v := m[k][locale]; v != "" {
			out = append(out, v)
		}
	}
	return out
}

func unitLabel(u entities.PriceUnit) string {
	if u == entities.PriceUnitSquareMeter {
		return "m²"
	}
	return "kom"
}

// invoiceNumber builds the printed document number: prefix + sale number +
// issue date, e.g. "POKU_004-13-07-26". Offer and Receipt use separate
// prefixes (NumberPrefix/ReceiptNumberPrefix) since a Sale can print both
// documents — same Number, so a shared prefix would otherwise make the two
// indistinguishable in file listings/accounting.
func invoiceNumber(docType entities.InvoiceDocumentType, sl *entities.Sale, cfg *appconfig.InvoiceConfig) string {
	prefix := cfg.NumberPrefix
	if docType == entities.InvoiceDocumentTypeReceipt && cfg.ReceiptNumberPrefix != "" {
		prefix = cfg.ReceiptNumberPrefix
	}
	date := sl.CreatedAt.Format("02-01-06")
	if prefix == "" {
		return fmt.Sprintf("%03d-%s", sl.Number, date)
	}
	return fmt.Sprintf("%s_%03d-%s", prefix, sl.Number, date)
}

// sendEmail composes and delivers the buyer-facing notification that
// accompanies the generated PDF. FromName is set to the seller's own
// company name so the message shows up as coming from the business, not
// from whichever personal mailbox SMTPConfig happens to authenticate as.
func (s *Service) sendEmail(ctx context.Context, docType entities.InvoiceDocumentType, number, saleID string, b buyer, company appconfig.InvoiceCompanyConfig, override *string, pdf []byte) {
	to := b.Email
	if override != nil {
		to = *override
	}
	if to == "" {
		s.logger.WarnContext(ctx, "invoice email requested but buyer has no email on file", slog.String("saleId", saleID))
		return
	}

	docLabel, docLabelTitle, closingLine := "ponudu", "Ponuda", "Za sva pitanja slobodno nas kontaktirajte."
	if docType == entities.InvoiceDocumentTypeReceipt {
		docLabel, docLabelTitle, closingLine = "račun", "Račun", "Hvala Vam na poverenju."
	}

	subject := fmt.Sprintf("%s %s — %s", docLabelTitle, number, company.Name)
	body := fmt.Sprintf(
		"Poštovani/a,\n\nU prilogu Vam dostavljamo %s broj %s.\n\n%s\n\nSrdačan pozdrav,\n%s",
		docLabel, number, closingLine, company.Name,
	)

	err := s.mailer.Send(ctx, mailer.Message{
		To:       to,
		Subject:  subject,
		Body:     body,
		FromName: company.Name,
		Attachments: []mailer.Attachment{{
			Filename:    number + ".pdf",
			ContentType: "application/pdf",
			Data:        pdf,
		}},
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "send invoice email failed",
			slog.String("saleId", saleID), slog.String("to", to), slog.String("error", err.Error()))
	}
}

// pageMargin is used on all four sides — usable width on A4 is then
// 210 - 2*pageMargin = 190mm, the figure every hand-tuned width below sums
// to.
const pageMargin = 10.0

// headerFill/headerText are the light-blue section-header color used for
// the "{docLabel} broj/PORUČILAC" box and the item table's header row,
// matching the reference layout in tmp/POKU_004-13-07-26.pdf.
var headerFill = [3]int{198, 217, 241}

// minItemRows is the item table's fixed row count: the reference layout
// always prints 10 rows, padding with blank placeholders when a Sale has
// fewer items, so the box's height doesn't vary from one document to the
// next. A Sale with more than 10 items simply gets that many rows — no
// truncation.
const minItemRows = 10

func (s *Service) render(docType entities.InvoiceDocumentType, number string, sl *entities.Sale, b buyer, lines []line, includeVAT bool, invCfg *appconfig.InvoiceConfig) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8FontFromBytes("DejaVu", "", regularFontBytes)
	pdf.AddUTF8FontFromBytes("DejaVu", "B", boldFontBytes)
	pdf.SetMargins(pageMargin, pageMargin, pageMargin)
	pdf.SetAutoPageBreak(true, 30)
	pdf.SetFooterFunc(func() { renderFooter(pdf, invCfg.Company) })

	// itemsHeader/itemsWidths are set by renderItemsTable for the duration
	// of the item loop, so a page break mid-table repeats the column
	// header on the new page — a Sale with more than minItemRows items
	// would otherwise continue the table with no header at all.
	var itemsHeader []string
	var itemsWidths []float64
	pdf.SetHeaderFunc(func() {
		if itemsHeader != nil {
			drawItemsHeaderRow(pdf, itemsHeader, itemsWidths)
		}
	})
	pdf.AddPage()

	currency := s.cfg.CRM.Currency
	company := invCfg.Company

	// docLabel is the printed Serbian document name (Offer -> "Ponuda",
	// Receipt -> "Račun"), not a code identifier.
	docLabel := "Ponuda"
	if docType == entities.InvoiceDocumentTypeReceipt {
		docLabel = "Račun"
	}

	renderLetterhead(pdf, company)
	renderDocInfoBox(pdf, docLabel, number, sl, b, company, currency)
	totalNet, totalVAT, totalDiscount := renderItemsTable(pdf, lines, includeVAT, currency, invCfg.VAT.Percentage, &itemsHeader, &itemsWidths)
	renderTotals(pdf, totalNet, totalVAT, totalDiscount, includeVAT, invCfg.VAT.Percentage, currency)
	renderNotes(pdf, docType, number, sl, invCfg.Terms, company)
	renderSignature(pdf)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("render invoice pdf: %w", err)
	}
	return buf.Bytes(), nil
}

// renderLetterhead draws the big company wordmark on the left and the full
// legal details (name, address, phone/email, PIB/Matični broj) right-
// aligned on the right, matching a typical Serbian invoice header.
func renderLetterhead(pdf *fpdf.Fpdf, company appconfig.InvoiceCompanyConfig) {
	const rightX, rightW = pageMargin + 100, 90.0

	titleY := pdf.GetY()
	pdf.SetXY(pageMargin, titleY)
	pdf.SetFont("DejaVu", "B", fitFontSize(pdf, "DejaVu", "B", company.Name, 95))
	pdf.CellFormat(95, 14, company.Name, "", 0, "L", false, 0, "")
	titleBottom := titleY + 14

	name := company.Name
	if company.City != "" {
		name = fmt.Sprintf("%s %s", company.Name, company.City)
	}
	rightLines := []string{name, company.Address}
	if company.Phone != "" || company.Email != "" {
		rightLines = append(rightLines, strings.TrimSpace(fmt.Sprintf("Tel: %s  e-mail: %s", company.Phone, company.Email)))
	}
	if company.TaxID != "" || company.RegistrationNumber != "" {
		rightLines = append(rightLines, fmt.Sprintf("PIB: %s   Matični broj: %s", company.TaxID, company.RegistrationNumber))
	}
	y := titleY
	for i, l := range rightLines {
		pdf.SetFont("DejaVu", boldIf(i == 0), 9)
		pdf.SetXY(rightX, y)
		pdf.CellFormat(rightW, 5, l, "", 0, "R", false, 0, "")
		y += 5
	}

	pdf.SetY(maxFloat(titleBottom, y) + 4)
}

func boldIf(b bool) string {
	if b {
		return "B"
	}
	return ""
}

// renderDocInfoBox draws the two side-by-side boxes ("{docLabel} broj /
// dates" on the left, "PORUČILAC" on the right) that make up the
// reference layout's document-info block. The two columns are padded to
// the same total height with a blank filler row so the boxes line up
// regardless of which optional fields (Kupac code, buyer PIB, ...) are
// present.
func renderDocInfoBox(pdf *fpdf.Fpdf, docLabel, number string, sl *entities.Sale, b buyer, company appconfig.InvoiceCompanyConfig, currency string) {
	const leftX, leftW = pageMargin, 115.0
	const rightX, rightW = pageMargin + leftW, 75.0

	startY := pdf.GetY()

	// Left column: document number + dates.
	y := startY
	y = boxRow(pdf, leftX, y, leftW, 7, fmt.Sprintf("%s broj: %s", docLabel, number), "B", "L", true)
	y = boxRow(pdf, leftX, y, leftW, 5, fmt.Sprintf("Datum izdavanja: %s", sl.CreatedAt.Format("2.1.2006")), "", "L", false)
	y = boxRow2(pdf, leftX, y, leftW, 5,
		fmt.Sprintf("Datum prometa: %s", sl.CreatedAt.Format("2.1.2006")),
		fmt.Sprintf("Datum valute: %s", sl.CreatedAt.Format("2.1.2006")))
	y = boxRow2(pdf, leftX, y, leftW, 5,
		fmt.Sprintf("Mesto izdavanja: %s", company.City),
		fmt.Sprintf("Oznaka valute: %s", currency))
	if b.Code != "" {
		y = boxRow(pdf, leftX, y, leftW, 5, fmt.Sprintf("Kupac: %s", b.Code), "", "R", false)
	}
	if company.ResponsiblePerson != "" {
		y = boxRow(pdf, leftX, y, leftW, 5, fmt.Sprintf("Odgovorna osoba: %s", company.ResponsiblePerson), "", "L", false)
	}
	leftBottom := y

	// Right column: buyer.
	y = startY
	y = boxRow(pdf, rightX, y, rightW, 7, "PORUČILAC", "B", "C", true)
	y = boxRow(pdf, rightX, y, rightW, 6, b.Name, "B", "L", false)
	if b.Address != "" {
		y = boxRow(pdf, rightX, y, rightW, 5, b.Address, "", "L", false)
	}
	if b.TaxID != "" || b.RegistrationNumber != "" {
		y = boxRow(pdf, rightX, y, rightW, 5, fmt.Sprintf("PIB: %s   Matični broj: %s", b.TaxID, b.RegistrationNumber), "", "L", false)
	}
	rightBottom := y

	// Pad the shorter column so both boxes end at the same Y.
	bottom := maxFloat(leftBottom, rightBottom)
	if leftBottom < bottom {
		boxRow(pdf, leftX, leftBottom, leftW, bottom-leftBottom, "", "", "L", false)
	}
	if rightBottom < bottom {
		boxRow(pdf, rightX, rightBottom, rightW, bottom-rightBottom, "", "", "L", false)
	}

	pdf.SetY(bottom + 4)
}

// boxRow draws one full-width bordered cell of a document-info column and
// returns the Y coordinate right below it.
func boxRow(pdf *fpdf.Fpdf, x, y, w, h float64, text, style, align string, fill bool) float64 {
	pdf.SetXY(x, y)
	if fill {
		pdf.SetFillColor(headerFill[0], headerFill[1], headerFill[2])
	}
	size := 9.0
	if style == "B" && align == "C" {
		size = 10
	}
	pdf.SetFont("DejaVu", style, size)
	pdf.CellFormat(w, h, text, "1", 0, align, fill, 0, "")
	return y + h
}

// boxRow2 splits a document-info row into two cells (60/40 split) — used
// for the "Datum prometa / Datum valute" and "Mesto izdavanja / Oznaka
// valute" line pairs.
func boxRow2(pdf *fpdf.Fpdf, x, y, w, h float64, left, right string) float64 {
	leftW := w * 0.6
	pdf.SetXY(x, y)
	pdf.SetFont("DejaVu", "", 9)
	pdf.CellFormat(leftW, h, left, "1", 0, "L", false, 0, "")
	pdf.CellFormat(w-leftW, h, right, "1", 0, "L", false, 0, "")
	return y + h
}

// drawItemsHeaderRow draws the item table's column-header row at the
// current cursor position — shared by renderItemsTable's initial draw and
// render()'s SetHeaderFunc callback that repeats it on continuation pages.
func drawItemsHeaderRow(pdf *fpdf.Fpdf, headers []string, widths []float64) {
	pdf.SetFont("DejaVu", "B", 8)
	pdf.SetFillColor(headerFill[0], headerFill[1], headerFill[2])
	for i, h := range headers {
		pdf.CellFormat(widths[i], 7, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)
}

// renderItemsTable draws the item table (10 rows minimum, see
// minItemRows) and returns the running net/VAT/discount totals used by
// renderTotals. headerOut/widthsOut are set for the duration of the item
// loop so render()'s SetHeaderFunc can repeat the column header when the
// table spills onto a new page — cleared before returning so a page break
// elsewhere in the document doesn't redraw a stale item-table header.
func renderItemsTable(pdf *fpdf.Fpdf, lines []line, includeVAT bool, currency string, vatPercent int, headerOut *[]string, widthsOut *[]float64) (totalNet, totalVAT, totalDiscount float64) {
	// Column widths are hand-tuned to sum to exactly the usable page width
	// (190mm) — the header/widths slices are kept in lockstep so a column
	// can never be added to one without the other.
	var headers []string
	var widths []float64
	if includeVAT {
		headers = []string{"Rbr.", "Šifra", "Naziv", "JM", "Količina", "Cena (RSD)", "Rabat (%)", "PDV (%)", "Iznos PDV", "Iznost (RSD)"}
		widths = []float64{8, 16, 50, 10, 16, 22, 14, 12, 20, 22}
	} else {
		headers = []string{"Rbr.", "Šifra", "Naziv", "JM", "Količina", "Cena (RSD)", "Rabat (%)", "Iznost (RSD)"}
		widths = []float64{8, 16, 60, 12, 18, 24, 16, 36}
	}
	*headerOut = headers
	*widthsOut = widths
	defer func() { *headerOut = nil }()

	drawItemsHeaderRow(pdf, headers, widths)

	pdf.SetFont("DejaVu", "", 8)
	rowCount := len(lines)
	if rowCount < minItemRows {
		rowCount = minItemRows
	}
	descWidth := widths[2]
	for i := 0; i < rowCount; i++ {
		row := []string{fmt.Sprintf("%d", i+1)}
		if i < len(lines) {
			l := lines[i]
			grossBeforeDiscount := l.UnitPrice * l.Quantity
			totalDiscount += grossBeforeDiscount - l.NetAmount
			totalNet += l.NetAmount
			totalVAT += l.VATAmount

			row = append(row,
				l.Code,
				fitText(pdf, l.Description, descWidth-2),
				l.Unit,
				formatAmount(l.Quantity),
				fmt.Sprintf("%s %s", formatAmount(l.UnitPrice), currency),
				fmt.Sprintf("%d%%", l.DiscountPercentage),
			)
			if includeVAT {
				row = append(row,
					fmt.Sprintf("%d%%", vatPercent),
					fmt.Sprintf("%s %s", formatAmount(l.VATAmount), currency),
					fmt.Sprintf("%s %s", formatAmount(l.GrossAmount), currency),
				)
			} else {
				row = append(row, fmt.Sprintf("%s %s", formatAmount(l.NetAmount), currency))
			}
		} else {
			// Blank placeholder row — keeps the table a fixed 10 rows tall.
			row = append(row, "", "", "m²", "0", "- "+currency, "0%")
			if includeVAT {
				row = append(row, fmt.Sprintf("%d%%", vatPercent), "- "+currency, "- "+currency)
			} else {
				row = append(row, "- "+currency)
			}
		}
		for i, cell := range row {
			pdf.CellFormat(widths[i], 6, cell, "1", 0, "C", false, 0, "")
		}
		pdf.Ln(-1)
	}
	pdf.Ln(4)
	return totalNet, totalVAT, totalDiscount
}

// renderTotals draws the right-aligned Osnovica/PDV/Rabat/Ukupno block
// under the item table, label and amount in their own fixed-width column
// so every row's amount lines up on the same right edge.
func renderTotals(pdf *fpdf.Fpdf, totalNet, totalVAT, totalDiscount float64, includeVAT bool, vatPercent int, currency string) {
	const labelW, valueW = 55.0, 35.0
	x := pageMargin + 190 - labelW - valueW

	row := func(label, value string, bold bool) {
		pdf.SetXY(x, pdf.GetY())
		pdf.SetFont("DejaVu", boldIf(bold), floatOr(bold, 10, 9))
		pdf.CellFormat(labelW, 6, label, "", 0, "R", false, 0, "")
		pdf.CellFormat(valueW, 6, value, "", 1, "R", false, 0, "")
	}

	row("Osnovica (cena bez PDV-a):", fmt.Sprintf("%s %s", formatAmount(totalNet), currency), false)
	if includeVAT {
		row(fmt.Sprintf("PDV %d%%:", vatPercent), fmt.Sprintf("%s %s", formatAmount(totalVAT), currency), false)
	}
	row("Rabat:", fmt.Sprintf("%s %s", formatAmount(totalDiscount), currency), false)
	label := "Ukupno:"
	if includeVAT {
		label = "Ukupno sa PDV:"
	}
	row(label, fmt.Sprintf("%s %s", formatAmount(totalNet+totalVAT), currency), true)
	pdf.Ln(4)
}

func floatOr(cond bool, ifTrue, ifFalse float64) float64 {
	if cond {
		return ifTrue
	}
	return ifFalse
}

// renderNotes draws the VAT-status note, payment-reference note (Offer
// only), and the commercial terms / payment-confirmation boilerplate.
func renderNotes(pdf *fpdf.Fpdf, docType entities.InvoiceDocumentType, number string, sl *entities.Sale, terms appconfig.InvoiceTermsConfig, company appconfig.InvoiceCompanyConfig) {
	pdf.SetFont("DejaVu", "", 9)
	pdf.CellFormat(0, 5, fmt.Sprintf("Napomena o poreskom oslobođenju: %s je u sistemu PDV-a.", company.Name), "", 1, "L", false, 0, "")
	pdf.Ln(2)

	if docType != entities.InvoiceDocumentTypeReceipt {
		pdf.CellFormat(0, 5, fmt.Sprintf("Molimo Vas da se pri uplati pozovete na broj: %s uključujući slovni prefiks.", number), "", 1, "L", false, 0, "")
		pdf.Ln(2)
	}

	// Terms — an Offer still needs the commercial terms/offer-validity
	// boilerplate; a Receipt documents payment already received instead,
	// so none of that applies.
	if docType == entities.InvoiceDocumentTypeReceipt {
		pdf.CellFormat(0, 5, "NAČIN PLAĆANJA: Uplata izvršena u celosti.", "", 1, "L", false, 0, "")
		pdf.CellFormat(0, 5, fmt.Sprintf("Datum uplate: %s", sl.UpdatedAt.Format("2.1.2006")), "", 1, "L", false, 0, "")
		return
	}
	pdf.CellFormat(0, 5, "USLOVI PLAĆANJA:", "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 5, fmt.Sprintf("AVANS %d%%", terms.AdvancePercentage), "", 1, "L", false, 0, "")
	pdf.Ln(2)
	if terms.DeliveryLeadTime != "" {
		pdf.CellFormat(0, 5, fmt.Sprintf("ROK ISPORUKE: %s", terms.DeliveryLeadTime), "", 1, "L", false, 0, "")
	}
	pdf.CellFormat(0, 5, fmt.Sprintf("Ponuda važi %d dana.", terms.OfferValidityDays), "", 1, "L", false, 0, "")
}

// renderSignature draws the blank signature line + "Potpis" caption near
// the end of the document body.
func renderSignature(pdf *fpdf.Fpdf) {
	pdf.Ln(14)
	y := pdf.GetY()
	pdf.Line(pageMargin, y, pageMargin+55, y)
	pdf.SetFont("DejaVu", "", 9)
	pdf.SetXY(pageMargin, y+1)
	pdf.CellFormat(55, 5, "Potpis", "", 1, "C", false, 0, "")
}

// renderFooter draws the bank/PIB/activity-code strip repeated at the
// bottom of every page, plus the page number — installed via
// pdf.SetFooterFunc so it appears whether the document is one page or
// several.
func renderFooter(pdf *fpdf.Fpdf, company appconfig.InvoiceCompanyConfig) {
	y := -20.0
	pdf.SetY(y)
	pdf.SetFont("DejaVu", "", 8)

	colW := 190.0 / 3
	x := pageMargin

	pdf.SetXY(x, pdf.GetY())
	pdf.CellFormat(colW, 4, "Broj dinarskog računa:", "", 2, "L", false, 0, "")
	pdf.SetX(x)
	pdf.CellFormat(colW, 4, fmt.Sprintf("%s: %s", company.BankName, company.BankAccount), "", 0, "L", false, 0, "")

	x += colW
	pdf.SetXY(x, y)
	pdf.CellFormat(colW, 4, fmt.Sprintf("PIB: %s", company.TaxID), "", 2, "C", false, 0, "")
	pdf.SetX(x)
	pdf.CellFormat(colW, 4, fmt.Sprintf("Matični broj: %s", company.RegistrationNumber), "", 0, "C", false, 0, "")

	x += colW
	pdf.SetXY(x, y)
	pdf.CellFormat(colW, 4, fmt.Sprintf("Šifra delatnosti: %s", company.ActivityCode), "", 2, "R", false, 0, "")
	pdf.SetX(x)
	pdf.CellFormat(colW, 4, fmt.Sprintf("%d", pdf.PageNo()), "", 0, "R", false, 0, "")
}

// fitFontSize shrinks size from 28 down to 14 until text fits within
// maxWidth at that font — the letterhead wordmark otherwise overruns the
// company-details column for a long legal name.
func fitFontSize(pdf *fpdf.Fpdf, family, style, text string, maxWidth float64) float64 {
	for size := 28.0; size > 14; size -= 2 {
		pdf.SetFont(family, style, size)
		if pdf.GetStringWidth(text) <= maxWidth {
			return size
		}
	}
	return 14
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// formatAmount renders v in Serbian number format ("." as the thousands
// separator, "," as the decimal point — e.g. 12411.24 -> "12.411,24"),
// matching the reference layout in tmp/POKU_004-13-07-26.pdf.
func formatAmount(v float64) string {
	whole := fmt.Sprintf("%.2f", v)
	intPart, decPart, _ := strings.Cut(whole, ".")
	neg := strings.HasPrefix(intPart, "-")
	intPart = strings.TrimPrefix(intPart, "-")

	var grouped strings.Builder
	for i, r := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			grouped.WriteByte('.')
		}
		grouped.WriteRune(r)
	}

	sign := ""
	if neg {
		sign = "-"
	}
	return fmt.Sprintf("%s%s,%s", sign, grouped.String(), decPart)
}

// fitText truncates text with a trailing "…" until it fits maxWidth at
// pdf's current font — CellFormat draws a single fixed-height line and
// never wraps or clips on its own, so an overlong description would
// otherwise visibly spill into the neighboring column.
func fitText(pdf *fpdf.Fpdf, text string, maxWidth float64) string {
	if pdf.GetStringWidth(text) <= maxWidth {
		return text
	}
	runes := []rune(text)
	for i := len(runes) - 1; i > 0; i-- {
		candidate := string(runes[:i]) + "…"
		if pdf.GetStringWidth(candidate) <= maxWidth {
			return candidate
		}
	}
	return "…"
}
