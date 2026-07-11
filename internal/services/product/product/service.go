// Package product implements the product.Service interface.
package product

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/altessa-s/go-atlas/domain/normalizer"
	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	brandsvc "github.com/kitdoo/my-business-crm-go/internal/services/brand"
	categorysvc "github.com/kitdoo/my-business-crm-go/internal/services/category"
	productsvc "github.com/kitdoo/my-business-crm-go/internal/services/product"
	"github.com/kitdoo/my-business-crm-go/internal/storages/products"
)

var _ productsvc.Service = (*Service)(nil)

// Service is the product.Service implementation. brands/categories are the
// respective entities' Service, not their Storage — see
// SERVICE_DEVELOPMENT_STANDARD.md's "A service controls only its own
// storage" rule. (The reverse checks, brand/category.Service.Delete asking
// whether a product still exists, stay Storage-shaped on the other side —
// wiring both directions through the Service layer would be a circular
// dependency: this service already needs brand/category to construct.)
type Service struct {
	storage        products.Storage
	brands         brandsvc.Service
	categories     categorysvc.Service
	requiredLocale string
	logger         *slog.Logger
}

// New builds a Service. requiredLocale is the locale every non-empty
// LocalizedString field must include (see
// entities.LocalizedString.Validate); it comes from
// CRMConfig.DefaultLocale.
func New(storage products.Storage, brands brandsvc.Service, categories categorysvc.Service, requiredLocale string) *Service {
	return &Service{
		storage:        storage,
		brands:         brands,
		categories:     categories,
		requiredLocale: requiredLocale,
		logger:         slog.Default().With(slogx.Module("service:product")),
	}
}

// checkBrand confirms brandID resolves to an existing brand, translating a
// not-found into errs.ErrProductBrandNotFound.
func (s *Service) checkBrand(ctx context.Context, brandID string) error {
	if _, err := s.brands.Get(ctx, brandID); err != nil {
		if errors.Is(err, errs.ErrBrandNotFound) {
			return errs.ErrProductBrandNotFound
		}
		return err
	}
	return nil
}

// checkCategories confirms every id in categoryIDs resolves to an existing
// category, translating a not-found into errs.ErrProductCategoryNotFound.
func (s *Service) checkCategories(ctx context.Context, categoryIDs []string) error {
	for _, id := range categoryIDs {
		if _, err := s.categories.Get(ctx, id); err != nil {
			if errors.Is(err, errs.ErrCategoryNotFound) {
				return errs.ErrProductCategoryNotFound
			}
			return err
		}
	}
	return nil
}

func (s *Service) Create(ctx context.Context, in *entities.ProductCreate) (*entities.Product, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	if err := in.Validate(s.requiredLocale); err != nil {
		return nil, err
	}

	if err := s.checkBrand(ctx, in.BrandID); err != nil {
		return nil, err
	}
	if err := s.checkCategories(ctx, in.CategoryIDs); err != nil {
		return nil, err
	}

	p := entities.ProductNew()
	in.Merge(p)

	if err := s.storage.Insert(ctx, p); err != nil {
		s.logger.DebugContext(ctx, "insert product failed", slog.String("sku", p.SKU), slogx.Error(err))
		return nil, err
	}
	return p, nil
}

func (s *Service) Get(ctx context.Context, id string) (*entities.Product, error) {
	return s.storage.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, in *entities.ProductsList) (*entities.List[entities.Product], error) {
	return s.storage.List(ctx, in)
}

func (s *Service) Update(ctx context.Context, in *entities.ProductUpdate) (*entities.Product, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	if err := in.Validate(s.requiredLocale); err != nil {
		return nil, err
	}

	p, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if in.Etag != nil && *in.Etag != p.Etag {
		return nil, errs.ErrStaleEntity
	}
	if in.BrandID != nil {
		if err := s.checkBrand(ctx, *in.BrandID); err != nil {
			return nil, err
		}
	}
	if in.CategoryIDs != nil {
		if err := s.checkCategories(ctx, in.CategoryIDs); err != nil {
			return nil, err
		}
	}

	oldEtag := p.Etag
	in.Merge(p)
	p.BeforeUpdate()
	if err := s.storage.Update(ctx, p, oldEtag); err != nil {
		s.logger.DebugContext(ctx, "update product failed", slog.String("id", p.ID), slogx.Error(err))
		return nil, err
	}
	return p, nil
}

func (s *Service) Delete(ctx context.Context, in *entities.ProductDelete) error {
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
		s.logger.DebugContext(ctx, "soft delete product failed", slog.String("id", p.ID), slogx.Error(err))
		return err
	}
	return nil
}
