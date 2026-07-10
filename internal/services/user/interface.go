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
	// Login verifies login (phone or email) + password and issues a
	// self-contained, HMAC-signed session token that expires after the
	// configured TTL — nothing is persisted for the session itself.
	// Returns errs.ErrUserInvalidCredentials on mismatch and
	// errs.ErrUserInactive if the user's status is not active.
	Login(ctx context.Context, in *entities.UserLogin) (token string, u *entities.User, err error)
	// ChangePassword verifies currentPassword before setting newPassword.
	// Returns errs.ErrUserInvalidCredentials on mismatch.
	ChangePassword(ctx context.Context, in *entities.UserChangePassword) error
	// Authenticate verifies a token (as issued by Login) and resolves it to
	// its active owner. Returns errs.ErrUnauthenticated if the token is
	// malformed, has an invalid signature, has expired, or its owner is no
	// longer an active user. Used by both the gRPC auth interceptor
	// (internal/transports/grpc/interceptors/auth) and non-gRPC transports
	// (e.g. image upload) that need to identify the caller.
	Authenticate(ctx context.Context, token string) (*entities.User, error)
}
