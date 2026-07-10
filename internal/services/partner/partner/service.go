// Package partner implements the partner.Service interface.
package partner

import (
	"context"
	"log/slog"
	"time"

	"github.com/altessa-s/go-atlas/domain/normalizer"
	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	partnersvc "github.com/kitdoo/my-business-crm-go/internal/services/partner"
	"github.com/kitdoo/my-business-crm-go/internal/storages/partners"
)

var _ partnersvc.Service = (*Service)(nil)

// Service is the partner.Service implementation.
type Service struct {
	storage partners.Storage
	logger  *slog.Logger
}

// New builds a Service.
func New(storage partners.Storage) *Service {
	return &Service{
		storage: storage,
		logger:  slog.Default().With(slogx.Module("service:partner")),
	}
}

func (s *Service) Create(ctx context.Context, in *entities.PartnerCreate) (*entities.Partner, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	p := entities.PartnerNew()
	in.Merge(p)

	if err := s.storage.Insert(ctx, p); err != nil {
		s.logger.DebugContext(ctx, "insert partner failed", slogx.Error(err))
		return nil, err
	}
	return p, nil
}

func (s *Service) Get(ctx context.Context, id string) (*entities.Partner, error) {
	return s.storage.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, in *entities.PartnersList) (*entities.List[entities.Partner], error) {
	return s.storage.List(ctx, in)
}

func (s *Service) Update(ctx context.Context, in *entities.PartnerUpdate) (*entities.Partner, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	p, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if in.Etag != nil && *in.Etag != p.Etag {
		return nil, errs.ErrStaleEntity
	}
	oldEtag := p.Etag
	in.Merge(p)
	p.BeforeUpdate()
	if err := s.storage.Update(ctx, p, oldEtag); err != nil {
		s.logger.DebugContext(ctx, "update partner failed", slog.String("id", p.ID), slogx.Error(err))
		return nil, err
	}
	return p, nil
}

func (s *Service) Delete(ctx context.Context, in *entities.PartnerDelete) error {
	_ = normalizer.Normalize(in) //nolint:errcheck

	p, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		return err
	}
	if in.Etag != nil && *in.Etag != p.Etag {
		return errs.ErrStaleEntity
	}

	oldEtag := p.Etag
	now := time.Now().UTC()
	p.DeletedAt = &now
	p.BeforeUpdate()
	if err := s.storage.SoftDelete(ctx, &entities.SoftDelete{
		ID:           p.ID,
		Etag:         oldEtag,
		NewUpdatedAt: *p.DeletedAt,
		NewEtag:      p.Etag,
	}); err != nil {
		s.logger.DebugContext(ctx, "soft delete partner failed", slog.String("id", p.ID), slogx.Error(err))
		return err
	}
	return nil
}
