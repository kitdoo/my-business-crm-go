package entities

import "time"

// Inventory is the current stock level for a (ProductID, WarehouseID) pair.
// It has no CreatedAt/DeletedAt — it is a running counter, not a
// soft-deletable resource — and is never written directly by a service
// consumer; only InventoryMovement mutates it (see
// storages/inventory.Storage.ApplyMovement).
type Inventory struct {
	ID          string
	ProductID   string
	WarehouseID string
	Quantity    int64
	UpdatedAt   time.Time
	Etag        string // OCC token; rolled on every ApplyMovement
}

// InventoryList is the single List input; scope/filters/pagination all live
// inside it, per the List(ctx, in *XxxList) convention.
type InventoryList struct {
	ProductID   *string
	WarehouseID *string
	MinQuantity *int64
	MaxQuantity *int64
	Pagination  ListPagination
}
