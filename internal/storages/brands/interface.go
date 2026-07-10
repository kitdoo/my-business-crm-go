// Package brands defines the Brand storage interface; the MongoDB
// implementation lives in the mongo subpackage.
package brands

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Storage persists brands. All reads filter out soft-deleted rows.
type Storage interface {
	Insert(ctx context.Context, b *entities.Brand) error
	Get(ctx context.Context, id string) (*entities.Brand, error)
	List(ctx context.Context, in *entities.BrandsList) (*entities.List[entities.Brand], error)
	// Update writes b, guarding on oldEtag when non-empty. Returns
	// errs.ErrStaleEntity if the guard fails to match.
	Update(ctx context.Context, b *entities.Brand, oldEtag string) error
	// SoftDelete hides a brand by stamping deleted_at. The caller supplies
	// the new timestamp and etag via in, so this method generates nothing.
	SoftDelete(ctx context.Context, in *entities.SoftDelete) error
}
