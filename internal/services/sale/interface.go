// Package sale defines the Sale service interface; the implementation
// lives in the sale subpackage.
package sale

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Service orchestrates sale creation and status transitions on top of
// storages.Storage, Inventory, and InventoryMovements. There is no generic
// Update — items are immutable and status moves only through UpdateStatus
// and Cancel.
type Service interface {
	// Create validates client/warehouse/partner/product existence, checks
	// stock availability, captures each item's current price, then records
	// one Sale-type InventoryMovement per item (decrementing Inventory).
	// Returns errs.ErrInsufficientStock if any item's quantity exceeds
	// available stock.
	Create(ctx context.Context, in *entities.SaleCreate) (*entities.Sale, error)
	Get(ctx context.Context, id string) (*entities.Sale, error)
	List(ctx context.Context, in *entities.SalesList) (*entities.List[entities.Sale], error)
	// UpdateStatus returns errs.ErrSaleTerminalStatus if the sale is
	// already cancelled or refunded.
	UpdateStatus(ctx context.Context, in *entities.SaleUpdateStatus) (*entities.Sale, error)
	// Cancel reverses the sale's stock impact via an Adjustment-type
	// InventoryMovement per item, then marks the sale cancelled. Returns
	// errs.ErrSaleTerminalStatus if already cancelled or refunded.
	Cancel(ctx context.Context, in *entities.SaleCancel) (*entities.Sale, error)
}
