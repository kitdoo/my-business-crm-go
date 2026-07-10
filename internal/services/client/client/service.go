// Package client implements the client.Service interface.
package client

import (
	"context"
	"log/slog"
	"time"

	"github.com/altessa-s/go-atlas/domain/normalizer"
	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	clientsvc "github.com/kitdoo/my-business-crm-go/internal/services/client"
	"github.com/kitdoo/my-business-crm-go/internal/storages/clients"
)

var _ clientsvc.Service = (*Service)(nil)

// Service is the client.Service implementation.
type Service struct {
	storage clients.Storage
	logger  *slog.Logger
}

// New builds a Service.
func New(storage clients.Storage) *Service {
	return &Service{
		storage: storage,
		logger:  slog.Default().With(slogx.Module("service:client")),
	}
}

func (s *Service) Create(ctx context.Context, in *entities.ClientCreate) (*entities.Client, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	c := entities.ClientNew()
	in.Merge(c)

	if err := s.storage.Insert(ctx, c); err != nil {
		s.logger.DebugContext(ctx, "insert client failed", slogx.Error(err))
		return nil, err
	}
	return c, nil
}

func (s *Service) Get(ctx context.Context, id string) (*entities.Client, error) {
	return s.storage.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, in *entities.ClientsList) (*entities.List[entities.Client], error) {
	return s.storage.List(ctx, in)
}

func (s *Service) Update(ctx context.Context, in *entities.ClientUpdate) (*entities.Client, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	c, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if in.Etag != nil && *in.Etag != c.Etag {
		return nil, errs.ErrStaleEntity
	}
	oldEtag := c.Etag
	in.Merge(c)
	c.BeforeUpdate()
	if err := s.storage.Update(ctx, c, oldEtag); err != nil {
		s.logger.DebugContext(ctx, "update client failed", slog.String("id", c.ID), slogx.Error(err))
		return nil, err
	}
	return c, nil
}

func (s *Service) Delete(ctx context.Context, in *entities.ClientDelete) error {
	_ = normalizer.Normalize(in) //nolint:errcheck

	c, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		return err
	}
	if in.Etag != nil && *in.Etag != c.Etag {
		return errs.ErrStaleEntity
	}

	oldEtag := c.Etag
	now := time.Now().UTC()
	c.DeletedAt = &now
	c.BeforeUpdate()
	if err := s.storage.SoftDelete(ctx, &entities.SoftDelete{
		ID:           c.ID,
		Etag:         oldEtag,
		NewUpdatedAt: *c.DeletedAt,
		NewEtag:      c.Etag,
	}); err != nil {
		s.logger.DebugContext(ctx, "soft delete client failed", slog.String("id", c.ID), slogx.Error(err))
		return err
	}
	return nil
}
