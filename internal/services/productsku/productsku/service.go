// Package productsku implements the productsku.Service interface.
package productsku

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/altessa-s/go-atlas/domain/normalizer"
	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	skusvc "github.com/kitdoo/my-business-crm-go/internal/services/productsku"
	variantsvc "github.com/kitdoo/my-business-crm-go/internal/services/productvariant"
	"github.com/kitdoo/my-business-crm-go/internal/storages/productskus"
)

var _ skusvc.Service = (*Service)(nil)

// Service is the productsku.Service implementation. variants is
// productvariant.Service, not productvariants.Storage — see
// SERVICE_DEVELOPMENT_STANDARD.md's "A service controls only its own
// storage" rule.
type Service struct {
	storage  productskus.Storage
	variants variantsvc.Service
	logger   *slog.Logger
}

// New builds a Service.
func New(storage productskus.Storage, variants variantsvc.Service) *Service {
	return &Service{
		storage:  storage,
		variants: variants,
		logger:   slog.Default().With(slogx.Module("service:productsku")),
	}
}

// checkVariant confirms variantID resolves to an existing variant,
// translating a not-found into errs.ErrProductSkuVariantNotFound.
func (s *Service) checkVariant(ctx context.Context, variantID string) error {
	if _, err := s.variants.Get(ctx, variantID); err != nil {
		if errors.Is(err, errs.ErrProductVariantNotFound) {
			return errs.ErrProductSkuVariantNotFound
		}
		return err
	}
	return nil
}

func (s *Service) Create(ctx context.Context, in *entities.ProductSKUCreate) (*entities.ProductSKU, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	if err := in.Validate(); err != nil {
		return nil, err
	}

	if err := s.checkVariant(ctx, in.VariantID); err != nil {
		return nil, err
	}

	v := entities.ProductSKUNew()
	in.Merge(v)

	if err := s.storage.Insert(ctx, v); err != nil {
		s.logger.DebugContext(ctx, "insert product sku failed", slog.String("id", v.ID), slog.String("error", err.Error()))
		return nil, err
	}
	return v, nil
}

func (s *Service) Get(ctx context.Context, id string) (*entities.ProductSKU, error) {
	v, err := s.storage.Get(ctx, id)
	if err != nil {
		s.logger.DebugContext(ctx, "get product sku failed", slog.String("id", id), slog.String("error", err.Error()))
		return nil, err
	}
	return v, nil
}

func (s *Service) List(ctx context.Context, in *entities.ProductSKUsList) (*entities.List[entities.ProductSKU], error) {
	list, err := s.storage.List(ctx, in)
	if err != nil {
		s.logger.DebugContext(ctx, "list product skus failed", slog.String("error", err.Error()))
		return nil, err
	}
	return list, nil
}

func (s *Service) Update(ctx context.Context, in *entities.ProductSKUUpdate) (*entities.ProductSKU, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	v, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		s.logger.DebugContext(ctx, "get product sku failed", slog.String("id", in.ID), slog.String("error", err.Error()))
		return nil, err
	}
	if in.Etag != nil && *in.Etag != v.Etag {
		return nil, errs.ErrStaleEntity
	}

	oldEtag := v.Etag
	in.Merge(v)
	v.BeforeUpdate()
	if err := s.storage.Update(ctx, v, oldEtag); err != nil {
		s.logger.DebugContext(ctx, "update product sku failed", slog.String("id", v.ID), slog.String("error", err.Error()))
		return nil, err
	}
	return v, nil
}

func (s *Service) Delete(ctx context.Context, in *entities.ProductSKUDelete) error {
	_ = normalizer.Normalize(in) //nolint:errcheck

	v, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		s.logger.DebugContext(ctx, "get product sku failed", slog.String("id", in.ID), slog.String("error", err.Error()))
		return err
	}
	if in.Etag != nil && *in.Etag != v.Etag {
		return errs.ErrStaleEntity
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
		s.logger.DebugContext(ctx, "soft delete product sku failed", slog.String("id", v.ID), slog.String("error", err.Error()))
		return err
	}
	return nil
}
