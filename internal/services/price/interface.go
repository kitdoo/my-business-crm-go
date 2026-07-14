// Package price defines the ProductPrice service interface; the
// implementation lives in the price subpackage.
package price

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Service orchestrates product-price business rules on top of storages.Storage.
type Service interface {
	// Create returns errs.ErrProductVariantNotFound if variantId does not
	// resolve to an existing variant, or errs.ErrProductPriceExists if the
	// variant already has an active price.
	Create(ctx context.Context, in *entities.ProductPriceCreate) (*entities.ProductPrice, error)
	// Get looks up the current price for a variant.
	Get(ctx context.Context, variantID string) (*entities.ProductPrice, error)
	// Update snapshots the pre-change values into the price history log
	// before applying in.
	Update(ctx context.Context, in *entities.ProductPriceUpdate) (*entities.ProductPrice, error)
	Delete(ctx context.Context, in *entities.ProductPriceDelete) error
	GetHistory(ctx context.Context, in *entities.ProductPriceGetHistory) (*entities.List[entities.ProductPrice], error)
}
