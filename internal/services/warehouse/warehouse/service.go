// Package warehouse implements the warehouse.Service interface.
package warehouse

import (
	"context"
	"log/slog"
	"time"

	"github.com/altessa-s/go-atlas/domain/normalizer"
	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	inventorysvc "github.com/kitdoo/my-business-crm-go/internal/services/inventory"
	warehousesvc "github.com/kitdoo/my-business-crm-go/internal/services/warehouse"
	"github.com/kitdoo/my-business-crm-go/internal/storages/warehouses"
)

var _ warehousesvc.Service = (*Service)(nil)

// Service is the warehouse.Service implementation. inventory is
// inventory.Service, not inventory.Storage — see
// SERVICE_DEVELOPMENT_STANDARD.md's "A service controls only its own
// storage" rule.
type Service struct {
	storage        warehouses.Storage
	inventory      inventorysvc.Service // nil is treated as "no inventory aggregate to check against"
	requiredLocale string
	logger         *slog.Logger
}

// New builds a Service. inventory may be nil, in which case the
// stock-guard on Delete/Deactivate is skipped. requiredLocale is the
// locale every non-empty LocalizedString field must include (see
// entities.LocalizedString.Validate); it comes from
// CRMConfig.DefaultLocale.
func New(storage warehouses.Storage, inventory inventorysvc.Service, requiredLocale string) *Service {
	return &Service{
		storage:        storage,
		inventory:      inventory,
		requiredLocale: requiredLocale,
		logger:         slog.Default().With(slogx.Module("service:warehouse")),
	}
}

func (s *Service) Create(ctx context.Context, in *entities.WarehouseCreate) (*entities.Warehouse, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	if err := in.Validate(s.requiredLocale); err != nil {
		return nil, err
	}

	w := entities.WarehouseNew()
	in.Merge(w)

	if err := s.storage.Insert(ctx, w); err != nil {
		s.logger.DebugContext(ctx, "insert warehouse failed", slogx.Error(err))
		return nil, err
	}
	return w, nil
}

func (s *Service) Get(ctx context.Context, id string) (*entities.Warehouse, error) {
	return s.storage.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, in *entities.WarehousesList) (*entities.List[entities.Warehouse], error) {
	return s.storage.List(ctx, in)
}

func (s *Service) Update(ctx context.Context, in *entities.WarehouseUpdate) (*entities.Warehouse, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	if err := in.Validate(s.requiredLocale); err != nil {
		return nil, err
	}

	w, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if in.Etag != nil && *in.Etag != w.Etag {
		return nil, errs.ErrStaleEntity
	}
	oldEtag := w.Etag
	in.Merge(w)
	w.BeforeUpdate()
	if err := s.storage.Update(ctx, w, oldEtag); err != nil {
		s.logger.DebugContext(ctx, "update warehouse failed", slog.String("id", w.ID), slogx.Error(err))
		return nil, err
	}
	return w, nil
}

func (s *Service) Delete(ctx context.Context, in *entities.WarehouseDelete) error {
	_ = normalizer.Normalize(in) //nolint:errcheck

	w, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		return err
	}
	if in.Etag != nil && *in.Etag != w.Etag {
		return errs.ErrStaleEntity
	}

	if err := s.checkNoStock(ctx, w.ID); err != nil {
		return err
	}

	oldEtag := w.Etag
	now := time.Now().UTC()
	w.DeletedAt = &now
	w.BeforeUpdate()
	if err := s.storage.SoftDelete(ctx, &entities.SoftDelete{
		ID:           w.ID,
		Etag:         oldEtag,
		NewUpdatedAt: *w.DeletedAt,
		NewEtag:      w.Etag,
	}); err != nil {
		s.logger.DebugContext(ctx, "soft delete warehouse failed", slog.String("id", w.ID), slogx.Error(err))
		return err
	}
	return nil
}

func (s *Service) Deactivate(ctx context.Context, in *entities.WarehouseDeactivate) (*entities.Warehouse, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	w, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if in.Etag != nil && *in.Etag != w.Etag {
		return nil, errs.ErrStaleEntity
	}

	if err := s.checkNoStock(ctx, w.ID); err != nil {
		return nil, err
	}

	oldEtag := w.Etag
	w.Status = entities.WarehouseStatusInactive
	w.BeforeUpdate()
	if err := s.storage.Update(ctx, w, oldEtag); err != nil {
		s.logger.DebugContext(ctx, "deactivate warehouse failed", slog.String("id", w.ID), slogx.Error(err))
		return nil, err
	}
	return w, nil
}

// checkNoStock blocks the caller (Delete or Deactivate) while the warehouse
// still carries inventory. The caller must transfer stock to another
// warehouse first (per the CRM TD); this only guards, it never transfers.
func (s *Service) checkNoStock(ctx context.Context, warehouseID string) error {
	if s.inventory == nil {
		return nil
	}
	hasStock, err := s.inventory.HasStock(ctx, warehouseID)
	if err != nil {
		return err
	}
	if hasStock {
		return errs.ErrWarehouseHasStock
	}
	return nil
}
