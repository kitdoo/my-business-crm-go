// Package salesreport implements the salesreport.Service interface:
// renders every Sale created within a period as an .xlsx workbook (one row
// per sale, plus a totals row).
package salesreport

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"github.com/xuri/excelize/v2"

	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/pkg/appconfig"
	clientsvc "github.com/kitdoo/my-business-crm-go/internal/services/client"
	partnersvc "github.com/kitdoo/my-business-crm-go/internal/services/partner"
	salesvc "github.com/kitdoo/my-business-crm-go/internal/services/sale"
	salesreportsvc "github.com/kitdoo/my-business-crm-go/internal/services/salesreport"
	usersvc "github.com/kitdoo/my-business-crm-go/internal/services/user"
)

var _ salesreportsvc.Service = (*Service)(nil)

// pageSize is how many Sales are fetched per List call while paging
// through the whole period — large enough that a typical month/quarter
// report needs only a couple of round trips.
const pageSize = 200

// Service is the salesreport.Service implementation. sales/clients/
// partners/users are the respective entities' Service, not their Storage —
// see SERVICE_DEVELOPMENT_STANDARD.md's "A service controls only its own
// storage" rule. There is no storage of its own: Generate is stateless.
type Service struct {
	sales    salesvc.Service
	clients  clientsvc.Service
	partners partnersvc.Service
	users    usersvc.Service
	cfg      *appconfig.Config
	logger   *slog.Logger
}

// New builds a Service.
func New(sales salesvc.Service, clients clientsvc.Service, partners partnersvc.Service, users usersvc.Service, cfg *appconfig.Config) *Service {
	return &Service{
		sales:    sales,
		clients:  clients,
		partners: partners,
		users:    users,
		cfg:      cfg,
		logger:   slog.Default().With(slogx.Module("service:salesreport")),
	}
}

// buyerCache/sellerCache avoid re-fetching the same Client/Partner/User
// once per row when a period's Sales repeat the same buyer or seller, as
// they typically do.
type nameCache struct {
	clients  map[string]string
	partners map[string]string
	users    map[string]string
}

func (s *Service) Generate(ctx context.Context, in *entities.SalesReportGenerate) ([]byte, error) {
	sales, err := s.collectSales(ctx, in)
	if err != nil {
		return nil, err
	}

	locale := resolveLocale(s.cfg)
	currency := s.cfg.CRM.Currency
	cache := &nameCache{clients: map[string]string{}, partners: map[string]string{}, users: map[string]string{}}

	f := excelize.NewFile()
	defer f.Close() //nolint:errcheck
	const sheet = "Sales"
	f.SetSheetName(f.GetSheetName(0), sheet)

	headerStyle, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}, Fill: excelize.Fill{Type: "pattern", Color: []string{"#DAE6F1"}, Pattern: 1}})
	if err != nil {
		return nil, fmt.Errorf("build header style: %w", err)
	}
	totalsStyle, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	if err != nil {
		return nil, fmt.Errorf("build totals style: %w", err)
	}

	headers := []string{"Broj", "Datum", "Kupac", "Prodavac", "Status", fmt.Sprintf("Iznos (%s)", currency)}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	headerRange := fmt.Sprintf("A1:%s1", string(rune('A'+len(headers)-1)))
	f.SetCellStyle(sheet, "A1", headerRange, headerStyle)

	row := 2
	var totalAmount int64
	for _, sl := range sales {
		buyer := s.resolveBuyerName(ctx, sl, cache)
		seller := s.resolveSellerName(ctx, sl.CreatedBy, locale, cache)

		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), sl.Number)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), sl.CreatedAt.Format("2.1.2006"))
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), buyer)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), seller)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), statusLabel(sl.Status))
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), float64(sl.TotalAmount)/100)

		totalAmount += sl.TotalAmount
		row++
	}

	f.SetCellValue(sheet, fmt.Sprintf("D%d", row), "Ukupno:")
	f.SetCellValue(sheet, fmt.Sprintf("E%d", row), fmt.Sprintf("%d prodaja", len(sales)))
	f.SetCellValue(sheet, fmt.Sprintf("F%d", row), float64(totalAmount)/100)
	f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("F%d", row), totalsStyle)

	for col, width := range map[string]float64{"A": 10, "B": 12, "C": 30, "D": 22, "E": 14, "F": 16} {
		_ = f.SetColWidth(sheet, col, col, width)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("write sales report xlsx: %w", err)
	}
	return buf.Bytes(), nil
}

// collectSales pages through SalesService.List for the whole period —
// there is no "list everything" call, so this loops on NextCursor until
// exhausted.
func (s *Service) collectSales(ctx context.Context, in *entities.SalesReportGenerate) ([]*entities.Sale, error) {
	from, to := in.From, in.To
	var out []*entities.Sale
	cursor := ""
	for {
		res, err := s.sales.List(ctx, &entities.SalesList{
			CreatedAt:  &entities.PeriodFilter{From: &from, To: &to},
			Sort:       entities.SalesListSort{Field: entities.SalesListSortFieldCreatedAt, Direction: entities.SortDirectionAsc},
			Pagination: entities.ListPagination{Limit: pageSize, Cursor: cursor},
		})
		if err != nil {
			return nil, err
		}
		out = append(out, res.Items...)
		if res.NextCursor == "" {
			break
		}
		cursor = res.NextCursor
	}
	return out, nil
}

// resolveBuyerName resolves a Sale's Client (or, for a partner-only sale,
// its Partner) to a display name, tolerating a since-deleted buyer (logs a
// warning and falls back to the raw ID) rather than failing the whole
// report over one bad row.
func (s *Service) resolveBuyerName(ctx context.Context, sl *entities.Sale, cache *nameCache) string {
	if sl.ClientID != "" {
		if name, ok := cache.clients[sl.ClientID]; ok {
			return name
		}
		c, err := s.clients.Get(ctx, sl.ClientID)
		if err != nil {
			s.logger.WarnContext(ctx, "resolve client for sales report failed", slog.String("clientId", sl.ClientID), slog.String("error", err.Error()))
			cache.clients[sl.ClientID] = sl.ClientID
			return sl.ClientID
		}
		cache.clients[sl.ClientID] = c.Name
		return c.Name
	}
	if sl.PartnerID != nil {
		id := *sl.PartnerID
		if name, ok := cache.partners[id]; ok {
			return name
		}
		p, err := s.partners.Get(ctx, id)
		if err != nil {
			s.logger.WarnContext(ctx, "resolve partner for sales report failed", slog.String("partnerId", id), slog.String("error", err.Error()))
			cache.partners[id] = id
			return id
		}
		cache.partners[id] = p.Name
		return p.Name
	}
	return "-"
}

func (s *Service) resolveSellerName(ctx context.Context, userID, locale string, cache *nameCache) string {
	if name, ok := cache.users[userID]; ok {
		return name
	}
	u, err := s.users.Get(ctx, userID)
	if err != nil {
		s.logger.WarnContext(ctx, "resolve seller for sales report failed", slog.String("userId", userID), slog.String("error", err.Error()))
		cache.users[userID] = userID
		return userID
	}
	name := fmt.Sprintf("%s %s", u.Name[locale], u.LastName[locale])
	cache.users[userID] = name
	return name
}

// statusLabel prints the Serbian label matching the CRM UI's own
// StatusBadge/i18n keys (web/i18n/locales/sr.json fields.sale.status.*).
func statusLabel(status entities.SaleStatus) string {
	switch status {
	case entities.SaleStatusDraft:
		return "Nacrt"
	case entities.SaleStatusPaid:
		return "Plaćeno"
	case entities.SaleStatusShipped:
		return "Isporučeno"
	case entities.SaleStatusCompleted:
		return "Završeno"
	case entities.SaleStatusCancelled:
		return "Otkazano"
	case entities.SaleStatusRefunded:
		return "Refundirano"
	default:
		return "-"
	}
}

// resolveLocale mirrors invoice.Service's own resolveDefaultLocale — kept
// local since importing across sibling service packages for one constant
// isn't worth the coupling.
func resolveLocale(cfg *appconfig.Config) string {
	if cfg.CRM != nil && cfg.CRM.DefaultLocale != "" {
		return cfg.CRM.DefaultLocale
	}
	return "sr"
}
