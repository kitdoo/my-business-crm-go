// Package inventorymovements defines the InventoryMovement storage
// interface; the MongoDB implementation lives in the mongo subpackage. The
// ledger is append-only — there is no Update/SoftDelete.
package inventorymovements

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Storage persists inventory movements.
type Storage interface {
	Insert(ctx context.Context, m *entities.InventoryMovement) error
	List(ctx context.Context, in *entities.InventoryMovementsList) (*entities.List[entities.InventoryMovement], error)
	GetHistory(ctx context.Context, in *entities.InventoryMovementGetHistory) (*entities.List[entities.InventoryMovement], error)
}
