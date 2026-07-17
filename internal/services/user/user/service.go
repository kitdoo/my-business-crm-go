// Package user implements the user.Service interface.
package user

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/altessa-s/go-atlas/domain/normalizer"
	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	usersvc "github.com/kitdoo/my-business-crm-go/internal/services/user"
	"github.com/kitdoo/my-business-crm-go/internal/storages/users"
)

var _ usersvc.Service = (*Service)(nil)

// Service is the user.Service implementation. Session tokens are
// self-contained HMAC-signed bearer tokens (see token.go) — no session
// state (raw token, hash, or otherwise) is ever persisted to storage.
type Service struct {
	storage        users.Storage
	logger         *slog.Logger
	signingKey     []byte
	tokenTTL       time.Duration
	requiredLocale string
}

// New builds a Service. signingKey must be non-empty — it is the HMAC-SHA256
// secret used to sign and verify session tokens issued by Login; every
// process that must accept those tokens (every gRPC/HTTP node) needs the
// same key. tokenTTL <= 0 falls back to defaultTokenTTL. requiredLocale is
// the locale every non-empty LocalizedString field must include (see
// entities.LocalizedString.Validate); it comes from CRMConfig.DefaultLocale.
func New(storage users.Storage, signingKey []byte, tokenTTL time.Duration, requiredLocale string) (*Service, error) {
	if len(signingKey) == 0 {
		return nil, fmt.Errorf("%w: user session signing key must be configured", errs.ErrInvalidArgument)
	}
	if tokenTTL <= 0 {
		tokenTTL = defaultTokenTTL
	}
	return &Service{
		storage:        storage,
		logger:         slog.Default().With(slogx.Module("service:user")),
		signingKey:     signingKey,
		tokenTTL:       tokenTTL,
		requiredLocale: requiredLocale,
	}, nil
}

// defaultTokenTTL is used when the configured tokenTTL is unset.
const defaultTokenTTL = 72 * time.Hour

// hashPassword bcrypts plaintext, translating bcrypt's own length guard into
// the domain's invalid-argument sentinel.
func hashPassword(plaintext string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("%w: %s", errs.ErrInvalidArgument, err.Error())
	}
	return string(hash), nil
}

func (s *Service) Create(ctx context.Context, in *entities.UserCreate) (*entities.User, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	if err := in.Validate(s.requiredLocale); err != nil {
		return nil, err
	}

	hash, err := hashPassword(in.Password)
	if err != nil {
		return nil, err
	}
	in.PasswordHash = hash

	u := entities.UserNew()
	in.Merge(u)

	if err := s.storage.Insert(ctx, u); err != nil {
		s.logger.DebugContext(ctx, "insert user failed", slog.String("error", err.Error()))
		return nil, err
	}
	return u, nil
}

func (s *Service) Get(ctx context.Context, id string) (*entities.User, error) {
	u, err := s.storage.Get(ctx, id)
	if err != nil {
		s.logger.DebugContext(ctx, "get user failed", slog.String("id", id), slog.String("error", err.Error()))
		return nil, err
	}
	return u, nil
}

func (s *Service) List(ctx context.Context, in *entities.UsersList) (*entities.List[entities.User], error) {
	list, err := s.storage.List(ctx, in)
	if err != nil {
		s.logger.DebugContext(ctx, "list users failed", slog.String("error", err.Error()))
		return nil, err
	}
	return list, nil
}

func (s *Service) Update(ctx context.Context, in *entities.UserUpdate) (*entities.User, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	if err := in.Validate(s.requiredLocale); err != nil {
		return nil, err
	}

	u, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		s.logger.DebugContext(ctx, "get user failed", slog.String("id", in.ID), slog.String("error", err.Error()))
		return nil, err
	}
	if in.Etag != nil && *in.Etag != u.Etag {
		return nil, errs.ErrStaleEntity
	}
	oldEtag := u.Etag
	in.Merge(u)
	u.BeforeUpdate()
	if err := s.storage.Update(ctx, u, oldEtag); err != nil {
		s.logger.DebugContext(ctx, "update user failed", slog.String("id", u.ID), slog.String("error", err.Error()))
		return nil, err
	}
	return u, nil
}

func (s *Service) Delete(ctx context.Context, in *entities.UserDelete) error {
	_ = normalizer.Normalize(in) //nolint:errcheck

	u, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		s.logger.DebugContext(ctx, "get user failed", slog.String("id", in.ID), slog.String("error", err.Error()))
		return err
	}
	if u.Role == entities.UserRoleAdmin {
		return errs.ErrUserAdminProtected
	}
	if in.Etag != nil && *in.Etag != u.Etag {
		return errs.ErrStaleEntity
	}

	oldEtag := u.Etag
	now := time.Now().UTC()
	u.DeletedAt = &now
	u.BeforeUpdate()
	if err := s.storage.SoftDelete(ctx, &entities.SoftDelete{
		ID:           u.ID,
		Etag:         oldEtag,
		NewUpdatedAt: *u.DeletedAt,
		NewEtag:      u.Etag,
	}); err != nil {
		s.logger.DebugContext(ctx, "soft delete user failed", slog.String("id", u.ID), slog.String("error", err.Error()))
		return err
	}
	return nil
}

func (s *Service) Login(ctx context.Context, in *entities.UserLogin) (string, *entities.User, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	u, err := s.storage.GetByLogin(ctx, in.Login)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return "", nil, errs.ErrUserInvalidCredentials
		}
		s.logger.DebugContext(ctx, "get user by login failed", slog.String("error", err.Error()))
		return "", nil, err
	}
	if u.Status != entities.UserStatusActive {
		return "", nil, errs.ErrUserInactive
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password)); err != nil {
		return "", nil, errs.ErrUserInvalidCredentials
	}

	return s.issueToken(u.ID), u, nil
}

// Authenticate verifies token's signature and expiry (see token.go — no
// storage lookup is needed for that part, since nothing is persisted) and
// then loads the current user record to confirm it still exists and is
// active. This is what lets ChangePassword/Delete/deactivation take effect
// immediately even though the token itself remains valid until it expires.
func (s *Service) Authenticate(ctx context.Context, token string) (*entities.User, error) {
	if token == "" {
		return nil, errs.ErrUnauthenticated
	}

	userID, ok := s.verifyToken(token)
	if !ok {
		return nil, errs.ErrUnauthenticated
	}

	u, err := s.storage.Get(ctx, userID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return nil, errs.ErrUnauthenticated
		}
		s.logger.DebugContext(ctx, "get user for authenticate failed", slog.String("id", userID), slog.String("error", err.Error()))
		return nil, err
	}
	if u.Status != entities.UserStatusActive {
		return nil, errs.ErrUnauthenticated
	}
	return u, nil
}

func (s *Service) ChangePassword(ctx context.Context, in *entities.UserChangePassword) error {
	_ = normalizer.Normalize(in) //nolint:errcheck

	u, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		s.logger.DebugContext(ctx, "get user for change password failed", slog.String("id", in.ID), slog.String("error", err.Error()))
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.CurrentPassword)); err != nil {
		return errs.ErrUserInvalidCredentials
	}

	hash, err := hashPassword(in.NewPassword)
	if err != nil {
		return err
	}

	oldEtag := u.Etag
	u.PasswordHash = hash
	u.BeforeUpdate()
	if err := s.storage.Update(ctx, u, oldEtag); err != nil {
		s.logger.DebugContext(ctx, "change password failed", slog.String("id", u.ID), slog.String("error", err.Error()))
		return err
	}
	return nil
}
