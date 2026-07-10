package fx

import (
	"go.uber.org/fx"

	"github.com/kitdoo/my-business-crm-go/internal/services/brand"
	brandservice "github.com/kitdoo/my-business-crm-go/internal/services/brand/brand"
	"github.com/kitdoo/my-business-crm-go/internal/services/category"
	categoryservice "github.com/kitdoo/my-business-crm-go/internal/services/category/category"
	"github.com/kitdoo/my-business-crm-go/internal/services/client"
	clientservice "github.com/kitdoo/my-business-crm-go/internal/services/client/client"
	"github.com/kitdoo/my-business-crm-go/internal/services/partner"
	partnerservice "github.com/kitdoo/my-business-crm-go/internal/services/partner/partner"
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
	"github.com/kitdoo/my-business-crm-go/internal/storages/partners"
	partnersmongo "github.com/kitdoo/my-business-crm-go/internal/storages/partners/mongo"
	"github.com/kitdoo/my-business-crm-go/internal/storages/users"
	usersmongo "github.com/kitdoo/my-business-crm-go/internal/storages/users/mongo"
	"github.com/kitdoo/my-business-crm-go/internal/storages/warehouses"
	warehousesmongo "github.com/kitdoo/my-business-crm-go/internal/storages/warehouses/mongo"
	brandhandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/brand"
	categoryhandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/category"
	clienthandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/client"
	partnerhandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/partner"
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
	)
}

// newBrandService wires brand.Service without a ProductsExistenceChecker:
// the products aggregate does not exist yet in this codebase, so there is no
// implementation to provide here. Service.Delete treats a nil checker as
// "no products aggregate to guard against" and skips the check; wire a real
// checker once the products aggregate lands.
func newBrandService(storage brands.Storage) *brandservice.Service {
	return brandservice.New(storage, nil)
}

// newCategoryService wires category.Service without a
// ProductsExistenceChecker; see newBrandService for the rationale.
func newCategoryService(storage categories.Storage) *categoryservice.Service {
	return categoryservice.New(storage, nil)
}

// newWarehouseService wires warehouse.Service without an InventoryChecker:
// the inventory aggregate does not exist yet in this codebase, so there is
// no implementation to provide here. Service treats a nil checker as "no
// inventory aggregate to guard against" and skips the stock guard; wire a
// real checker once the inventory aggregate lands.
func newWarehouseService(storage warehouses.Storage) *warehouseservice.Service {
	return warehouseservice.New(storage, nil)
}
