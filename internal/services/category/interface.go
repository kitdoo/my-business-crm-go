// Package category defines the Category service interface; the
// implementation lives in the category subpackage.
package category

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Service orchestrates category business rules on top of storages.Storage.
type Service interface {
	Create(ctx context.Context, in *entities.CategoryCreate) (*entities.Category, error)
	Get(ctx context.Context, id string) (*entities.Category, error)
	List(ctx context.Context, in *entities.CategoriesList) (*entities.List[entities.Category], error)
	Update(ctx context.Context, in *entities.CategoryUpdate) (*entities.Category, error)
	// Delete soft-deletes a category. Returns errs.ErrCategoryHasProducts if a
	// ProductsExistenceChecker is configured and reports active products.
	Delete(ctx context.Context, in *entities.CategoryDelete) error
}

// ProductsExistenceChecker reports whether any active product still
// references a category. The products aggregate does not exist yet in this
// codebase, so no implementation is wired in fx today; Service.Delete
// treats a nil checker as "no products aggregate to check against" and
// skips the guard entirely.
type ProductsExistenceChecker interface {
	ExistsForCategory(ctx context.Context, categoryID string) (bool, error)
}
