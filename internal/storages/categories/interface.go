// Package categories defines the Category storage interface; the MongoDB
// implementation lives in the mongo subpackage.
package categories

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Storage persists categories. All reads filter out soft-deleted rows.
type Storage interface {
	Insert(ctx context.Context, c *entities.Category) error
	Get(ctx context.Context, id string) (*entities.Category, error)
	List(ctx context.Context, in *entities.CategoriesList) (*entities.List[entities.Category], error)
	// Update writes c, guarding on oldEtag when non-empty. Returns
	// errs.ErrStaleEntity if the guard fails to match.
	Update(ctx context.Context, c *entities.Category, oldEtag string) error
	// SoftDelete hides a category by stamping deleted_at. The caller supplies
	// the new timestamp and etag via in, so this method generates nothing.
	SoftDelete(ctx context.Context, in *entities.SoftDelete) error
}
