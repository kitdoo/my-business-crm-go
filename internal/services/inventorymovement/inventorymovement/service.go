// Package inventorymovement implements the inventorymovement.Service
// interface.
package inventorymovement

import (
	"context"
	"log/slog"

	"github.com/altessa-s/go-atlas/domain/normalizer"
	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	invsvc "github.com/kitdoo/my-business-crm-go/internal/services/inventorymovement"
	"github.com/kitdoo/my-business-crm-go/internal/storages/inventory"
	"github.com/kitdoo/my-business-crm-go/internal/storages/inventorymovements"
	"github.com/kitdoo/my-business-crm-go/internal/storages/products"
	"github.com/kitdoo/my-business-crm-go/internal/storages/warehouses"
)

var _ invsvc.Service = (*Service)(nil)

// Service is the inventorymovement.Service implementation.
type Service struct {
	storage    inventorymovements.Storage
	inventory  inventory.Storage
	products   products.Storage
	warehouses warehouses.Storage
	logger     *slog.Logger
}

// New builds a Service.
func New(storage inventorymovements.Storage, inv inventory.Storage, products products.Storage, warehouses warehouses.Storage) *Service {
	return &Service{
		storage:    storage,
		inventory:  inv,
		products:   products,
		warehouses: warehouses,
		logger:     slog.Default().With(slogx.Module("service:inventorymovement")),
	}
}

func (s *Service) Create(ctx context.Context, in *entities.InventoryMovementCreate) (*entities.InventoryMovement, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	if _, err := s.products.Get(ctx, in.ProductID); err != nil {
		return nil, err
	}
	if _, err := s.warehouses.Get(ctx, in.WarehouseID); err != nil {
		return nil, err
	}

	if _, err := s.inventory.ApplyMovement(ctx, in.ProductID, in.WarehouseID, in.Quantity); err != nil {
		return nil, err
	}

	m := entities.InventoryMovementNew()
	in.Merge(m)

	if err := s.storage.Insert(ctx, m); err != nil {
		// Inventory has already been adjusted at this point; the ledger
		// entry is best-effort audit trail, not the source of truth for
		// the quantity itself, so this is logged loudly rather than
		// treated as a rollback candidate (no cross-collection
		// transaction is used here, matching the rest of this codebase).
		s.logger.ErrorContext(ctx, "insert inventory movement failed after stock was already adjusted",
			slog.String("productId", m.ProductID), slog.String("warehouseId", m.WarehouseID), slogx.Error(err))
		return nil, err
	}
	return m, nil
}

func (s *Service) List(ctx context.Context, in *entities.InventoryMovementsList) (*entities.List[entities.InventoryMovement], error) {
	return s.storage.List(ctx, in)
}

func (s *Service) GetHistory(ctx context.Context, in *entities.InventoryMovementGetHistory) (*entities.List[entities.InventoryMovement], error) {
	return s.storage.GetHistory(ctx, in)
}
