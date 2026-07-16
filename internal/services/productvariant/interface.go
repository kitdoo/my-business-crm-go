// Package productvariant defines the ProductVariant service interface;
// the implementation lives in the productvariant subpackage.
package productvariant

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Service orchestrates product variant business rules on top of
// storages.Storage.
type Service interface {
	// Create returns errs.ErrProductVariantProductNotFound if productId
	// does not resolve to an existing product.
	Create(ctx context.Context, in *entities.ProductVariantCreate) (*entities.ProductVariant, error)
	Get(ctx context.Context, id string) (*entities.ProductVariant, error)
	List(ctx context.Context, in *entities.ProductVariantsList) (*entities.List[entities.ProductVariant], error)
	Update(ctx context.Context, in *entities.ProductVariantUpdate) (*entities.ProductVariant, error)
	// Delete cascades: any active ProductSKU still referencing the
	// variant is deactivated (Status=Inactive) rather than blocking the
	// variant's own deletion.
	Delete(ctx context.Context, in *entities.ProductVariantDelete) error
}

// SKUsExistenceChecker reports whether any active ProductSKU still
// references a variant, and cascades their deactivation on Delete. This is
// satisfied by productskus.Storage directly (not productsku.Service) as a
// deliberate, narrow exception to "depend on the foreign entity's Service"
// (see SERVICE_DEVELOPMENT_STANDARD.md, Services Layer §1):
// productsku.Service already depends on productvariant.Service for its own
// Create FK validation, so wiring this direction through
// productsku.Service too would be a circular dependency.
type SKUsExistenceChecker interface {
	ExistsForVariant(ctx context.Context, variantID string) (bool, error)
	// DeactivateForVariant sets every active SKU of variantID to
	// Inactive. Called by Delete in place of the guard the name might
	// suggest.
	DeactivateForVariant(ctx context.Context, variantID string) error
}
