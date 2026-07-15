// Package productsku defines the ProductSKU service interface; the
// implementation lives in the productsku subpackage.
package productsku

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Service orchestrates product SKU business rules on top of
// storages.Storage.
type Service interface {
	// Create returns errs.ErrProductSkuVariantNotFound if variantId does
	// not resolve to an existing variant.
	Create(ctx context.Context, in *entities.ProductSKUCreate) (*entities.ProductSKU, error)
	Get(ctx context.Context, id string) (*entities.ProductSKU, error)
	List(ctx context.Context, in *entities.ProductSKUsList) (*entities.List[entities.ProductSKU], error)
	Update(ctx context.Context, in *entities.ProductSKUUpdate) (*entities.ProductSKU, error)
	Delete(ctx context.Context, in *entities.ProductSKUDelete) error
}
