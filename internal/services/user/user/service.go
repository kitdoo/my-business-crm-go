// Package user implements the user.Service interface.
package user

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"

	corehash "github.com/altessa-s/go-atlas/core/encoding/hash"
	"github.com/altessa-s/go-atlas/domain/normalizer"
	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	usersvc "github.com/kitdoo/my-business-crm-go/internal/services/user"
	"github.com/kitdoo/my-business-crm-go/internal/storages/users"
)

var _ usersvc.Service = (*Service)(nil)

// Service is the user.Service implementation.
type Service struct {
	storage users.Storage
	logger  *slog.Logger
}

// New builds a Service.
func New(storage users.Storage) *Service {
	return &Service{
		storage: storage,
		logger:  slog.Default().With(slogx.Module("service:user")),
	}
}

// hashPassword bcrypts plaintext, translating bcrypt's own length guard into
// the domain's invalid-argument sentinel.
func hashPassword(plaintext string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("%w: %s", errs.ErrInvalidArgument, err.Error())
	}
	return string(hash), nil
}

// newSessionToken returns a fresh random token and the hash stored on the
// user document. Only the hash is persisted; the raw token is returned to
// the caller once and never stored.
func newSessionToken() (token, hash string, err error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", fmt.Errorf("generate session token: %w", err)
	}
	token = base64.RawURLEncoding.EncodeToString(raw)
	return token, corehash.SHA256HexString(token), nil
}

func (s *Service) Create(ctx context.Context, in *entities.UserCreate) (*entities.User, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	hash, err := hashPassword(in.Password)
	if err != nil {
		return nil, err
	}
	in.PasswordHash = hash

	u := entities.UserNew()
	in.Merge(u)

	if err := s.storage.Insert(ctx, u); err != nil {
		s.logger.DebugContext(ctx, "insert user failed", slogx.Error(err))
		return nil, err
	}
	return u, nil
}

func (s *Service) Get(ctx context.Context, id string) (*entities.User, error) {
	return s.storage.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, in *entities.UsersList) (*entities.List[entities.User], error) {
	return s.storage.List(ctx, in)
}

func (s *Service) Update(ctx context.Context, in *entities.UserUpdate) (*entities.User, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	u, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if in.Etag != nil && *in.Etag != u.Etag {
		return nil, errs.ErrStaleEntity
	}
	oldEtag := u.Etag
	in.Merge(u)
	u.BeforeUpdate()
	if err := s.storage.Update(ctx, u, oldEtag); err != nil {
		s.logger.DebugContext(ctx, "update user failed", slog.String("id", u.ID), slogx.Error(err))
		return nil, err
	}
	return u, nil
}

func (s *Service) Delete(ctx context.Context, in *entities.UserDelete) error {
	_ = normalizer.Normalize(in) //nolint:errcheck

	u, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		return err
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
		s.logger.DebugContext(ctx, "soft delete user failed", slog.String("id", u.ID), slogx.Error(err))
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
		return "", nil, err
	}
	if u.Status != entities.UserStatusActive {
		return "", nil, errs.ErrUserInactive
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password)); err != nil {
		return "", nil, errs.ErrUserInvalidCredentials
	}

	token, hash, err := newSessionToken()
	if err != nil {
		return "", nil, err
	}

	oldEtag := u.Etag
	u.TokenHash = &hash
	u.BeforeUpdate()
	if err := s.storage.Update(ctx, u, oldEtag); err != nil {
		s.logger.DebugContext(ctx, "persist login token failed", slog.String("id", u.ID), slogx.Error(err))
		return "", nil, err
	}
	return token, u, nil
}

func (s *Service) ChangePassword(ctx context.Context, in *entities.UserChangePassword) error {
	_ = normalizer.Normalize(in) //nolint:errcheck

	u, err := s.storage.Get(ctx, in.ID)
	if err != nil {
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
		s.logger.DebugContext(ctx, "change password failed", slog.String("id", u.ID), slogx.Error(err))
		return err
	}
	return nil
}
