// Package inventory defines the Inventory service interface; the
// implementation lives in the inventory subpackage.
package inventory

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Service reads and adjusts stock levels on top of storages.Storage.
type Service interface {
	Get(ctx context.Context, skuID, warehouseID string) (*entities.Inventory, error)
	List(ctx context.Context, in *entities.InventoryList) (*entities.List[entities.Inventory], error)
	// ApplyMovement atomically adjusts quantity by delta (positive or
	// negative) for (skuID, warehouseID), creating the row on first
	// touch. Returns errs.ErrInsufficientStock if delta would drive
	// quantity below zero. By convention, only inventorymovement.Service
	// calls this — every other stock change should go through recording a
	// movement there first, not this directly.
	ApplyMovement(ctx context.Context, skuID, warehouseID string, delta int64) (*entities.Inventory, error)
	// HasStock reports whether warehouseID carries any non-zero quantity;
	// backs warehouse.InventoryChecker.
	HasStock(ctx context.Context, warehouseID string) (bool, error)
}
