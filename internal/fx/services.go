package fx

import (
	"time"

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
	invmovementsvc "github.com/kitdoo/my-business-crm-go/internal/services/inventorymovement"
	invmovementservice "github.com/kitdoo/my-business-crm-go/internal/services/inventorymovement/inventorymovement"
	"github.com/kitdoo/my-business-crm-go/internal/services/mailer"
	mailerservice "github.com/kitdoo/my-business-crm-go/internal/services/mailer/mailer"
	"github.com/kitdoo/my-business-crm-go/internal/services/notification"
	notificationservice "github.com/kitdoo/my-business-crm-go/internal/services/notification/notification"
	"github.com/kitdoo/my-business-crm-go/internal/services/partner"
	partnerservice "github.com/kitdoo/my-business-crm-go/internal/services/partner/partner"
	"github.com/kitdoo/my-business-crm-go/internal/services/price"
	priceservice "github.com/kitdoo/my-business-crm-go/internal/services/price/price"
	"github.com/kitdoo/my-business-crm-go/internal/services/product"
	productservice "github.com/kitdoo/my-business-crm-go/internal/services/product/product"
	productattributedefinitionsvc "github.com/kitdoo/my-business-crm-go/internal/services/productattributedefinition"
	productattributedefinitionservice "github.com/kitdoo/my-business-crm-go/internal/services/productattributedefinition/productattributedefinition"
	reportsvc "github.com/kitdoo/my-business-crm-go/internal/services/report"
	reportservice "github.com/kitdoo/my-business-crm-go/internal/services/report/report"
	salesvc "github.com/kitdoo/my-business-crm-go/internal/services/sale"
	saleservice "github.com/kitdoo/my-business-crm-go/internal/services/sale/sale"
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
	invmovements "github.com/kitdoo/my-business-crm-go/internal/storages/inventorymovements"
	invmovementsmongo "github.com/kitdoo/my-business-crm-go/internal/storages/inventorymovements/mongo"
	"github.com/kitdoo/my-business-crm-go/internal/storages/partners"
	partnersmongo "github.com/kitdoo/my-business-crm-go/internal/storages/partners/mongo"
	"github.com/kitdoo/my-business-crm-go/internal/storages/prices"
	pricesmongo "github.com/kitdoo/my-business-crm-go/internal/storages/prices/mongo"
	"github.com/kitdoo/my-business-crm-go/internal/storages/products"
	productsmongo "github.com/kitdoo/my-business-crm-go/internal/storages/products/mongo"
	productattributedefinitions "github.com/kitdoo/my-business-crm-go/internal/storages/productattributedefinitions"
	productattributedefinitionsmongo "github.com/kitdoo/my-business-crm-go/internal/storages/productattributedefinitions/mongo"
	"github.com/kitdoo/my-business-crm-go/internal/storages/reports"
	reportsmongo "github.com/kitdoo/my-business-crm-go/internal/storages/reports/mongo"
	"github.com/kitdoo/my-business-crm-go/internal/storages/sales"
	salesmongo "github.com/kitdoo/my-business-crm-go/internal/storages/sales/mongo"
	"github.com/kitdoo/my-business-crm-go/internal/storages/users"
	usersmongo "github.com/kitdoo/my-business-crm-go/internal/storages/users/mongo"
	"github.com/kitdoo/my-business-crm-go/internal/storages/warehouses"
	warehousesmongo "github.com/kitdoo/my-business-crm-go/internal/storages/warehouses/mongo"
	brandhandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/brand"
	categoryhandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/category"
	clienthandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/client"
	invhandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/inventory"
	invmovementhandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/inventorymovement"
	notificationhandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/notification"
	partnerhandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/partner"
	pricehandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/price"
	producthandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/product"
	productattributedefinitionhandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/productattributedefinition"
	reporthandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/report"
	salehandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/sale"
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

		fx.Provide(fx.Annotate(invmovementsmongo.New, fx.As(new(invmovements.Storage)))),
		fx.Provide(fx.Annotate(invmovementservice.New, fx.As(new(invmovementsvc.Service)))),
		fx.Provide(AsGRPCHandler(invmovementhandler.New)),

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
		fx.Provide(fx.Annotate(newUserService, fx.As(new(user.Service)))),
		fx.Provide(AsGRPCHandler(userhandler.New)),
		fx.Invoke(registerBootstrapAdmin),
		imagesModule(),

		fx.Provide(fx.Annotate(productsmongo.New, fx.As(new(products.Storage)))),
		fx.Provide(fx.Annotate(newProductService, fx.As(new(product.Service)))),
		fx.Provide(AsGRPCHandler(producthandler.New)),

		fx.Provide(fx.Annotate(productattributedefinitionsmongo.New, fx.As(new(productattributedefinitions.Storage)))),
		fx.Provide(fx.Annotate(productattributedefinitionservice.New, fx.As(new(productattributedefinitionsvc.Service)))),
		fx.Provide(AsGRPCHandler(productattributedefinitionhandler.New)),

		fx.Provide(fx.Annotate(pricesmongo.New, fx.As(new(prices.Storage)))),
		fx.Provide(fx.Annotate(newPriceService, fx.As(new(price.Service)))),
		fx.Provide(AsGRPCHandler(pricehandler.New)),

		fx.Provide(fx.Annotate(salesmongo.New, fx.As(new(sales.Storage)))),
		fx.Provide(fx.Annotate(saleservice.New, fx.As(new(salesvc.Service)))),
		fx.Provide(AsGRPCHandler(salehandler.New)),

		fx.Provide(fx.Annotate(reportsmongo.New, fx.As(new(reports.Storage)))),
		fx.Provide(fx.Annotate(reportservice.New, fx.As(new(reportsvc.Service)))),
		fx.Provide(AsGRPCHandler(reporthandler.New)),

		fx.Provide(fx.Annotate(newMailerService, fx.As(new(mailer.Service)))),
		fx.Provide(fx.Annotate(newNotificationService, fx.As(new(notification.Service)))),
		fx.Provide(AsGRPCHandler(notificationhandler.New)),
	)
}

// defaultLocale is used when the crm config section (or its defaultLocale
// field) is absent, matching CRMConfig.DefaultLocale's own default tag.
const defaultLocale = "sr"

// resolveDefaultLocale resolves the locale every non-empty LocalizedString
// field must include (see entities.LocalizedString.Validate) from
// cfg.CRM.DefaultLocale, falling back to defaultLocale when the optional
// crm config section (or just this field) is absent. Shared by every
// service that validates a LocalizedString field.
func resolveDefaultLocale(cfg *appconfig.Config) string {
	if cfg.CRM != nil && cfg.CRM.DefaultLocale != "" {
		return cfg.CRM.DefaultLocale
	}
	return defaultLocale
}

// newBrandService wires brand.Service with products.Storage as its
// ProductsExistenceChecker. This stays Storage-shaped rather than
// depending on product.Service — see brand.ProductsExistenceChecker's doc
// for why (product.Service already depends on brand.Service, so the
// reverse direction through the Service layer would be circular).
func newBrandService(storage brands.Storage, productsStorage products.Storage, cfg *appconfig.Config) *brandservice.Service {
	return brandservice.New(storage, productsStorage, resolveDefaultLocale(cfg))
}

// newCategoryService wires category.Service with products.Storage as its
// ProductsExistenceChecker; see newBrandService for the rationale.
func newCategoryService(storage categories.Storage, productsStorage products.Storage, cfg *appconfig.Config) *categoryservice.Service {
	return categoryservice.New(storage, productsStorage, resolveDefaultLocale(cfg))
}

// newUserService resolves the session-token signing key/TTL from
// cfg.CRM.Auth. There is no fallback for the signing key — userservice.New
// fails construction (and so app boot) if it is empty, rather than starting
// with tokens nobody can verify or, worse, a predictable default key.
func newUserService(storage users.Storage, cfg *appconfig.Config) (*userservice.Service, error) {
	var signingKey string
	var tokenTTL time.Duration
	if cfg.CRM != nil && cfg.CRM.Auth != nil {
		signingKey = cfg.CRM.Auth.SigningKey.Expose()
		tokenTTL = cfg.CRM.Auth.TokenTTL
	}
	return userservice.New(storage, []byte(signingKey), tokenTTL, resolveDefaultLocale(cfg))
}

// defaultCurrency is used when the crm config section (or its currency
// field) is absent, matching the TD's own example ("RSD").
const defaultCurrency = "RSD"

// newPriceService resolves the system-wide currency from cfg.CRM.Currency,
// falling back to defaultCurrency when the optional crm config section (or
// just its currency field) is absent.
func newPriceService(storage prices.Storage, productsSvc product.Service, cfg *appconfig.Config) *priceservice.Service {
	currency := defaultCurrency
	if cfg.CRM != nil && cfg.CRM.Currency != "" {
		currency = cfg.CRM.Currency
	}
	return priceservice.New(storage, productsSvc, currency)
}

// newWarehouseService wires warehouse.Service with inventory.Service as
// its stock-guard dependency for Delete/Deactivate.
func newWarehouseService(storage warehouses.Storage, inventorySvc invsvc.Service, cfg *appconfig.Config) *warehouseservice.Service {
	return warehouseservice.New(storage, inventorySvc, resolveDefaultLocale(cfg))
}

// newProductService wires product.Service with brand.Service/
// category.Service for FK validation (Create/Update), and resolves the
// required LocalizedString locale from cfg.CRM.DefaultLocale.
func newProductService(storage products.Storage, brandSvc brand.Service, categorySvc category.Service, cfg *appconfig.Config) *productservice.Service {
	return productservice.New(storage, brandSvc, categorySvc, resolveDefaultLocale(cfg))
}

// newMailerService resolves the SMTP connection/auth data from
// cfg.CRM.Smtp. A nil section (optional — see CRMConfig.Smtp) yields a
// Service whose Send always fails with errs.ErrSMTPNotConfigured, rather
// than failing app boot.
func newMailerService(cfg *appconfig.Config) *mailerservice.Service {
	if cfg.CRM == nil || cfg.CRM.Smtp == nil {
		return mailerservice.New(nil)
	}
	smtp := cfg.CRM.Smtp
	return mailerservice.New(&mailerservice.Config{
		Host:     smtp.Host,
		Port:     smtp.Port,
		Username: smtp.Username,
		Password: smtp.Password.Expose(),
		From:     smtp.From,
	})
}

// newNotificationService resolves the message recipient from
// cfg.CRM.Smtp.To.
func newNotificationService(mailerSvc mailer.Service, cfg *appconfig.Config) *notificationservice.Service {
	var recipient string
	if cfg.CRM != nil && cfg.CRM.Smtp != nil {
		recipient = cfg.CRM.Smtp.To
	}
	return notificationservice.New(mailerSvc, recipient)
}
