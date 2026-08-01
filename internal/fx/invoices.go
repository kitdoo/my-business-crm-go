package fx

import (
	"go.uber.org/fx"

	httpserver "github.com/altessa-s/go-atlas/transport/http/server"

	invoicehandler "github.com/kitdoo/my-business-crm-go/internal/transports/http/handlers/invoice"
)

// invoicesModule wires the invoice-PDF-generation endpoint (plain HTTP, not
// gRPC — see internal/transports/http/handlers/invoice; it has no
// dedicated entity or gRPC method, same reasoning as imagesModule).
func invoicesModule() fx.Option {
	return fx.Options(
		fx.Provide(invoicehandler.New),
		fx.Invoke(registerInvoiceHandler),
	)
}

func registerInvoiceHandler(srv *httpserver.Server, h *invoicehandler.Handler) {
	if srv == nil {
		return
	}
	srv.RegisterHandlers(h)
}
