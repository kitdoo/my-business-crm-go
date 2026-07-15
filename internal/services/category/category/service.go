// Package category implements the category.Service interface.
package category

import (
	"context"
	"log/slog"
	"time"

	"github.com/altessa-s/go-atlas/domain/normalizer"
	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	categorysvc "github.com/kitdoo/my-business-crm-go/internal/services/category"
	"github.com/kitdoo/my-business-crm-go/internal/storages/categories"
)

var _ categorysvc.Service = (*Service)(nil)

// Service is the category.Service implementation.
type Service struct {
	storage        categories.Storage
	products       categorysvc.ProductsExistenceChecker // optional; nil until the products aggregate exists
	requiredLocale string
	logger         *slog.Logger
}

// New builds a Service. products may be nil (see ProductsExistenceChecker).
// requiredLocale is the locale every non-empty LocalizedString field must
// include (see entities.LocalizedString.Validate); it comes from
// CRMConfig.DefaultLocale.
func New(storage categories.Storage, products categorysvc.ProductsExistenceChecker, requiredLocale string) *Service {
	return &Service{
		storage:        storage,
		products:       products,
		requiredLocale: requiredLocale,
		logger:         slog.Default().With(slogx.Module("service:category")),
	}
}

func (s *Service) Create(ctx context.Context, in *entities.CategoryCreate) (*entities.Category, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	if err := in.Validate(s.requiredLocale); err != nil {
		return nil, err
	}

	c := entities.CategoryNew()
	in.Merge(c)

	if err := s.storage.Insert(ctx, c); err != nil {
		s.logger.DebugContext(ctx, "insert category failed", slog.String("error", err.Error()))
		return nil, err
	}
	return c, nil
}

func (s *Service) Get(ctx context.Context, id string) (*entities.Category, error) {
	c, err := s.storage.Get(ctx, id)
	if err != nil {
		s.logger.DebugContext(ctx, "get category failed", slog.String("id", id), slog.String("error", err.Error()))
		return nil, err
	}
	return c, nil
}

func (s *Service) List(ctx context.Context, in *entities.CategoriesList) (*entities.List[entities.Category], error) {
	list, err := s.storage.List(ctx, in)
	if err != nil {
		s.logger.DebugContext(ctx, "list categories failed", slog.String("error", err.Error()))
		return nil, err
	}
	return list, nil
}

func (s *Service) Update(ctx context.Context, in *entities.CategoryUpdate) (*entities.Category, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	if err := in.Validate(s.requiredLocale); err != nil {
		return nil, err
	}

	c, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		s.logger.DebugContext(ctx, "get category failed", slog.String("id", in.ID), slog.String("error", err.Error()))
		return nil, err
	}
	if in.Etag != nil && *in.Etag != c.Etag {
		return nil, errs.ErrStaleEntity
	}
	oldEtag := c.Etag
	in.Merge(c)
	c.BeforeUpdate()
	if err := s.storage.Update(ctx, c, oldEtag); err != nil {
		s.logger.DebugContext(ctx, "update category failed", slog.String("id", c.ID), slog.String("error", err.Error()))
		return nil, err
	}
	return c, nil
}

func (s *Service) Delete(ctx context.Context, in *entities.CategoryDelete) error {
	_ = normalizer.Normalize(in) //nolint:errcheck

	c, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		s.logger.DebugContext(ctx, "get category failed", slog.String("id", in.ID), slog.String("error", err.Error()))
		return err
	}
	if in.Etag != nil && *in.Etag != c.Etag {
		return errs.ErrStaleEntity
	}

	if s.products != nil {
		hasProducts, err := s.products.ExistsForCategory(ctx, c.ID)
		if err != nil {
			s.logger.DebugContext(ctx, "check category products existence failed", slog.String("id", c.ID), slog.String("error", err.Error()))
			return err
		}
		if hasProducts {
			return errs.ErrCategoryHasProducts
		}
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
		s.logger.DebugContext(ctx, "soft delete category failed", slog.String("id", c.ID), slog.String("error", err.Error()))
		return err
	}
	return nil
}
