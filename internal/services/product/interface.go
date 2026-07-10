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
	Delete(ctx context.Context, in *entities.ProductDelete) error
}
