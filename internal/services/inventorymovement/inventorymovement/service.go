// Package inventorymovement implements the inventorymovement.Service
// interface.
package inventorymovement

import (
	"context"
	"log/slog"

	"github.com/altessa-s/go-atlas/domain/normalizer"
	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	inventorysvc "github.com/kitdoo/my-business-crm-go/internal/services/inventory"
	invsvc "github.com/kitdoo/my-business-crm-go/internal/services/inventorymovement"
	productsvc "github.com/kitdoo/my-business-crm-go/internal/services/product"
	warehousesvc "github.com/kitdoo/my-business-crm-go/internal/services/warehouse"
	"github.com/kitdoo/my-business-crm-go/internal/storages/inventorymovements"
)

var _ invsvc.Service = (*Service)(nil)

// Service is the inventorymovement.Service implementation. inventory,
// products, and warehouses are the respective entities' Service, not
// their Storage — see SERVICE_DEVELOPMENT_STANDARD.md's "A service
// controls only its own storage" rule. inventory.Service.ApplyMovement in
// particular is the sanctioned write path into Inventory; every other
// stock change should go through recording a movement here first.
type Service struct {
	storage    inventorymovements.Storage
	inventory  inventorysvc.Service
	products   productsvc.Service
	warehouses warehousesvc.Service
	logger     *slog.Logger
}

// New builds a Service.
func New(storage inventorymovements.Storage, inv inventorysvc.Service, products productsvc.Service, warehouses warehousesvc.Service) *Service {
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
	list, err := s.storage.List(ctx, in)
	if err != nil {
		s.logger.DebugContext(ctx, "list inventory movements failed", slogx.Error(err))
		return nil, err
	}
	return list, nil
}

func (s *Service) GetHistory(ctx context.Context, in *entities.InventoryMovementGetHistory) (*entities.List[entities.InventoryMovement], error) {
	list, err := s.storage.GetHistory(ctx, in)
	if err != nil {
		s.logger.DebugContext(ctx, "get inventory movement history failed", slogx.Error(err))
		return nil, err
	}
	return list, nil
}
