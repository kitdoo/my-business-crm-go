package fx

import (
	"path/filepath"

	"go.uber.org/fx"

	"github.com/altessa-s/go-atlas/core/runtime/appinfo"
	httpserver "github.com/altessa-s/go-atlas/transport/http/server"

	"github.com/kitdoo/my-business-crm-go/internal/pkg/appconfig"
	"github.com/kitdoo/my-business-crm-go/internal/services/user"
	imagehandler "github.com/kitdoo/my-business-crm-go/internal/transports/http/handlers/image"
	"github.com/kitdoo/my-business-crm-go/internal/storages/images"
	imagesmongo "github.com/kitdoo/my-business-crm-go/internal/storages/images/mongo"
)

// imagesModule wires the admin-only product-image upload/serve endpoint
// (plain HTTP, not gRPC — see internal/transports/http/handlers/image; the
// TD gives images no dedicated entity or gRPC method), plus the Mongo-backed
// metadata store (content type/size) that upload writes and serve reads
// back, alongside the file bytes on disk.
func imagesModule() fx.Option {
	return fx.Options(
		fx.Provide(fx.Annotate(imagesmongo.New, fx.As(new(images.Storage)))),
		fx.Provide(newImageHandler),
		fx.Invoke(registerImageHandler),
	)
}

// newImageHandler resolves the upload directory/size limit from the
// optional crm.images config section, falling back to "<var-dir>/images"
// and imagehandler.DefaultMaxSizeBytes. cfg.CRM itself is required by
// appconfig.Config.Validate; the nil check only protects a hand-built
// *Config that bypassed Load/Validate (e.g. in tests).
func newImageHandler(cfg *appconfig.Config, users user.Service, imagesStorage images.Storage) (*imagehandler.Handler, error) {
	dir := filepath.Join(appinfo.VarDir(), "images")
	var maxSize int64
	if cfg.CRM != nil && cfg.CRM.Images != nil {
		if cfg.CRM.Images.Dir != "" {
			dir = cfg.CRM.Images.Dir
		}
		maxSize = cfg.CRM.Images.MaxSizeBytes
	}
	return imagehandler.New(dir, maxSize, users, imagesStorage)
}

// registerImageHandler mounts the image handler on the HTTP server. srv is
// required by appconfig.Config.Validate (image upload is a plain-HTTP-only,
// always-on feature); the nil check only protects a hand-built *Config
// that bypassed Load/Validate.
func registerImageHandler(srv *httpserver.Server, h *imagehandler.Handler) {
	if srv == nil {
		return
	}
	srv.RegisterHandlers(h)
}
