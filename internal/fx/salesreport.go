package fx

import (
	"go.uber.org/fx"

	httpserver "github.com/altessa-s/go-atlas/transport/http/server"

	salesreporthandler "github.com/kitdoo/my-business-crm-go/internal/transports/http/handlers/salesreport"
)

// salesReportModule wires the sales-report-generation endpoint (plain
// HTTP, not gRPC — see internal/transports/http/handlers/salesreport; it
// has no dedicated entity or gRPC method, same reasoning as
// invoicesModule).
func salesReportModule() fx.Option {
	return fx.Options(
		fx.Provide(salesreporthandler.New),
		fx.Invoke(registerSalesReportHandler),
	)
}

func registerSalesReportHandler(srv *httpserver.Server, h *salesreporthandler.Handler) {
	if srv == nil {
		return
	}
	srv.RegisterHandlers(h)
}
