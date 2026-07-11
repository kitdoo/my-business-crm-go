// Package sale implements the sale.Service interface.
package sale

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/altessa-s/go-atlas/domain/normalizer"
	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	clientsvc "github.com/kitdoo/my-business-crm-go/internal/services/client"
	inventorysvc "github.com/kitdoo/my-business-crm-go/internal/services/inventory"
	invmovementsvc "github.com/kitdoo/my-business-crm-go/internal/services/inventorymovement"
	partnersvc "github.com/kitdoo/my-business-crm-go/internal/services/partner"
	pricesvc "github.com/kitdoo/my-business-crm-go/internal/services/price"
	productsvc "github.com/kitdoo/my-business-crm-go/internal/services/product"
	salesvc "github.com/kitdoo/my-business-crm-go/internal/services/sale"
	warehousesvc "github.com/kitdoo/my-business-crm-go/internal/services/warehouse"
	"github.com/kitdoo/my-business-crm-go/internal/storages/sales"
)

var _ salesvc.Service = (*Service)(nil)

// Service is the sale.Service implementation. clients/warehouses/partners/
// products/prices/inventory are the respective entities' Service, not
// their Storage — see SERVICE_DEVELOPMENT_STANDARD.md's "A service
// controls only its own storage" rule.
type Service struct {
	storage    sales.Storage
	clients    clientsvc.Service
	warehouses warehousesvc.Service
	partners   partnersvc.Service
	products   productsvc.Service
	prices     pricesvc.Service
	inventory  inventorysvc.Service
	movements  invmovementsvc.Service
	logger     *slog.Logger
}

// New builds a Service.
func New(
	storage sales.Storage,
	clients clientsvc.Service,
	warehouses warehousesvc.Service,
	partners partnersvc.Service,
	products productsvc.Service,
	prices pricesvc.Service,
	inventory inventorysvc.Service,
	movements invmovementsvc.Service,
) *Service {
	return &Service{
		storage:    storage,
		clients:    clients,
		warehouses: warehouses,
		partners:   partners,
		products:   products,
		prices:     prices,
		inventory:  inventory,
		movements:  movements,
		logger:     slog.Default().With(slogx.Module("service:sale")),
	}
}

func (s *Service) Create(ctx context.Context, in *entities.SaleCreate) (*entities.Sale, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	if _, err := s.clients.Get(ctx, in.ClientID); err != nil {
		return nil, err
	}
	wh, err := s.warehouses.Get(ctx, in.WarehouseID)
	if err != nil {
		return nil, err
	}
	if wh.Status != entities.WarehouseStatusActive {
		return nil, errs.ErrSaleWarehouseInactive
	}
	if in.PartnerID != nil {
		if _, err := s.partners.Get(ctx, *in.PartnerID); err != nil {
			return nil, err
		}
	}

	items, total, err := s.buildItems(ctx, in.WarehouseID, in.Items)
	if err != nil {
		return nil, err
	}

	sale := entities.SaleNew(func(sl *entities.Sale) {
		sl.ClientID = in.ClientID
		sl.WarehouseID = in.WarehouseID
		sl.PartnerID = in.PartnerID
		sl.Items = items
		sl.TotalAmount = total
		sl.Status = entities.SaleStatusDraft
		sl.CreatedBy = in.CreatedBy
	})

	if err := s.storage.Insert(ctx, sale); err != nil {
		s.logger.DebugContext(ctx, "insert sale failed", slogx.Error(err))
		return nil, err
	}

	// Best-effort past this point: the Sale row already exists, so a
	// failure partway through decrementing stock for later items is logged
	// loudly rather than rolled back — no cross-collection transaction is
	// used here, consistent with InventoryMovement.Create.
	for _, item := range sale.Items {
		if _, err := s.movements.Create(ctx, &entities.InventoryMovementCreate{
			ProductID:   item.ProductID,
			WarehouseID: sale.WarehouseID,
			Type:        entities.MovementTypeSale,
			Quantity:    -item.Quantity,
			Comment:     fmt.Sprintf("sale %s", sale.ID),
			CreatedBy:   sale.CreatedBy,
		}); err != nil {
			s.logger.ErrorContext(ctx, "record sale movement failed; sale and prior items' stock already committed",
				slog.String("saleId", sale.ID), slog.String("productId", item.ProductID), slogx.Error(err))
			return nil, err
		}
	}

	return sale, nil
}

// buildItems validates each product's existence and stock availability at
// warehouseID, captures its current price, and computes per-line and total
// amounts (basis points).
func (s *Service) buildItems(ctx context.Context, warehouseID string, in []entities.SaleCreateItem) ([]entities.SaleItem, int64, error) {
	items := make([]entities.SaleItem, 0, len(in))
	var total int64

	for _, req := range in {
		if _, err := s.products.Get(ctx, req.ProductID); err != nil {
			return nil, 0, err
		}

		price, err := s.prices.Get(ctx, req.ProductID)
		if err != nil {
			return nil, 0, err
		}

		stock, err := s.inventory.Get(ctx, req.ProductID, warehouseID)
		if err != nil {
			if errors.Is(err, errs.ErrInventoryNotFound) {
				return nil, 0, errs.ErrInsufficientStock
			}
			return nil, 0, err
		}
		if stock.Quantity < req.Quantity {
			return nil, 0, errs.ErrInsufficientStock
		}

		line := price.PriceAmount * req.Quantity * int64(100-req.DiscountPercentage) / 100
		total += line

		items = append(items, entities.SaleItem{
			ProductID:          req.ProductID,
			Quantity:           req.Quantity,
			PriceAmount:        price.PriceAmount,
			DiscountPercentage: req.DiscountPercentage,
		})
	}

	return items, total, nil
}

func (s *Service) Get(ctx context.Context, id string) (*entities.Sale, error) {
	return s.storage.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, in *entities.SalesList) (*entities.List[entities.Sale], error) {
	return s.storage.List(ctx, in)
}

func isTerminal(status entities.SaleStatus) bool {
	return status == entities.SaleStatusCancelled || status == entities.SaleStatusRefunded
}

func (s *Service) UpdateStatus(ctx context.Context, in *entities.SaleUpdateStatus) (*entities.Sale, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	sl, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if in.Etag != nil && *in.Etag != sl.Etag {
		return nil, errs.ErrStaleEntity
	}
	if isTerminal(sl.Status) {
		return nil, errs.ErrSaleTerminalStatus
	}

	oldEtag := sl.Etag
	sl.Status = in.Status
	sl.BeforeUpdate()
	if err := s.storage.Update(ctx, sl, oldEtag); err != nil {
		s.logger.DebugContext(ctx, "update sale status failed", slog.String("id", sl.ID), slogx.Error(err))
		return nil, err
	}
	return sl, nil
}

func (s *Service) Cancel(ctx context.Context, in *entities.SaleCancel) (*entities.Sale, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	sl, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if in.Etag != nil && *in.Etag != sl.Etag {
		return nil, errs.ErrStaleEntity
	}
	if isTerminal(sl.Status) {
		return nil, errs.ErrSaleTerminalStatus
	}

	comment := fmt.Sprintf("cancel sale %s", sl.ID)
	if in.Reason != nil {
		comment = fmt.Sprintf("%s: %s", comment, *in.Reason)
	}

	// Best-effort restock: see Create for why this isn't transactional.
	for _, item := range sl.Items {
		if _, err := s.movements.Create(ctx, &entities.InventoryMovementCreate{
			ProductID:   item.ProductID,
			WarehouseID: sl.WarehouseID,
			Type:        entities.MovementTypeAdjustment,
			Quantity:    item.Quantity,
			Comment:     comment,
			CreatedBy:   in.CreatedBy,
		}); err != nil {
			s.logger.ErrorContext(ctx, "restock movement on sale cancel failed",
				slog.String("saleId", sl.ID), slog.String("productId", item.ProductID), slogx.Error(err))
			return nil, err
		}
	}

	oldEtag := sl.Etag
	sl.Status = entities.SaleStatusCancelled
	sl.BeforeUpdate()
	if err := s.storage.Update(ctx, sl, oldEtag); err != nil {
		s.logger.DebugContext(ctx, "cancel sale failed", slog.String("id", sl.ID), slogx.Error(err))
		return nil, err
	}
	return sl, nil
}
