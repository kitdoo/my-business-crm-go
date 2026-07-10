// Package reports defines the read-only Reports storage interface; the
// MongoDB implementation lives in the mongo subpackage. Every method is an
// aggregation pipeline over Sales (and, for GetSalesByPartner, Partners) —
// there is no backing collection of its own, per the TD.
package reports

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Storage computes report rows on the fly.
type Storage interface {
	GetSalesReport(ctx context.Context, period *entities.PeriodFilter) ([]entities.SalesReportRow, error)
	GetSalesByStaff(ctx context.Context, period *entities.PeriodFilter) ([]entities.SalesByStaffRow, error)
	GetSalesByPartner(ctx context.Context, period *entities.PeriodFilter) ([]entities.SalesByPartnerRow, error)
	GetPopularProducts(ctx context.Context, period *entities.PeriodFilter, limit int32) ([]entities.PopularProductRow, error)
	GetTurnover(ctx context.Context, period *entities.PeriodFilter) ([]entities.TurnoverRow, error)
}
