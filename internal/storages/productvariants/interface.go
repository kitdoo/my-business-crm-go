// Package productvariants defines the ProductVariant storage interface;
// the MongoDB implementation lives in the mongo subpackage.
package productvariants

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Storage persists product variants. All reads filter out soft-deleted
// rows.
type Storage interface {
	// Insert returns errs.ErrProductVariantSKUConflict when sku collides
	// with another active variant.
	Insert(ctx context.Context, v *entities.ProductVariant) error
	Get(ctx context.Context, id string) (*entities.ProductVariant, error)
	List(ctx context.Context, in *entities.ProductVariantsList) (*entities.List[entities.ProductVariant], error)
	// Update writes v, guarding on oldEtag when non-empty. Returns
	// errs.ErrStaleEntity if the guard fails to match, or
	// errs.ErrProductVariantSKUConflict on a SKU collision.
	Update(ctx context.Context, v *entities.ProductVariant, oldEtag string) error
	// SoftDelete hides a variant by stamping deleted_at. The caller
	// supplies the new timestamp and etag via in, so this method generates
	// nothing.
	SoftDelete(ctx context.Context, in *entities.SoftDelete) error
	// ExistsForProduct reports whether any active variant still
	// references productID — used to block Product.Delete while variants
	// remain.
	ExistsForProduct(ctx context.Context, productID string) (bool, error)
}
