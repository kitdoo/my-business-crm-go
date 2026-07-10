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
// references a brand. The products aggregate does not exist yet in this
// codebase, so no implementation is wired in fx today; Service.Delete
// treats a nil checker as "no products aggregate to check against" and
// skips the guard entirely.
type ProductsExistenceChecker interface {
	ExistsForBrand(ctx context.Context, brandID string) (bool, error)
}
