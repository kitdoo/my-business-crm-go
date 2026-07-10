package fx

import (
	"go.uber.org/fx"

	"github.com/kitdoo/my-business-crm-go/internal/pkg/appconfig"
	"github.com/kitdoo/my-business-crm-go/internal/services/brand"
	brandservice "github.com/kitdoo/my-business-crm-go/internal/services/brand/brand"
	"github.com/kitdoo/my-business-crm-go/internal/services/category"
	categoryservice "github.com/kitdoo/my-business-crm-go/internal/services/category/category"
	"github.com/kitdoo/my-business-crm-go/internal/services/client"
	clientservice "github.com/kitdoo/my-business-crm-go/internal/services/client/client"
	invsvc "github.com/kitdoo/my-business-crm-go/internal/services/inventory"
	invservice "github.com/kitdoo/my-business-crm-go/internal/services/inventory/inventory"
	"github.com/kitdoo/my-business-crm-go/internal/services/partner"
	partnerservice "github.com/kitdoo/my-business-crm-go/internal/services/partner/partner"
	"github.com/kitdoo/my-business-crm-go/internal/services/price"
	priceservice "github.com/kitdoo/my-business-crm-go/internal/services/price/price"
	"github.com/kitdoo/my-business-crm-go/internal/services/product"
	productservice "github.com/kitdoo/my-business-crm-go/internal/services/product/product"
	"github.com/kitdoo/my-business-crm-go/internal/services/user"
	userservice "github.com/kitdoo/my-business-crm-go/internal/services/user/user"
	"github.com/kitdoo/my-business-crm-go/internal/services/warehouse"
	warehouseservice "github.com/kitdoo/my-business-crm-go/internal/services/warehouse/warehouse"
	"github.com/kitdoo/my-business-crm-go/internal/storages/brands"
	brandsmongo "github.com/kitdoo/my-business-crm-go/internal/storages/brands/mongo"
	"github.com/kitdoo/my-business-crm-go/internal/storages/categories"
	categoriesmongo "github.com/kitdoo/my-business-crm-go/internal/storages/categories/mongo"
	"github.com/kitdoo/my-business-crm-go/internal/storages/clients"
	clientsmongo "github.com/kitdoo/my-business-crm-go/internal/storages/clients/mongo"
	invstorage "github.com/kitdoo/my-business-crm-go/internal/storages/inventory"
	invmongo "github.com/kitdoo/my-business-crm-go/internal/storages/inventory/mongo"
	"github.com/kitdoo/my-business-crm-go/internal/storages/partners"
	partnersmongo "github.com/kitdoo/my-business-crm-go/internal/storages/partners/mongo"
	"github.com/kitdoo/my-business-crm-go/internal/storages/prices"
	pricesmongo "github.com/kitdoo/my-business-crm-go/internal/storages/prices/mongo"
	"github.com/kitdoo/my-business-crm-go/internal/storages/products"
	productsmongo "github.com/kitdoo/my-business-crm-go/internal/storages/products/mongo"
	"github.com/kitdoo/my-business-crm-go/internal/storages/users"
	usersmongo "github.com/kitdoo/my-business-crm-go/internal/storages/users/mongo"
	"github.com/kitdoo/my-business-crm-go/internal/storages/warehouses"
	warehousesmongo "github.com/kitdoo/my-business-crm-go/internal/storages/warehouses/mongo"
	brandhandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/brand"
	categoryhandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/category"
	clienthandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/client"
	invhandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/inventory"
	partnerhandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/partner"
	pricehandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/price"
	producthandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/product"
	userhandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/user"
	warehousehandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/warehouse"
)

// ServicesModule wires the domain aggregates: storage, service, and gRPC
// handler for each.
func ServicesModule() fx.Option {
	return fx.Module("services",
		fx.Provide(fx.Annotate(brandsmongo.New, fx.As(new(brands.Storage)))),
		fx.Provide(fx.Annotate(newBrandService, fx.As(new(brand.Service)))),
		fx.Provide(AsGRPCHandler(brandhandler.New)),

		fx.Provide(fx.Annotate(categoriesmongo.New, fx.As(new(categories.Storage)))),
		fx.Provide(fx.Annotate(newCategoryService, fx.As(new(category.Service)))),
		fx.Provide(AsGRPCHandler(categoryhandler.New)),

		fx.Provide(fx.Annotate(invmongo.New, fx.As(new(invstorage.Storage)))),
		fx.Provide(fx.Annotate(invservice.New, fx.As(new(invsvc.Service)))),
		fx.Provide(AsGRPCHandler(invhandler.New)),

		fx.Provide(fx.Annotate(warehousesmongo.New, fx.As(new(warehouses.Storage)))),
		fx.Provide(fx.Annotate(newWarehouseService, fx.As(new(warehouse.Service)))),
		fx.Provide(AsGRPCHandler(warehousehandler.New)),

		fx.Provide(fx.Annotate(partnersmongo.New, fx.As(new(partners.Storage)))),
		fx.Provide(fx.Annotate(partnerservice.New, fx.As(new(partner.Service)))),
		fx.Provide(AsGRPCHandler(partnerhandler.New)),

		fx.Provide(fx.Annotate(clientsmongo.New, fx.As(new(clients.Storage)))),
		fx.Provide(fx.Annotate(clientservice.New, fx.As(new(client.Service)))),
		fx.Provide(AsGRPCHandler(clienthandler.New)),

		fx.Provide(fx.Annotate(usersmongo.New, fx.As(new(users.Storage)))),
		fx.Provide(fx.Annotate(userservice.New, fx.As(new(user.Service)))),
		fx.Provide(AsGRPCHandler(userhandler.New)),
		fx.Invoke(registerBootstrapAdmin),

		fx.Provide(fx.Annotate(productsmongo.New, fx.As(new(products.Storage)))),
		fx.Provide(fx.Annotate(productservice.New, fx.As(new(product.Service)))),
		fx.Provide(AsGRPCHandler(producthandler.New)),

		fx.Provide(fx.Annotate(pricesmongo.New, fx.As(new(prices.Storage)))),
		fx.Provide(fx.Annotate(newPriceService, fx.As(new(price.Service)))),
		fx.Provide(AsGRPCHandler(pricehandler.New)),
	)
}

// newBrandService wires brand.Service with products.Storage as its
// ProductsExistenceChecker: products.Storage.ExistsForBrand satisfies that
// interface directly, so the Delete guard is real once a product exists.
func newBrandService(storage brands.Storage, productsStorage products.Storage) *brandservice.Service {
	return brandservice.New(storage, productsStorage)
}

// newCategoryService wires category.Service with products.Storage as its
// ProductsExistenceChecker; see newBrandService for the rationale.
func newCategoryService(storage categories.Storage, productsStorage products.Storage) *categoryservice.Service {
	return categoryservice.New(storage, productsStorage)
}

// defaultCurrency is used when the crm config section (or its currency
// field) is absent, matching the TD's own example ("RSD").
const defaultCurrency = "RSD"

// newPriceService resolves the system-wide currency from cfg.CRM.Currency,
// falling back to defaultCurrency when the optional crm config section (or
// just its currency field) is absent.
func newPriceService(storage prices.Storage, productsStorage products.Storage, cfg *appconfig.Config) *priceservice.Service {
	currency := defaultCurrency
	if cfg.CRM != nil && cfg.CRM.Currency != "" {
		currency = cfg.CRM.Currency
	}
	return priceservice.New(storage, productsStorage, currency)
}

// newWarehouseService wires warehouse.Service with invstorage.Storage as
// its InventoryChecker: invstorage.Storage.HasStock satisfies that
// interface directly, so the Delete/Deactivate guards are real now that
// Inventory exists (the nil placeholder used since Warehouse landed).
func newWarehouseService(storage warehouses.Storage, inventoryStorage invstorage.Storage) *warehouseservice.Service {
	return warehouseservice.New(storage, inventoryStorage)
}
