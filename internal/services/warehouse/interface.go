// Package warehouse defines the Warehouse service interface; the
// implementation lives in the warehouse subpackage.
package warehouse

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Service orchestrates warehouse business rules on top of storages.Storage.
type Service interface {
	Create(ctx context.Context, in *entities.WarehouseCreate) (*entities.Warehouse, error)
	Get(ctx context.Context, id string) (*entities.Warehouse, error)
	List(ctx context.Context, in *entities.WarehousesList) (*entities.List[entities.Warehouse], error)
	Update(ctx context.Context, in *entities.WarehouseUpdate) (*entities.Warehouse, error)
	// Delete soft-deletes a warehouse. Returns errs.ErrWarehouseHasStock if
	// inventory.Service reports non-zero stock.
	Delete(ctx context.Context, in *entities.WarehouseDelete) error
	// Deactivate flips status to Inactive. Returns errs.ErrWarehouseHasStock
	// if inventory.Service reports non-zero stock; the caller is expected
	// to have transferred stock away first.
	Deactivate(ctx context.Context, in *entities.WarehouseDeactivate) (*entities.Warehouse, error)
}
