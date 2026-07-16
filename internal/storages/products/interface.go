// Package products defines the Product storage interface; the MongoDB
// implementation lives in the mongo subpackage.
package products

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Storage persists products. All reads filter out soft-deleted rows.
//
// Storage also satisfies brand.ProductsExistenceChecker and
// category.ProductsExistenceChecker via ExistsForBrand/ExistsForCategory, so
// it is wired directly as both once the Brand/Category delete guards need a
// real implementation instead of the nil placeholder.
type Storage interface {
	// Insert returns errs.ErrProductSKUConflict when sku collides with
	// another active product.
	Insert(ctx context.Context, p *entities.Product) error
	Get(ctx context.Context, id string) (*entities.Product, error)
	List(ctx context.Context, in *entities.ProductsList) (*entities.List[entities.Product], error)
	// Update writes p, guarding on oldEtag when non-empty. Returns
	// errs.ErrStaleEntity if the guard fails to match, or
	// errs.ErrProductSKUConflict on a SKU collision.
	Update(ctx context.Context, p *entities.Product, oldEtag string) error
	// SoftDelete hides a product by stamping deleted_at. The caller
	// supplies the new timestamp and etag via in, so this method generates
	// nothing.
	SoftDelete(ctx context.Context, in *entities.SoftDelete) error
	// ExistsForBrand reports whether any active product still references brandID.
	ExistsForBrand(ctx context.Context, brandID string) (bool, error)
	// ExistsForCategory reports whether any active product still references categoryID.
	ExistsForCategory(ctx context.Context, categoryID string) (bool, error)
	// SetHasStock writes the system-maintained HasStock flag (see
	// entities.Product.HasStock) without touching etag/updated_at — this
	// is a derived cache field, not a user-visible mutation. Returns
	// errs.ErrProductNotFound if id has no active row.
	SetHasStock(ctx context.Context, id string, hasStock bool) error
}
