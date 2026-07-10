// Package inventory defines the Inventory service interface; the
// implementation lives in the inventory subpackage. Inventory is read-only
// at this layer — stock changes only through the InventoryMovements
// service.
package inventory

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Service reads stock levels on top of storages.Storage.
type Service interface {
	Get(ctx context.Context, productID, warehouseID string) (*entities.Inventory, error)
	List(ctx context.Context, in *entities.InventoryList) (*entities.List[entities.Inventory], error)
}
