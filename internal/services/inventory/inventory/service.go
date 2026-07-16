// Package inventory implements the inventory.Service interface.
package inventory

import (
	"context"
	"log/slog"

	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	inventorysvc "github.com/kitdoo/my-business-crm-go/internal/services/inventory"
	productsvc "github.com/kitdoo/my-business-crm-go/internal/services/product"
	productskusvc "github.com/kitdoo/my-business-crm-go/internal/services/productsku"
	productvariantsvc "github.com/kitdoo/my-business-crm-go/internal/services/productvariant"
	"github.com/kitdoo/my-business-crm-go/internal/storages/inventory"
)

var _ inventorysvc.Service = (*Service)(nil)

// Service is the inventory.Service implementation. skus/variants/products
// are the respective entities' Service, not their Storage — see
// SERVICE_DEVELOPMENT_STANDARD.md's "A service controls only its own
// storage" rule. None of the three depend back on inventory.Service (the
// dependency graph is inventory -> productsku -> productvariant ->
// product), so this stays a plain constructor injection with no
// checker-interface indirection.
type Service struct {
	storage  inventory.Storage
	skus     productskusvc.Service
	variants productvariantsvc.Service
	products productsvc.Service
	logger   *slog.Logger
}

// New builds a Service.
func New(storage inventory.Storage, skus productskusvc.Service, variants productvariantsvc.Service, products productsvc.Service) *Service {
	return &Service{
		storage:  storage,
		skus:     skus,
		variants: variants,
		products: products,
		logger:   slog.Default().With(slogx.Module("service:inventory")),
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
	// Best-effort: HasStock is a derived cache field (see
	// entities.Product.HasStock), not the source of truth for the
	// quantity itself, so a failure here is logged loudly rather than
	// rolling back the movement that already succeeded above — matching
	// inventorymovement.Service.Create's ledger-insert-after-adjust
	// precedent.
	s.recomputeProductHasStock(ctx, skuID)
	return i, nil
}

// recomputeProductHasStock walks skuID up to its owning Product (via its
// ProductVariant) and recomputes whether that product has any active
// variant with an active, in-stock SKU, writing the result via
// product.Service.SetHasStock. Every step is logged and swallowed on
// failure — see ApplyMovement's doc for why this never surfaces to the
// caller.
func (s *Service) recomputeProductHasStock(ctx context.Context, skuID string) {
	sku, err := s.skus.Get(ctx, skuID)
	if err != nil {
		s.logger.ErrorContext(ctx, "recompute product has_stock: get sku failed", slog.String("skuID", skuID), slog.String("error", err.Error()))
		return
	}
	variant, err := s.variants.Get(ctx, sku.VariantID)
	if err != nil {
		s.logger.ErrorContext(ctx, "recompute product has_stock: get variant failed", slog.String("variantID", sku.VariantID), slog.String("error", err.Error()))
		return
	}

	hasStock, err := s.productHasStock(ctx, variant.ProductID)
	if err != nil {
		s.logger.ErrorContext(ctx, "recompute product has_stock: compute failed", slog.String("productID", variant.ProductID), slog.String("error", err.Error()))
		return
	}
	if err := s.products.SetHasStock(ctx, variant.ProductID, hasStock); err != nil {
		s.logger.ErrorContext(ctx, "recompute product has_stock: set failed", slog.String("productID", variant.ProductID), slog.String("error", err.Error()))
	}
}

// productHasStock reports whether productID has any active
// ProductVariant with an active ProductSKU carrying positive Inventory
// quantity in any warehouse — the same definition the public catalog BFF
// uses per-SKU (see web-public's isSkuInStock), rolled up to product
// level.
func (s *Service) productHasStock(ctx context.Context, productID string) (bool, error) {
	variants, err := s.variants.List(ctx, &entities.ProductVariantsList{
		ProductIDs: []string{productID},
		Statuses:   []entities.ProductVariantStatus{entities.ProductVariantStatusActive},
		Pagination: entities.ListPagination{Limit: 200},
	})
	if err != nil {
		return false, err
	}
	if len(variants.Items) == 0 {
		return false, nil
	}

	variantIDs := make([]string, len(variants.Items))
	for i, v := range variants.Items {
		variantIDs[i] = v.ID
	}
	skus, err := s.skus.List(ctx, &entities.ProductSKUsList{
		VariantIDs: variantIDs,
		Statuses:   []entities.ProductSkuStatus{entities.ProductSkuStatusActive},
		Pagination: entities.ListPagination{Limit: 200},
	})
	if err != nil {
		return false, err
	}

	for _, sku := range skus.Items {
		items, err := s.storage.List(ctx, &entities.InventoryList{
			SKUID:      &sku.ID,
			Pagination: entities.ListPagination{Limit: 200},
		})
		if err != nil {
			return false, err
		}
		for _, inv := range items.Items {
			if inv.Quantity > 0 {
				return true, nil
			}
		}
	}
	return false, nil
}

func (s *Service) HasStock(ctx context.Context, warehouseID string) (bool, error) {
	has, err := s.storage.HasStock(ctx, warehouseID)
	if err != nil {
		s.logger.DebugContext(ctx, "check warehouse stock failed", slog.String("warehouseID", warehouseID), slog.String("error", err.Error()))
		return false, err
	}
	return has, nil
}
