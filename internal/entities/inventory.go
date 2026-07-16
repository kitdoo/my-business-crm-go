package entities

import "time"

// Inventory is the current stock level for a (SKUID, WarehouseID) pair.
// It has no CreatedAt/DeletedAt — it is a running counter, not a
// soft-deletable resource — and is never written directly by a service
// consumer; only InventoryMovement mutates it (see
// storages/inventory.Storage.ApplyMovement).
type Inventory struct {
	ID          string
	SKUID       string
	WarehouseID string
	Quantity    int64
	UpdatedAt   time.Time
	Etag        string // OCC token; rolled on every ApplyMovement
}

// InventoryListSortField is int32-aligned with
// InventoryListRequest.Sort.Field. Quantity is the only sortable field —
// Inventory carries no product name to sort by.
type InventoryListSortField int32

const (
	InventoryListSortFieldQuantity InventoryListSortField = iota
)

type InventoryListSort struct {
	Field     InventoryListSortField
	Direction SortDirection
}

// InventoryList is the single List input; scope/filters/pagination all live
// inside it, per the List(ctx, in *XxxList) convention.
type InventoryList struct {
	SKUID *string
	// SKUIDs batches a lookup across multiple SKUs in one query — set this
	// instead of SKUID for a batch lookup, not alongside it.
	SKUIDs      []string
	WarehouseID *string
	MinQuantity *int64
	MaxQuantity *int64
	Sort        *InventoryListSort
	Pagination  ListPagination
}
