// Package inventory defines the Inventory storage interface; the MongoDB
// implementation lives in the mongo subpackage. Inventory has no
// Create/Update/Delete of its own — only ApplyMovement mutates it, called
// by the InventoryMovements service.
package inventory

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Storage persists stock levels. Get is cache-aside over Redis per the TD
// (fast reads of a hot row); ApplyMovement invalidates the cache entry it
// touches.
type Storage interface {
	// Get returns errs.ErrInventoryNotFound if no row exists yet for the pair.
	Get(ctx context.Context, variantID, warehouseID string) (*entities.Inventory, error)
	List(ctx context.Context, in *entities.InventoryList) (*entities.List[entities.Inventory], error)
	// ApplyMovement atomically adjusts quantity by delta (positive or
	// negative), creating the row on first touch. Returns
	// errs.ErrInsufficientStock if delta would drive quantity below zero.
	ApplyMovement(ctx context.Context, variantID, warehouseID string, delta int64) (*entities.Inventory, error)
	// HasStock reports whether warehouseID carries any non-zero quantity;
	// backs warehouse.InventoryChecker.
	HasStock(ctx context.Context, warehouseID string) (bool, error)
}
