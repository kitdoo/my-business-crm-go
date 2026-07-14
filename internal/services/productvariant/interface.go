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
	Delete(ctx context.Context, in *entities.ProductVariantDelete) error
}
