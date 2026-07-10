package entities

import (
	"time"

	"github.com/google/uuid"

	"github.com/altessa-s/go-atlas/domain/converter"
)

// WarehouseStatus is int32-aligned with crm.types.warehouse.WarehouseStatus
// so converter.Convert maps it both ways as a plain scalar.
type WarehouseStatus int32

const (
	WarehouseStatusUnspecified WarehouseStatus = iota
	WarehouseStatusActive
	WarehouseStatusInactive
)

// Warehouse is a physical stock location.
type Warehouse struct {
	ID          string
	Name        LocalizedString
	Description LocalizedString
	Address     string // factual data, not localized
	Status      WarehouseStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time // nil = active
	Etag        string     // OCC token; rolled on every write
}

// WarehouseNew creates a Warehouse with a fresh ID, timestamps, and etag.
func WarehouseNew(init ...func(*Warehouse)) *Warehouse {
	w := &Warehouse{ID: uuid.NewString()}
	if len(init) > 0 {
		init[0](w)
	}
	w.UpdateTimestamps()
	w.UpdateEtag()
	return w
}

func (w *Warehouse) UpdateTimestamps() {
	now := time.Now().UTC()
	w.UpdatedAt = now
	if w.CreatedAt.IsZero() {
		w.CreatedAt = now
	}
}

func (w *Warehouse) UpdateEtag() { w.Etag = uuid.NewString() }

func (w *Warehouse) BeforeUpdate() {
	w.UpdateTimestamps()
	w.UpdateEtag()
}

// WarehouseCreate is the Create input. Merge applies it onto a freshly
// constructed Warehouse via the converter.
type WarehouseCreate struct {
	Name        LocalizedString
	Description LocalizedString
	Address     string `normalize:"trim"`
}

func (c *WarehouseCreate) Merge(dst *Warehouse) *Warehouse {
	if c == nil || dst == nil {
		return dst
	}
	converter.Convert(c, dst, converter.WithIgnoreNilValues())
	return dst
}

// WarehouseUpdate is the Update input. Nil fields mean "leave unchanged".
type WarehouseUpdate struct {
	ID          string `normalize:"trim"`
	Name        LocalizedString
	Description LocalizedString
	Address     *string `normalize:"trim,nil_on_empty"`
	Etag        *string `normalize:"trim,nil_on_empty"` // client OCC precondition
}

func (u *WarehouseUpdate) Merge(dst *Warehouse) *Warehouse {
	if u == nil || dst == nil {
		return dst
	}
	converter.Convert(u, dst, converter.WithIgnoreNilValues())
	return dst
}

// WarehouseDelete is the Delete input.
type WarehouseDelete struct {
	ID   string  `normalize:"trim"`
	Etag *string `normalize:"trim,nil_on_empty"`
}

// WarehouseDeactivate is the Deactivate input. Blocked while the warehouse
// still carries stock (see services/warehouse.InventoryChecker).
type WarehouseDeactivate struct {
	ID   string  `normalize:"trim"`
	Etag *string `normalize:"trim,nil_on_empty"`
}

// WarehousesListSortField is int32-aligned with WarehousesListRequest.Sort.Field.
type WarehousesListSortField int32

const (
	WarehousesListSortFieldCreatedAt WarehousesListSortField = iota
	WarehousesListSortFieldName
)

type WarehousesListSort struct {
	Field     WarehousesListSortField
	Direction SortDirection
}

// WarehousesList is the single List input; scope/filters/sort/pagination all
// live inside it, per the List(ctx, in *XxxList) convention.
type WarehousesList struct {
	Statuses          []WarehouseStatus
	CreatedAt         *PeriodFilter
	Sort              WarehousesListSort
	Pagination        ListPagination
	IncludeTotalCount bool
}
