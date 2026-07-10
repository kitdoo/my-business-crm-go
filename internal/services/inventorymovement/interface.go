// Package inventorymovement defines the InventoryMovement service
// interface; the implementation lives in the inventorymovement subpackage.
package inventorymovement

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Service orchestrates inventory movement recording on top of
// storages.Storage. Every Create applies the movement to Inventory
// (creating the row on first touch) before appending the immutable ledger
// entry; see errs.ErrInsufficientStock for the stock floor guard.
type Service interface {
	Create(ctx context.Context, in *entities.InventoryMovementCreate) (*entities.InventoryMovement, error)
	List(ctx context.Context, in *entities.InventoryMovementsList) (*entities.List[entities.InventoryMovement], error)
	GetHistory(ctx context.Context, in *entities.InventoryMovementGetHistory) (*entities.List[entities.InventoryMovement], error)
}
