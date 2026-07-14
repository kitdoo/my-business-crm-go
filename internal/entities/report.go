package entities

import "time"

// Report rows are computed on the fly from Sales/Inventory (aggregation
// pipelines) — there is no backing collection, per the TD.

// SalesReportRow is one day-bucket within the requested period.
type SalesReportRow struct {
	PeriodStart time.Time
	PeriodEnd   time.Time
	SalesCount  int64
	TotalAmount int64
}

// SalesByStaffRow aggregates sales created by one user over the period.
type SalesByStaffRow struct {
	UserID      string
	SalesCount  int64
	TotalAmount int64
}

// SalesByPartnerRow aggregates sales attributed to one partner over the
// period; CommissionAmount = TotalAmount * Partner.CommissionPercentage / 100.
type SalesByPartnerRow struct {
	PartnerID        string
	SalesCount       int64
	TotalAmount      int64
	CommissionAmount int64
}

// PopularProductRow ranks a variant by quantity sold over the period.
type PopularProductRow struct {
	VariantID    string
	QuantitySold int64
	TotalAmount  int64
}

// TurnoverRow is one day-bucket's total revenue within the requested period.
type TurnoverRow struct {
	PeriodStart time.Time
	PeriodEnd   time.Time
	TotalAmount int64
}

// StockLevelRow is a live Inventory row, optionally scoped to one warehouse.
type StockLevelRow struct {
	VariantID   string
	WarehouseID string
	Quantity    int64
}
