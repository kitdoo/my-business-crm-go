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

func (s *Service) Get(ctx context.Context, productID, warehouseID string) (*entities.Inventory, error) {
	i, err := s.storage.Get(ctx, productID, warehouseID)
	if err != nil {
		s.logger.DebugContext(ctx, "get inventory failed", slog.String("productID", productID), slog.String("warehouseID", warehouseID), slogx.Error(err))
		return nil, err
	}
	return i, nil
}

func (s *Service) List(ctx context.Context, in *entities.InventoryList) (*entities.List[entities.Inventory], error) {
	list, err := s.storage.List(ctx, in)
	if err != nil {
		s.logger.DebugContext(ctx, "list inventory failed", slogx.Error(err))
		return nil, err
	}
	return list, nil
}

func (s *Service) ApplyMovement(ctx context.Context, productID, warehouseID string, delta int64) (*entities.Inventory, error) {
	i, err := s.storage.ApplyMovement(ctx, productID, warehouseID, delta)
	if err != nil {
		s.logger.DebugContext(ctx, "apply inventory movement failed", slog.String("productID", productID), slog.String("warehouseID", warehouseID), slog.Int64("delta", delta), slogx.Error(err))
		return nil, err
	}
	return i, nil
}

func (s *Service) HasStock(ctx context.Context, warehouseID string) (bool, error) {
	has, err := s.storage.HasStock(ctx, warehouseID)
	if err != nil {
		s.logger.DebugContext(ctx, "check warehouse stock failed", slog.String("warehouseID", warehouseID), slogx.Error(err))
		return false, err
	}
	return has, nil
}
