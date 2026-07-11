// Package report implements the report.Service interface.
package report

import (
	"context"

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
}

// New builds a Service.
func New(storage reports.Storage, inv inventorysvc.Service) *Service {
	return &Service{storage: storage, inventory: inv}
}

func (s *Service) GetSalesReport(ctx context.Context, period *entities.PeriodFilter) ([]entities.SalesReportRow, error) {
	return s.storage.GetSalesReport(ctx, period)
}

func (s *Service) GetSalesByStaff(ctx context.Context, period *entities.PeriodFilter) ([]entities.SalesByStaffRow, error) {
	return s.storage.GetSalesByStaff(ctx, period)
}

func (s *Service) GetSalesByPartner(ctx context.Context, period *entities.PeriodFilter) ([]entities.SalesByPartnerRow, error) {
	return s.storage.GetSalesByPartner(ctx, period)
}

func (s *Service) GetPopularProducts(ctx context.Context, period *entities.PeriodFilter, limit int32) ([]entities.PopularProductRow, error) {
	return s.storage.GetPopularProducts(ctx, period, limit)
}

func (s *Service) GetTurnover(ctx context.Context, period *entities.PeriodFilter) ([]entities.TurnoverRow, error) {
	return s.storage.GetTurnover(ctx, period)
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
			return nil, err
		}
		for _, item := range result.Items {
			rows = append(rows, entities.StockLevelRow{
				ProductID:   item.ProductID,
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
