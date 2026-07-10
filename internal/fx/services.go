package fx

import (
	"go.uber.org/fx"

	"github.com/kitdoo/my-business-crm-go/internal/services/brand"
	brandservice "github.com/kitdoo/my-business-crm-go/internal/services/brand/brand"
	"github.com/kitdoo/my-business-crm-go/internal/storages/brands"
	brandsmongo "github.com/kitdoo/my-business-crm-go/internal/storages/brands/mongo"
	brandhandler "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/handlers/brand"
)

// ServicesModule wires the domain aggregates: storage, service, and gRPC
// handler for each. Brand is the only aggregate implemented so far.
func ServicesModule() fx.Option {
	return fx.Module("services",
		fx.Provide(fx.Annotate(brandsmongo.New, fx.As(new(brands.Storage)))),
		fx.Provide(fx.Annotate(newBrandService, fx.As(new(brand.Service)))),
		fx.Provide(AsGRPCHandler(brandhandler.New)),
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
