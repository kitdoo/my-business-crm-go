// Package user defines the User service interface; the implementation
// lives in the user subpackage.
package user

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Service orchestrates user business rules on top of storages.Storage.
type Service interface {
	Create(ctx context.Context, in *entities.UserCreate) (*entities.User, error)
	Get(ctx context.Context, id string) (*entities.User, error)
	List(ctx context.Context, in *entities.UsersList) (*entities.List[entities.User], error)
	Update(ctx context.Context, in *entities.UserUpdate) (*entities.User, error)
	Delete(ctx context.Context, in *entities.UserDelete) error
	// Login verifies login (phone or email) + password and issues a fresh,
	// non-expiring session token. Returns errs.ErrUserInvalidCredentials on
	// mismatch and errs.ErrUserInactive if the user's status is not active.
	Login(ctx context.Context, in *entities.UserLogin) (token string, u *entities.User, err error)
	// ChangePassword verifies currentPassword before setting newPassword.
	// Returns errs.ErrUserInvalidCredentials on mismatch.
	ChangePassword(ctx context.Context, in *entities.UserChangePassword) error
}
