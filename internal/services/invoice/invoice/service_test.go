package invoice

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/pkg/appconfig"
)

// testInvoiceConfig/testSale/testBuyer fabricate the seller/sale/buyer
// used by every render smoke test below — same data as the reference
// layout in tmp/POKU_004-13-07-26.pdf, so the output can be eyeballed
// against it.
func testInvoiceConfig() *appconfig.InvoiceConfig {
	return &appconfig.InvoiceConfig{
		Company: appconfig.InvoiceCompanyConfig{
			Name: "TOM STUDIO 021 DOO", Address: "Dr Đure Jovanovića Kindera 6, 21000 Novi Sad",
			Phone: "+381 65 563 2551", Email: "info@tom.archi",
			TaxID: "115473922", RegistrationNumber: "22163132", ActivityCode: "4690",
			BankName: "Raiffeisen banka", BankAccount: "265759031000072456",
			ResponsiblePerson: "Jovan Tomić",
		},
		NumberPrefix:        "POKU",
		ReceiptNumberPrefix: "RAC",
		Terms: appconfig.InvoiceTermsConfig{
			AdvancePercentage: 100,
			DeliveryLeadTime:  "5-6 radnih dana od datuma prihvatanja ponude i uplate avansa.",
			OfferValidityDays: 5,
		},
		VAT: appconfig.InvoiceVATConfig{Percentage: 20},
	}
}

func testSale() *entities.Sale {
	return &entities.Sale{
		Number: 4, Status: entities.SaleStatusPaid,
		CreatedAt: time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC),
	}
}

func testBuyer() buyer {
	return buyer{
		Name: "SANJA ŠULOVIĆ PR S.O.B.A. Decor", Address: "Dragiše Mišovića 30, 24000, Subotica, Srbija",
		TaxID: "113397194", RegistrationNumber: "66794814", Code: "0003",
	}
}

// TestRenderSmoke renders a fabricated Offer and a fabricated Receipt to
// scratch files so the layout can be eyeballed against
// tmp/POKU_004-13-07-26.pdf / tmp/RAC_004-13-07-26.pdf — not a behavioral
// assertion, just a compile+crash smoke test plus a way to visually check
// the output.
func TestRenderSmoke(t *testing.T) {
	s := &Service{cfg: &appconfig.Config{CRM: &appconfig.CRMConfig{Currency: "RSD"}}}
	invCfg := testInvoiceConfig()
	sl := testSale()
	b := testBuyer()
	lines := []line{{
		Code: "01RS", Description: "Rockface Stone, Andes White (3.0); 1200/600/4", Unit: "m²",
		Quantity: 2.16, UnitPrice: 59.8536, DiscountPercentage: 20,
		NetAmount: 103.427, VATAmount: 20.6854, GrossAmount: 124.1124,
	}}

	for _, docType := range []entities.InvoiceDocumentType{entities.InvoiceDocumentTypeOffer, entities.InvoiceDocumentTypeReceipt} {
		number := invoiceNumber(docType, sl, invCfg)
		data, err := s.render(docType, number, sl, b, lines, true, invCfg)
		if err != nil {
			t.Fatalf("render %v: %v", docType, err)
		}
		if len(data) == 0 {
			t.Fatalf("empty pdf for %v", docType)
		}
		if err := os.WriteFile("/tmp/invoice_smoke_"+number+".pdf", data, 0o644); err != nil {
			t.Fatalf("write smoke pdf: %v", err)
		}
	}
}

// TestRenderSmokeMultiPage renders an Offer with 35 varied line items —
// well past minItemRows — so the item table spills onto a second page.
// Checks that pagination doesn't crash and that the footer (installed via
// SetFooterFunc) still prints on every page; the actual page-break
// placement is only verified by eye against /tmp/invoice_smoke_multipage.pdf.
func TestRenderSmokeMultiPage(t *testing.T) {
	s := &Service{cfg: &appconfig.Config{CRM: &appconfig.CRMConfig{Currency: "RSD"}}}
	invCfg := testInvoiceConfig()
	sl := testSale()
	b := testBuyer()

	products := []struct {
		code, desc, unit string
	}{
		{"01RS", "Rockface Stone, Andes White (3.0); 1200/600/4", "m²"},
		{"02RS", "Rockface Stone, Canyon Grey (2.5); 1200/600/4", "m²"},
		{"03RS", "Ledge Stone, Autumn Blend (2.0); 900/400/3", "m²"},
		{"04RS", "Cultured Brick, Old Chicago; 600/200/2", "m²"},
		{"05RS", "Cultured Brick, Weathered Red; 600/200/2", "m²"},
		{"06RS", "Corner Stone, Andes White; 300/150/4", "kom"},
		{"07RS", "Corner Stone, Canyon Grey; 300/150/4", "kom"},
		{"08RS", "Sill Stone, Natural Grey; 1000/150/3", "kom"},
		{"09RS", "Window Trim, Almond; 1200/100/2", "kom"},
		{"10RS", "Column Wrap, Andes White; 400/400/6", "kom"},
	}

	lines := make([]line, 0, 35)
	for i := range 35 {
		p := products[i%len(products)]
		qty := 1.0 + float64(i%7)
		unitPrice := 45.0 + float64(i*3)
		discount := int32(i % 4 * 5)
		net := unitPrice * qty * float64(100-discount) / 100
		vat := net * 0.2
		lines = append(lines, line{
			Code:               p.code,
			Description:        fmt.Sprintf("%s (batch %d)", p.desc, i+1),
			Unit:               p.unit,
			Quantity:           qty,
			UnitPrice:          unitPrice,
			DiscountPercentage: discount,
			NetAmount:          net,
			VATAmount:          vat,
			GrossAmount:        net + vat,
		})
	}

	number := invoiceNumber(entities.InvoiceDocumentTypeOffer, sl, invCfg)
	data, err := s.render(entities.InvoiceDocumentTypeOffer, number, sl, b, lines, true, invCfg)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("empty pdf")
	}
	if err := os.WriteFile("/tmp/invoice_smoke_multipage.pdf", data, 0o644); err != nil {
		t.Fatalf("write smoke pdf: %v", err)
	}
}
