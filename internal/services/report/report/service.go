// Package report implements the report.Service interface.
package report

import (
	"context"
	"log/slog"

	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	inventorysvc "github.com/kitdoo/my-business-crm-go/internal/services/inventory"
	reportsvc "github.com/kitdoo/my-business-crm-go/internal/services/report"
	"github.com/kitdoo/my-business-crm-go/internal/storages/reports"
)

var _ reportsvc.Service = (*Service)(nil)

// Service is the report.Service implementation. inventory is
// inventory.Service, not inventory.Storage — see
// SERVICE_DEVELOPMENT_STANDARD.md's "A service controls only its own
// storage" rule.
type Service struct {
	storage   reports.Storage
	inventory inventorysvc.Service
	logger    *slog.Logger
}

// New builds a Service.
func New(storage reports.Storage, inv inventorysvc.Service) *Service {
	return &Service{
		storage:   storage,
		inventory: inv,
		logger:    slog.Default().With(slogx.Module("service:report")),
	}
}

func (s *Service) GetSalesReport(ctx context.Context, period *entities.PeriodFilter) ([]entities.SalesReportRow, error) {
	rows, err := s.storage.GetSalesReport(ctx, period)
	if err != nil {
		s.logger.DebugContext(ctx, "get sales report failed", slog.String("error", err.Error()))
		return nil, err
	}
	return rows, nil
}

func (s *Service) GetSalesByStaff(ctx context.Context, period *entities.PeriodFilter) ([]entities.SalesByStaffRow, error) {
	rows, err := s.storage.GetSalesByStaff(ctx, period)
	if err != nil {
		s.logger.DebugContext(ctx, "get sales by staff report failed", slog.String("error", err.Error()))
		return nil, err
	}
	return rows, nil
}

func (s *Service) GetSalesByPartner(ctx context.Context, period *entities.PeriodFilter) ([]entities.SalesByPartnerRow, error) {
	rows, err := s.storage.GetSalesByPartner(ctx, period)
	if err != nil {
		s.logger.DebugContext(ctx, "get sales by partner report failed", slog.String("error", err.Error()))
		return nil, err
	}
	return rows, nil
}

func (s *Service) GetPopularProducts(ctx context.Context, period *entities.PeriodFilter, limit int32) ([]entities.PopularProductRow, error) {
	rows, err := s.storage.GetPopularProducts(ctx, period, limit)
	if err != nil {
		s.logger.DebugContext(ctx, "get popular products report failed", slog.String("error", err.Error()))
		return nil, err
	}
	return rows, nil
}

func (s *Service) GetTurnover(ctx context.Context, period *entities.PeriodFilter) ([]entities.TurnoverRow, error) {
	rows, err := s.storage.GetTurnover(ctx, period)
	if err != nil {
		s.logger.DebugContext(ctx, "get turnover report failed", slog.String("error", err.Error()))
		return nil, err
	}
	return rows, nil
}

// GetStockLevels is not an aggregation of its own — it is a live read of
// Inventory (optionally scoped to one warehouse), paged internally since
// GetStockLevelsResponse carries no cursor.
func (s *Service) GetStockLevels(ctx context.Context, warehouseID *string) ([]entities.StockLevelRow, error) {
	var rows []entities.StockLevelRow
	cursor := ""
	for {
		result, err := s.inventory.List(ctx, &entities.InventoryList{
			WarehouseID: warehouseID,
			Pagination:  entities.ListPagination{Cursor: cursor},
		})
		if err != nil {
			s.logger.DebugContext(ctx, "get stock levels failed", slog.String("error", err.Error()))
			return nil, err
		}
		for _, item := range result.Items {
			rows = append(rows, entities.StockLevelRow{
				SKUID:       item.SKUID,
				WarehouseID: item.WarehouseID,
				Quantity:    item.Quantity,
			})
		}
		if result.NextCursor == "" {
			return rows, nil
		}
		cursor = result.NextCursor
	}
}
