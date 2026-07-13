// Package brand implements the brand.Service interface.
package brand

import (
	"context"
	"log/slog"
	"time"

	"github.com/altessa-s/go-atlas/domain/normalizer"
	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	brandsvc "github.com/kitdoo/my-business-crm-go/internal/services/brand"
	"github.com/kitdoo/my-business-crm-go/internal/storages/brands"
)

var _ brandsvc.Service = (*Service)(nil)

// Service is the brand.Service implementation.
type Service struct {
	storage        brands.Storage
	products       brandsvc.ProductsExistenceChecker // optional; nil until the products aggregate exists
	requiredLocale string
	logger         *slog.Logger
}

// New builds a Service. products may be nil (see ProductsExistenceChecker).
// requiredLocale is the locale every non-empty LocalizedString field must
// include (see entities.LocalizedString.Validate); it comes from
// CRMConfig.DefaultLocale.
func New(storage brands.Storage, products brandsvc.ProductsExistenceChecker, requiredLocale string) *Service {
	return &Service{
		storage:        storage,
		products:       products,
		requiredLocale: requiredLocale,
		logger:         slog.Default().With(slogx.Module("service:brand")),
	}
}

func (s *Service) Create(ctx context.Context, in *entities.BrandCreate) (*entities.Brand, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	if err := in.Validate(s.requiredLocale); err != nil {
		return nil, err
	}

	b := entities.BrandNew()
	in.Merge(b)

	if err := s.storage.Insert(ctx, b); err != nil {
		s.logger.DebugContext(ctx, "insert brand failed", slogx.Error(err))
		return nil, err
	}
	return b, nil
}

func (s *Service) Get(ctx context.Context, id string) (*entities.Brand, error) {
	b, err := s.storage.Get(ctx, id)
	if err != nil {
		s.logger.DebugContext(ctx, "get brand failed", slog.String("id", id), slogx.Error(err))
		return nil, err
	}
	return b, nil
}

func (s *Service) List(ctx context.Context, in *entities.BrandsList) (*entities.List[entities.Brand], error) {
	list, err := s.storage.List(ctx, in)
	if err != nil {
		s.logger.DebugContext(ctx, "list brands failed", slogx.Error(err))
		return nil, err
	}
	return list, nil
}

func (s *Service) Update(ctx context.Context, in *entities.BrandUpdate) (*entities.Brand, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	if err := in.Validate(s.requiredLocale); err != nil {
		return nil, err
	}

	b, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		s.logger.DebugContext(ctx, "get brand failed", slog.String("id", in.ID), slogx.Error(err))
		return nil, err
	}
	if in.Etag != nil && *in.Etag != b.Etag {
		return nil, errs.ErrStaleEntity
	}
	oldEtag := b.Etag
	in.Merge(b)
	b.BeforeUpdate()
	if err := s.storage.Update(ctx, b, oldEtag); err != nil {
		s.logger.DebugContext(ctx, "update brand failed", slog.String("id", b.ID), slogx.Error(err))
		return nil, err
	}
	return b, nil
}

func (s *Service) Delete(ctx context.Context, in *entities.BrandDelete) error {
	_ = normalizer.Normalize(in) //nolint:errcheck

	b, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		s.logger.DebugContext(ctx, "get brand failed", slog.String("id", in.ID), slogx.Error(err))
		return err
	}
	if in.Etag != nil && *in.Etag != b.Etag {
		return errs.ErrStaleEntity
	}

	if s.products != nil {
		hasProducts, err := s.products.ExistsForBrand(ctx, b.ID)
		if err != nil {
			s.logger.DebugContext(ctx, "check brand products existence failed", slog.String("id", b.ID), slogx.Error(err))
			return err
		}
		if hasProducts {
			return errs.ErrBrandHasProducts
		}
	}

	oldEtag := b.Etag
	now := time.Now().UTC()
	b.DeletedAt = &now
	b.BeforeUpdate()
	if err := s.storage.SoftDelete(ctx, &entities.SoftDelete{
		ID:           b.ID,
		Etag:         oldEtag,
		NewUpdatedAt: *b.DeletedAt,
		NewEtag:      b.Etag,
	}); err != nil {
		s.logger.DebugContext(ctx, "soft delete brand failed", slog.String("id", b.ID), slogx.Error(err))
		return err
	}
	return nil
}
