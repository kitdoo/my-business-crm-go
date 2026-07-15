// Package price implements the price.Service interface.
package price

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/altessa-s/go-atlas/domain/normalizer"
	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	skusvc "github.com/kitdoo/my-business-crm-go/internal/services/productsku"

	pricesvc "github.com/kitdoo/my-business-crm-go/internal/services/price"
	"github.com/kitdoo/my-business-crm-go/internal/storages/prices"
)

var _ pricesvc.Service = (*Service)(nil)

// Service is the price.Service implementation. skus is productsku.Service,
// not productskus.Storage — see SERVICE_DEVELOPMENT_STANDARD.md's "A
// service controls only its own storage" rule.
type Service struct {
	storage prices.Storage
	skus    skusvc.Service
	// currency is the system-wide ISO 4217 code from config, stamped onto
	// every price created; see PROTO_DEVELOPMENT_STANDARD.md's currency
	// note on crm.types.price.ProductPrice.
	currency string
	logger   *slog.Logger
}

// New builds a Service.
func New(storage prices.Storage, skus skusvc.Service, currency string) *Service {
	return &Service{
		storage:  storage,
		skus:     skus,
		currency: currency,
		logger:   slog.Default().With(slogx.Module("service:price")),
	}
}

func (s *Service) Create(ctx context.Context, in *entities.ProductPriceCreate) (*entities.ProductPrice, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	if _, err := s.skus.Get(ctx, in.SKUID); err != nil {
		return nil, err
	}
	in.Currency = s.currency

	p := entities.ProductPriceNew()
	in.Merge(p)

	if err := s.storage.Insert(ctx, p); err != nil {
		s.logger.DebugContext(ctx, "insert product price failed", slog.String("skuId", p.SKUID), slog.String("error", err.Error()))
		return nil, err
	}
	return p, nil
}

func (s *Service) Get(ctx context.Context, skuID string) (*entities.ProductPrice, error) {
	p, err := s.storage.GetBySkuID(ctx, skuID)
	if err != nil {
		s.logger.DebugContext(ctx, "get product price failed", slog.String("skuId", skuID), slog.String("error", err.Error()))
		return nil, err
	}
	return p, nil
}

func (s *Service) Update(ctx context.Context, in *entities.ProductPriceUpdate) (*entities.ProductPrice, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	p, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		s.logger.DebugContext(ctx, "get product price failed", slog.String("id", in.ID), slog.String("error", err.Error()))
		return nil, err
	}
	if in.Etag != nil && *in.Etag != p.Etag {
		return nil, errs.ErrStaleEntity
	}

	snapshot := &entities.ProductPrice{
		ID:             uuid.NewString(),
		SKUID:          p.SKUID,
		PriceAmount:    p.PriceAmount,
		Currency:       p.Currency,
		DiscountAmount: p.DiscountAmount,
		CreatedAt:      p.CreatedAt,
	}
	if err := s.storage.AppendHistory(ctx, snapshot); err != nil {
		s.logger.DebugContext(ctx, "append product price history failed", slog.String("id", p.ID), slog.String("error", err.Error()))
		return nil, err
	}

	oldEtag := p.Etag
	in.Merge(p)
	p.BeforeUpdate()
	if err := s.storage.Update(ctx, p, oldEtag); err != nil {
		s.logger.DebugContext(ctx, "update product price failed", slog.String("id", p.ID), slog.String("error", err.Error()))
		return nil, err
	}
	return p, nil
}

func (s *Service) Delete(ctx context.Context, in *entities.ProductPriceDelete) error {
	_ = normalizer.Normalize(in) //nolint:errcheck

	p, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		s.logger.DebugContext(ctx, "get product price failed", slog.String("id", in.ID), slog.String("error", err.Error()))
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
		s.logger.DebugContext(ctx, "soft delete product price failed", slog.String("id", p.ID), slog.String("error", err.Error()))
		return err
	}
	return nil
}

func (s *Service) GetHistory(ctx context.Context, in *entities.ProductPriceGetHistory) (*entities.List[entities.ProductPrice], error) {
	list, err := s.storage.GetHistory(ctx, in)
	if err != nil {
		s.logger.DebugContext(ctx, "get product price history failed", slog.String("error", err.Error()))
		return nil, err
	}
	return list, nil
}
