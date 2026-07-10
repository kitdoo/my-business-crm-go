// Package users defines the User storage interface; the MongoDB
// implementation lives in the mongo subpackage.
package users

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Storage persists users. All reads filter out soft-deleted rows.
type Storage interface {
	// Insert returns errs.ErrUserPhoneConflict / errs.ErrUserEmailConflict
	// when phone/email collides with another active user.
	Insert(ctx context.Context, u *entities.User) error
	Get(ctx context.Context, id string) (*entities.User, error)
	// GetByLogin looks up an active user by phone or email, whichever
	// matches login. Returns errs.ErrUserNotFound if neither matches.
	GetByLogin(ctx context.Context, login string) (*entities.User, error)
	List(ctx context.Context, in *entities.UsersList) (*entities.List[entities.User], error)
	// Update writes u, guarding on oldEtag when non-empty. Returns
	// errs.ErrStaleEntity if the guard fails to match, or a conflict
	// sentinel on a phone/email collision.
	Update(ctx context.Context, u *entities.User, oldEtag string) error
	// SoftDelete hides a user by stamping deleted_at. The caller supplies
	// the new timestamp and etag via in, so this method generates nothing.
	SoftDelete(ctx context.Context, in *entities.SoftDelete) error
}
