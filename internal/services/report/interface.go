// Package report defines the Reports service interface; the implementation
// lives in the report subpackage. Every method is read-only and computed
// on the fly — there is no backing collection, per the TD.
package report

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Service reads report rows on top of storages.Storage (Sales/Partners
// aggregation) and Inventory (for GetStockLevels).
type Service interface {
	GetSalesReport(ctx context.Context, period *entities.PeriodFilter) ([]entities.SalesReportRow, error)
	GetSalesByStaff(ctx context.Context, period *entities.PeriodFilter) ([]entities.SalesByStaffRow, error)
	GetSalesByPartner(ctx context.Context, period *entities.PeriodFilter) ([]entities.SalesByPartnerRow, error)
	GetPopularProducts(ctx context.Context, period *entities.PeriodFilter, limit int32) ([]entities.PopularProductRow, error)
	GetTurnover(ctx context.Context, period *entities.PeriodFilter) ([]entities.TurnoverRow, error)
	GetStockLevels(ctx context.Context, warehouseID *string) ([]entities.StockLevelRow, error)
}
