// Package productskus defines the ProductSKU storage interface; the
// MongoDB implementation lives in the mongo subpackage.
package productskus

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Storage persists product SKUs. All reads filter out soft-deleted rows.
type Storage interface {
	// Insert returns errs.ErrProductSkuSKUConflict when sku collides with
	// another active SKU.
	Insert(ctx context.Context, s *entities.ProductSKU) error
	Get(ctx context.Context, id string) (*entities.ProductSKU, error)
	List(ctx context.Context, in *entities.ProductSKUsList) (*entities.List[entities.ProductSKU], error)
	// Update writes s, guarding on oldEtag when non-empty. Returns
	// errs.ErrStaleEntity if the guard fails to match, or
	// errs.ErrProductSkuSKUConflict on a SKU collision.
	Update(ctx context.Context, s *entities.ProductSKU, oldEtag string) error
	// SoftDelete hides a SKU by stamping deleted_at. The caller supplies
	// the new timestamp and etag via in, so this method generates nothing.
	SoftDelete(ctx context.Context, in *entities.SoftDelete) error
	// ExistsForVariant reports whether any active SKU still references
	// variantID — used to block ProductVariant.Delete while SKUs remain.
	ExistsForVariant(ctx context.Context, variantID string) (bool, error)
}
