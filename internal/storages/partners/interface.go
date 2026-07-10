// Package partners defines the Partner storage interface; the MongoDB
// implementation lives in the mongo subpackage.
package partners

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Storage persists partners. All reads filter out soft-deleted rows.
type Storage interface {
	// Insert returns errs.ErrPartnerPhoneConflict when phone collides with
	// another active partner.
	Insert(ctx context.Context, p *entities.Partner) error
	Get(ctx context.Context, id string) (*entities.Partner, error)
	List(ctx context.Context, in *entities.PartnersList) (*entities.List[entities.Partner], error)
	// Update writes p, guarding on oldEtag when non-empty. Returns
	// errs.ErrStaleEntity if the guard fails to match, or
	// errs.ErrPartnerPhoneConflict on a phone collision.
	Update(ctx context.Context, p *entities.Partner, oldEtag string) error
	// SoftDelete hides a partner by stamping deleted_at. The caller supplies
	// the new timestamp and etag via in, so this method generates nothing.
	SoftDelete(ctx context.Context, in *entities.SoftDelete) error
}
