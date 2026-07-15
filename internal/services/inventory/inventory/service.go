// Package inventory implements the inventory.Service interface.
package inventory

import (
	"context"
	"log/slog"

	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	inventorysvc "github.com/kitdoo/my-business-crm-go/internal/services/inventory"
	"github.com/kitdoo/my-business-crm-go/internal/storages/inventory"
)

var _ inventorysvc.Service = (*Service)(nil)

// Service is the inventory.Service implementation.
type Service struct {
	storage inventory.Storage
	logger  *slog.Logger
}

// New builds a Service.
func New(storage inventory.Storage) *Service {
	return &Service{
		storage: storage,
		logger:  slog.Default().With(slogx.Module("service:inventory")),
	}
}

func (s *Service) Get(ctx context.Context, skuID, warehouseID string) (*entities.Inventory, error) {
	i, err := s.storage.Get(ctx, skuID, warehouseID)
	if err != nil {
		s.logger.DebugContext(ctx, "get inventory failed", slog.String("skuID", skuID), slog.String("warehouseID", warehouseID), slog.String("error", err.Error()))
		return nil, err
	}
	return i, nil
}

func (s *Service) List(ctx context.Context, in *entities.InventoryList) (*entities.List[entities.Inventory], error) {
	list, err := s.storage.List(ctx, in)
	if err != nil {
		s.logger.DebugContext(ctx, "list inventory failed", slog.String("error", err.Error()))
		return nil, err
	}
	return list, nil
}

func (s *Service) ApplyMovement(ctx context.Context, skuID, warehouseID string, delta int64) (*entities.Inventory, error) {
	i, err := s.storage.ApplyMovement(ctx, skuID, warehouseID, delta)
	if err != nil {
		s.logger.DebugContext(ctx, "apply inventory movement failed", slog.String("skuID", skuID), slog.String("warehouseID", warehouseID), slog.Int64("delta", delta), slog.String("error", err.Error()))
		return nil, err
	}
	return i, nil
}

func (s *Service) HasStock(ctx context.Context, warehouseID string) (bool, error) {
	has, err := s.storage.HasStock(ctx, warehouseID)
	if err != nil {
		s.logger.DebugContext(ctx, "check warehouse stock failed", slog.String("warehouseID", warehouseID), slog.String("error", err.Error()))
		return false, err
	}
	return has, nil
}
