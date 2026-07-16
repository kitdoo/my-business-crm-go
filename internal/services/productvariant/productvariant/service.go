// Package productvariant implements the productvariant.Service interface.
package productvariant

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/altessa-s/go-atlas/domain/normalizer"
	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	productsvc "github.com/kitdoo/my-business-crm-go/internal/services/product"
	variantsvc "github.com/kitdoo/my-business-crm-go/internal/services/productvariant"
	"github.com/kitdoo/my-business-crm-go/internal/storages/productvariants"
)

var _ variantsvc.Service = (*Service)(nil)

// Service is the productvariant.Service implementation. products is
// product.Service, not products.Storage — see
// SERVICE_DEVELOPMENT_STANDARD.md's "A service controls only its own
// storage" rule. skus is variantsvc.SKUsExistenceChecker (satisfied by
// productskus.Storage directly, not productsku.Service) — see that
// interface's doc for why.
type Service struct {
	storage  productvariants.Storage
	products productsvc.Service
	skus     variantsvc.SKUsExistenceChecker
	logger   *slog.Logger
}

// New builds a Service.
func New(storage productvariants.Storage, products productsvc.Service, skus variantsvc.SKUsExistenceChecker) *Service {
	return &Service{
		storage:  storage,
		products: products,
		skus:     skus,
		logger:   slog.Default().With(slogx.Module("service:productvariant")),
	}
}

// checkProduct confirms productID resolves to an existing product,
// translating a not-found into errs.ErrProductVariantProductNotFound.
func (s *Service) checkProduct(ctx context.Context, productID string) error {
	if _, err := s.products.Get(ctx, productID); err != nil {
		if errors.Is(err, errs.ErrProductNotFound) {
			return errs.ErrProductVariantProductNotFound
		}
		return err
	}
	return nil
}

func (s *Service) Create(ctx context.Context, in *entities.ProductVariantCreate) (*entities.ProductVariant, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	if err := in.Validate(); err != nil {
		return nil, err
	}

	if err := s.checkProduct(ctx, in.ProductID); err != nil {
		return nil, err
	}

	v := entities.ProductVariantNew()
	in.Merge(v)

	if err := s.storage.Insert(ctx, v); err != nil {
		s.logger.DebugContext(ctx, "insert product variant failed", slog.String("id", v.ID), slog.String("error", err.Error()))
		return nil, err
	}
	return v, nil
}

func (s *Service) Get(ctx context.Context, id string) (*entities.ProductVariant, error) {
	v, err := s.storage.Get(ctx, id)
	if err != nil {
		s.logger.DebugContext(ctx, "get product variant failed", slog.String("id", id), slog.String("error", err.Error()))
		return nil, err
	}
	return v, nil
}

func (s *Service) List(ctx context.Context, in *entities.ProductVariantsList) (*entities.List[entities.ProductVariant], error) {
	list, err := s.storage.List(ctx, in)
	if err != nil {
		s.logger.DebugContext(ctx, "list product variants failed", slog.String("error", err.Error()))
		return nil, err
	}
	return list, nil
}

func (s *Service) Update(ctx context.Context, in *entities.ProductVariantUpdate) (*entities.ProductVariant, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	v, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		s.logger.DebugContext(ctx, "get product variant failed", slog.String("id", in.ID), slog.String("error", err.Error()))
		return nil, err
	}
	if in.Etag != nil && *in.Etag != v.Etag {
		return nil, errs.ErrStaleEntity
	}

	oldEtag := v.Etag
	in.Merge(v)
	v.BeforeUpdate()
	if err := s.storage.Update(ctx, v, oldEtag); err != nil {
		s.logger.DebugContext(ctx, "update product variant failed", slog.String("id", v.ID), slog.String("error", err.Error()))
		return nil, err
	}
	return v, nil
}

func (s *Service) Delete(ctx context.Context, in *entities.ProductVariantDelete) error {
	_ = normalizer.Normalize(in) //nolint:errcheck

	v, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		s.logger.DebugContext(ctx, "get product variant failed", slog.String("id", in.ID), slog.String("error", err.Error()))
		return err
	}
	if in.Etag != nil && *in.Etag != v.Etag {
		return errs.ErrStaleEntity
	}

	if s.skus != nil {
		if err := s.skus.DeactivateForVariant(ctx, v.ID); err != nil {
			s.logger.DebugContext(ctx, "deactivate product skus failed", slog.String("id", v.ID), slog.String("error", err.Error()))
			return err
		}
	}

	oldEtag := v.Etag
	now := time.Now().UTC()
	v.DeletedAt = &now
	v.BeforeUpdate()
	if err := s.storage.SoftDelete(ctx, &entities.SoftDelete{
		ID:           v.ID,
		Etag:         oldEtag,
		NewUpdatedAt: *v.DeletedAt,
		NewEtag:      v.Etag,
	}); err != nil {
		s.logger.DebugContext(ctx, "soft delete product variant failed", slog.String("id", v.ID), slog.String("error", err.Error()))
		return err
	}
	return nil
}
