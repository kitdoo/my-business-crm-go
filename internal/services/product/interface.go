// Package product defines the Product service interface; the implementation
// lives in the product subpackage.
package product

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Service orchestrates product business rules on top of storages.Storage.
type Service interface {
	// Create returns errs.ErrProductBrandNotFound / errs.ErrProductCategoryNotFound
	// if brandId / any categoryId does not resolve to an existing row.
	Create(ctx context.Context, in *entities.ProductCreate) (*entities.Product, error)
	Get(ctx context.Context, id string) (*entities.Product, error)
	List(ctx context.Context, in *entities.ProductsList) (*entities.List[entities.Product], error)
	Update(ctx context.Context, in *entities.ProductUpdate) (*entities.Product, error)
	// Delete cascades: any active ProductVariant still referencing the
	// product (and, transitively, their ProductSKUs) is deactivated
	// (Status=Inactive) rather than blocking the product's own deletion.
	Delete(ctx context.Context, in *entities.ProductDelete) error
	// SetHasStock writes the system-maintained HasStock flag (see
	// entities.Product.HasStock). Called by inventory.Service.ApplyMovement,
	// never by a transport handler — there is no RPC for it.
	SetHasStock(ctx context.Context, id string, hasStock bool) error
}

// VariantsExistenceChecker reports whether any active ProductVariant still
// references a product, and cascades their deactivation on Delete. This is
// satisfied by a Storage-shaped adapter (not productvariant.Service) as a
// deliberate, narrow exception to "depend on the foreign entity's Service"
// (see SERVICE_DEVELOPMENT_STANDARD.md, Services Layer §1):
// productvariant.Service already depends on product.Service for its own
// Create FK validation, so wiring this direction through
// productvariant.Service too would be a circular dependency. A nil checker
// is treated as "no variants aggregate to check against" and skips the
// cascade entirely (used only where fx cannot wire a real one).
type VariantsExistenceChecker interface {
	ExistsForProduct(ctx context.Context, productID string) (bool, error)
	// DeactivateForProduct sets every active variant of productID (and
	// their SKUs) to Inactive. Called by Delete in place of the guard the
	// name might suggest.
	DeactivateForProduct(ctx context.Context, productID string) error
}
