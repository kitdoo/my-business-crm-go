package entities

import (
	"time"

	"github.com/google/uuid"

	"github.com/altessa-s/go-atlas/domain/converter"
)

// MovementType is int32-aligned with
// crm.types.inventory_movement.MovementType so converter.Convert maps it
// both ways as a plain scalar.
type MovementType int32

const (
	MovementTypeUnspecified MovementType = iota
	MovementTypeReceipt
	MovementTypeSale
	MovementTypeWriteOff
	MovementTypeAdjustment
	// MovementTypeTransfer is warehouse-to-warehouse; recorded as a paired
	// debit/credit movement — the caller issues two Create calls, one per
	// leg, since a single movement always targets one (variantId,
	// warehouseId) pair.
	MovementTypeTransfer
)

// InventoryMovement is an immutable stock ledger entry — no
// UpdatedAt/DeletedAt/Etag, no Update/SoftDelete, per the TD's exception to
// the general entity rules for this aggregate.
type InventoryMovement struct {
	ID          string
	VariantID   string
	WarehouseID string
	Type        MovementType
	// Quantity is signed: positive for Receipt, negative for Sale/WriteOff.
	Quantity int64
	Comment  string
	// SaleID links this movement to the Sale that caused it — always set
	// on Type=Sale (SalesService.Create sets it directly, not just via
	// Comment text); optionally settable on a manual Create too, for a
	// non-sale movement an operator wants to associate with a sale by
	// hand.
	SaleID    *string
	CreatedBy string
	CreatedAt time.Time
}

// InventoryMovementNew creates an InventoryMovement with a fresh ID and
// timestamp.
func InventoryMovementNew(init ...func(*InventoryMovement)) *InventoryMovement {
	m := &InventoryMovement{ID: uuid.NewString(), CreatedAt: time.Now().UTC()}
	if len(init) > 0 {
		init[0](m)
	}
	return m
}

// InventoryMovementCreate is the Create input. CreatedBy is set by the
// handler from the request context (see internal/pkg/reqctx), not by the
// client — the proto Create request carries no createdBy field.
type InventoryMovementCreate struct {
	VariantID   string `normalize:"trim"`
	WarehouseID string `normalize:"trim"`
	Type        MovementType
	Quantity    int64
	Comment     string  `normalize:"trim"`
	SaleID      *string `normalize:"trim,nil_on_empty"`
	CreatedBy   string  `normalize:"trim"`
}

func (c *InventoryMovementCreate) Merge(dst *InventoryMovement) *InventoryMovement {
	if c == nil || dst == nil {
		return dst
	}
	converter.Convert(c, dst, converter.WithIgnoreNilValues())
	return dst
}

// InventoryMovementsList is the single List input; scope/filters/pagination
// all live inside it, per the List(ctx, in *XxxList) convention.
type InventoryMovementsList struct {
	WarehouseID *string
	Types       []MovementType
	VariantIDs  []string
	CreatedBy   []string
	CreatedAt   *PeriodFilter
	Pagination  ListPagination
}

// InventoryMovementGetHistory is the GetHistory input — the full ledger for
// one (variantId, warehouseId) pair.
type InventoryMovementGetHistory struct {
	VariantID   string
	WarehouseID string
	Types       []MovementType
	CreatedAt   *PeriodFilter
	Pagination  ListPagination
}
