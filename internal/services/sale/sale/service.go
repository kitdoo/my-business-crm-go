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
	skusvc "github.com/kitdoo/my-business-crm-go/internal/services/productsku"
	salesvc "github.com/kitdoo/my-business-crm-go/internal/services/sale"
	warehousesvc "github.com/kitdoo/my-business-crm-go/internal/services/warehouse"
	"github.com/kitdoo/my-business-crm-go/internal/storages/sales"
)

var _ salesvc.Service = (*Service)(nil)

// Service is the sale.Service implementation. clients/warehouses/partners/
// skus/prices/inventory are the respective entities' Service, not
// their Storage — see SERVICE_DEVELOPMENT_STANDARD.md's "A service
// controls only its own storage" rule.
type Service struct {
	storage    sales.Storage
	clients    clientsvc.Service
	warehouses warehousesvc.Service
	partners   partnersvc.Service
	skus       skusvc.Service
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
	skus skusvc.Service,
	prices pricesvc.Service,
	inventory inventorysvc.Service,
	movements invmovementsvc.Service,
) *Service {
	return &Service{
		storage:    storage,
		clients:    clients,
		warehouses: warehouses,
		partners:   partners,
		skus:       skus,
		prices:     prices,
		inventory:  inventory,
		movements:  movements,
		logger:     slog.Default().With(slogx.Module("service:sale")),
	}
}

func (s *Service) Create(ctx context.Context, in *entities.SaleCreate) (*entities.Sale, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	clientID := in.ClientID
	hasClient := clientID != "" || in.NewClient != nil
	if !hasClient && in.PartnerID == nil {
		return nil, errs.ErrSaleMissingClientOrPartner
	}

	if in.NewClient != nil {
		// Find-or-create by email (TD §12.3) — the caller never has to
		// create a client as a separate step before this one.
		c, err := s.clients.FindOrCreateByEmail(ctx, in.NewClient)
		if err != nil {
			return nil, err
		}
		clientID = c.ID
	} else if clientID != "" {
		if _, err := s.clients.Get(ctx, clientID); err != nil {
			return nil, err
		}
	}

	var partnerDiscount *int32
	if in.PartnerID != nil {
		p, err := s.partners.Get(ctx, *in.PartnerID)
		if err != nil {
			return nil, err
		}
		// Partner buying directly (no client on this sale) — their
		// discount rate overrides whatever per-item discount was
		// requested, rather than trusting the caller to enter it
		// correctly. When a client is also present, the partner only
		// earns Partner.CommissionPercentage (tracked, not applied here)
		// and the item discount stays caller-controlled.
		if !hasClient {
			partnerDiscount = &p.DiscountPercentage
		}
	}

	items, total, err := s.buildItems(ctx, in.Items, partnerDiscount)
	if err != nil {
		return nil, err
	}

	sale := entities.SaleNew(func(sl *entities.Sale) {
		sl.ClientID = clientID
		sl.PartnerID = in.PartnerID
		sl.Items = items
		sl.TotalAmount = total
		sl.Status = entities.SaleStatusDraft
		sl.CreatedBy = in.CreatedBy
	})

	if err := s.storage.Insert(ctx, sale); err != nil {
		s.logger.DebugContext(ctx, "insert sale failed", slog.String("error", err.Error()))
		return nil, err
	}

	// Best-effort past this point: the Sale row already exists, so a
	// failure partway through decrementing stock for later items is logged
	// loudly rather than rolled back — no cross-collection transaction is
	// used here, consistent with InventoryMovement.Create.
	for _, item := range sale.Items {
		if _, err := s.movements.Create(ctx, &entities.InventoryMovementCreate{
			SKUID:       item.SKUID,
			WarehouseID: item.WarehouseID,
			Type:        entities.MovementTypeSale,
			Quantity:    -item.Quantity,
			Comment:     fmt.Sprintf("sale %s", sale.ID),
			SaleID:      &sale.ID,
			CreatedBy:   sale.CreatedBy,
		}); err != nil {
			s.logger.ErrorContext(ctx, "record sale movement failed; sale and prior items' stock already committed",
				slog.String("saleId", sale.ID), slog.String("skuId", item.SKUID), slog.String("error", err.Error()))
			return nil, err
		}
	}

	return sale, nil
}

// buildItems validates each item's warehouse and SKU existence and stock
// availability there, captures the SKU's current price, and computes
// per-line and total amounts (basis points). Each item may target a
// different warehouse, so warehouse lookups are cached per distinct id to
// avoid redundant Gets within one sale. forcedDiscount, when non-nil,
// overrides every item's requested DiscountPercentage — used when the
// sale's only counterparty is a discount-bearing partner.
func (s *Service) buildItems(ctx context.Context, in []entities.SaleCreateItem, forcedDiscount *int32) ([]entities.SaleItem, int64, error) {
	items := make([]entities.SaleItem, 0, len(in))
	warehouses := make(map[string]*entities.Warehouse, len(in))
	var total int64

	for _, req := range in {
		wh, ok := warehouses[req.WarehouseID]
		if !ok {
			var err error
			wh, err = s.warehouses.Get(ctx, req.WarehouseID)
			if err != nil {
				return nil, 0, err
			}
			warehouses[req.WarehouseID] = wh
		}
		if wh.Status != entities.WarehouseStatusActive {
			return nil, 0, errs.ErrSaleWarehouseInactive
		}

		if _, err := s.skus.Get(ctx, req.SKUID); err != nil {
			return nil, 0, err
		}

		price, err := s.prices.Get(ctx, req.SKUID)
		if err != nil {
			return nil, 0, err
		}

		stock, err := s.inventory.Get(ctx, req.SKUID, req.WarehouseID)
		if err != nil {
			if errors.Is(err, errs.ErrInventoryNotFound) {
				return nil, 0, errs.ErrInsufficientStock
			}
			return nil, 0, err
		}
		if stock.Quantity < req.Quantity {
			return nil, 0, errs.ErrInsufficientStock
		}

		discount := req.DiscountPercentage
		if forcedDiscount != nil {
			discount = *forcedDiscount
		}

		// req.Quantity is in hundredths of a unit (see SaleItem.Quantity),
		// hence the extra /100 versus a plain price*quantity*discount calc.
		line := price.PriceAmount * req.Quantity * int64(100-discount) / 100 / 100
		total += line

		items = append(items, entities.SaleItem{
			SKUID:              req.SKUID,
			Quantity:           req.Quantity,
			PriceAmount:        price.PriceAmount,
			DiscountPercentage: discount,
			WarehouseID:        req.WarehouseID,
		})
	}

	return items, total, nil
}

func (s *Service) Get(ctx context.Context, id string) (*entities.Sale, error) {
	sl, err := s.storage.Get(ctx, id)
	if err != nil {
		s.logger.DebugContext(ctx, "get sale failed", slog.String("id", id), slog.String("error", err.Error()))
		return nil, err
	}
	return sl, nil
}

func (s *Service) List(ctx context.Context, in *entities.SalesList) (*entities.List[entities.Sale], error) {
	list, err := s.storage.List(ctx, in)
	if err != nil {
		s.logger.DebugContext(ctx, "list sales failed", slog.String("error", err.Error()))
		return nil, err
	}
	return list, nil
}

func isTerminal(status entities.SaleStatus) bool {
	return status == entities.SaleStatusCancelled || status == entities.SaleStatusRefunded
}

func (s *Service) UpdateStatus(ctx context.Context, in *entities.SaleUpdateStatus) (*entities.Sale, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	sl, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		s.logger.DebugContext(ctx, "get sale failed", slog.String("id", in.ID), slog.String("error", err.Error()))
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
		s.logger.DebugContext(ctx, "update sale status failed", slog.String("id", sl.ID), slog.String("error", err.Error()))
		return nil, err
	}
	return sl, nil
}

func (s *Service) Cancel(ctx context.Context, in *entities.SaleCancel) (*entities.Sale, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	sl, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		s.logger.DebugContext(ctx, "get sale failed", slog.String("id", in.ID), slog.String("error", err.Error()))
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
			SKUID:       item.SKUID,
			WarehouseID: item.WarehouseID,
			Type:        entities.MovementTypeAdjustment,
			Quantity:    item.Quantity,
			Comment:     comment,
			SaleID:      &sl.ID,
			CreatedBy:   in.CreatedBy,
		}); err != nil {
			s.logger.ErrorContext(ctx, "restock movement on sale cancel failed",
				slog.String("saleId", sl.ID), slog.String("skuId", item.SKUID), slog.String("error", err.Error()))
			return nil, err
		}
	}

	oldEtag := sl.Etag
	sl.Status = entities.SaleStatusCancelled
	sl.BeforeUpdate()
	if err := s.storage.Update(ctx, sl, oldEtag); err != nil {
		s.logger.DebugContext(ctx, "cancel sale failed", slog.String("id", sl.ID), slog.String("error", err.Error()))
		return nil, err
	}
	return sl, nil
}
