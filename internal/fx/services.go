package fx

import (
	"go.uber.org/fx"

	"github.com/kitdoo/my-business-crm-go/internal/services/brand"
	brandservice "github.com/kitdoo/my-business-crm-go/internal/services/brand/brand"
	"github.com/kitdoo/my-business-crm-go/internal/services/category"
	categoryservice "github.com/kitdoo/my-business-crm-go/internal/services/category/category"
	"github.com/kitdoo/my-business-crm-go/internal/storages/brands"
	brandsmongo "github.com/kitdoo/my-business-crm-go/internal/storages/brands/mongo"
	"github.com/kitdoo/my-business-crm-go/internal/storages/categories"
	categoriesmongo "github.com/kitdoo/my-business-crm-go/internal/storages/categories/mongo"
	brandhandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/brand"
	categoryhandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/category"
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
