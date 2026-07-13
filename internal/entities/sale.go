package entities

import (
	"time"

	"github.com/google/uuid"
)

// SaleStatus is int32-aligned with crm.types.sale.SaleStatus so
// converter.Convert maps it both ways as a plain scalar.
type SaleStatus int32

const (
	SaleStatusUnspecified SaleStatus = iota
	SaleStatusDraft
	SaleStatusPaid
	SaleStatusShipped
	SaleStatusCompleted
	SaleStatusCancelled
	SaleStatusRefunded
)

// SaleItem is immutable once the Sale is created — see Sale.
type SaleItem struct {
	ProductID string
	Quantity  int64
	// PriceAmount is captured from the product's current ProductPrice at
	// creation time (basis points), so later price changes never affect an
	// existing sale.
	PriceAmount int64
	// DiscountPercentage is always 0-100, never a fixed amount.
	DiscountPercentage int32
}

// Sale is fulfilled from exactly one warehouse. Items are immutable once
// created — reverse a mistake with Cancel, not a generic Update (there is
// none).
type Sale struct {
	ID string
	// Number is a human-readable, sequential identifier assigned
	// atomically by the storage layer on Insert (see
	// storages/sales/mongo.Storage.nextNumber) — never set by callers.
	Number      int64
	ClientID    string
	WarehouseID string
	PartnerID   *string
	Items       []SaleItem
	// TotalAmount is basis points, sum of item lines after discount.
	TotalAmount int64
	Status      SaleStatus
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	// DeletedAt has no soft-delete path in this aggregate (the TD lists no
	// Delete method for Sale) — always nil. Kept for wire-schema parity
	// with crm.types.sale.Sale.
	DeletedAt *time.Time
	Etag      string // OCC token; rolled on every write
}

// SaleNew creates a Sale with a fresh ID, timestamps, and etag.
func SaleNew(init ...func(*Sale)) *Sale {
	s := &Sale{ID: uuid.NewString()}
	if len(init) > 0 {
		init[0](s)
	}
	s.UpdateTimestamps()
	s.UpdateEtag()
	return s
}

func (s *Sale) UpdateTimestamps() {
	now := time.Now().UTC()
	s.UpdatedAt = now
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
}

func (s *Sale) UpdateEtag() { s.Etag = uuid.NewString() }

func (s *Sale) BeforeUpdate() {
	s.UpdateTimestamps()
	s.UpdateEtag()
}

// SaleCreateItem is one requested line; PriceAmount is not part of it — the
// service looks up the product's current price.
type SaleCreateItem struct {
	ProductID          string `normalize:"trim"`
	Quantity           int64
	DiscountPercentage int32
}

// SaleCreate is the Create input. There is no generic Merge onto a fresh
// Sale here (unlike other aggregates): PriceAmount per item and TotalAmount
// are server-computed from current prices, not copied from the request, so
// the service assembles entities.Sale.Items itself.
// SaleCreate carries exactly one of ClientID (an existing client, picked
// by the caller) or NewClient (find-or-create by email — TD §12.3, the
// caller never has to create a client as a separate step first).
type SaleCreate struct {
	ClientID    string `normalize:"trim"`
	NewClient   *ClientCreate
	WarehouseID string  `normalize:"trim"`
	PartnerID   *string `normalize:"trim,nil_on_empty"`
	Items       []SaleCreateItem
	CreatedBy   string `normalize:"trim"`
}

// SaleUpdateStatus is the UpdateStatus input.
type SaleUpdateStatus struct {
	ID     string `normalize:"trim"`
	Status SaleStatus
	Etag   *string `normalize:"trim,nil_on_empty"`
}

// SaleCancel is the Cancel input. Reason is not persisted on the Sale (the
// proto type has no field for it) — it is carried onto the reversal
// InventoryMovement entries as their Comment. CreatedBy (the caller
// cancelling the sale) is likewise attributed only on those movements.
type SaleCancel struct {
	ID        string  `normalize:"trim"`
	Reason    *string `normalize:"trim,nil_on_empty"`
	CreatedBy string  `normalize:"trim"`
	Etag      *string `normalize:"trim,nil_on_empty"`
}

// SalesListSortField is int32-aligned with SalesListRequest.Sort.Field.
type SalesListSortField int32

const (
	SalesListSortFieldCreatedAt SalesListSortField = iota
)

type SalesListSort struct {
	Field     SalesListSortField
	Direction SortDirection
}

// SalesList is the single List input; scope/filters/sort/pagination all
// live inside it, per the List(ctx, in *XxxList) convention.
type SalesList struct {
	ClientID          *string `normalize:"trim,nil_on_empty"`
	WarehouseID       *string `normalize:"trim,nil_on_empty"`
	PartnerID         *string `normalize:"trim,nil_on_empty"`
	Statuses          []SaleStatus
	CreatedBy         []string
	ProductIDs        []string
	CreatedAt         *PeriodFilter
	Sort              SalesListSort
	Pagination        ListPagination
	IncludeTotalCount bool
}
