// Package warehouses defines the Warehouse storage interface; the MongoDB
// implementation lives in the mongo subpackage.
package warehouses

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Storage persists warehouses. All reads filter out soft-deleted rows.
type Storage interface {
	Insert(ctx context.Context, w *entities.Warehouse) error
	Get(ctx context.Context, id string) (*entities.Warehouse, error)
	List(ctx context.Context, in *entities.WarehousesList) (*entities.List[entities.Warehouse], error)
	// Update writes w, guarding on oldEtag when non-empty. Returns
	// errs.ErrStaleEntity if the guard fails to match.
	Update(ctx context.Context, w *entities.Warehouse, oldEtag string) error
	// SoftDelete hides a warehouse by stamping deleted_at. The caller
	// supplies the new timestamp and etag via in, so this method generates
	// nothing.
	SoftDelete(ctx context.Context, in *entities.SoftDelete) error
}
