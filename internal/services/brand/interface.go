// Package brand defines the Brand service interface; the implementation
// lives in the brand subpackage.
package brand

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Service orchestrates brand business rules on top of storages.Storage.
type Service interface {
	Create(ctx context.Context, in *entities.BrandCreate) (*entities.Brand, error)
	Get(ctx context.Context, id string) (*entities.Brand, error)
	List(ctx context.Context, in *entities.BrandsList) (*entities.List[entities.Brand], error)
	Update(ctx context.Context, in *entities.BrandUpdate) (*entities.Brand, error)
	// Delete soft-deletes a brand. Returns errs.ErrBrandHasProducts if a
	// ProductsExistenceChecker is configured and reports active products.
	Delete(ctx context.Context, in *entities.BrandDelete) error
}

// ProductsExistenceChecker reports whether any active product still
// references a brand. This is satisfied by products.Storage directly
// (not product.Service) as a deliberate, narrow exception to "depend on
// the foreign entity's Service" (see SERVICE_DEVELOPMENT_STANDARD.md,
// Services Layer §1): product.Service already depends on brand.Service
// for its own Create/Update FK validation, so wiring this direction
// through product.Service too would be a circular dependency. A nil
// checker is treated as "no products aggregate to check against" and
// skips the guard entirely (used only where fx cannot wire a real one).
type ProductsExistenceChecker interface {
	ExistsForBrand(ctx context.Context, brandID string) (bool, error)
}
