// Package inventory implements the inventory.Service interface.
package inventory

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	inventorysvc "github.com/kitdoo/my-business-crm-go/internal/services/inventory"
	"github.com/kitdoo/my-business-crm-go/internal/storages/inventory"
)

var _ inventorysvc.Service = (*Service)(nil)

// Service is the inventory.Service implementation.
type Service struct {
	storage inventory.Storage
}

// New builds a Service.
func New(storage inventory.Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) Get(ctx context.Context, productID, warehouseID string) (*entities.Inventory, error) {
	return s.storage.Get(ctx, productID, warehouseID)
}

func (s *Service) List(ctx context.Context, in *entities.InventoryList) (*entities.List[entities.Inventory], error) {
	return s.storage.List(ctx, in)
}
