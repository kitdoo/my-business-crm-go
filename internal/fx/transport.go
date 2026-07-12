package fx

import (
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"

	"github.com/altessa-s/go-atlas/data/idempotency"
	"github.com/altessa-s/go-atlas/data/limiters"
	"github.com/altessa-s/go-atlas/observability/tracing"
	"github.com/altessa-s/go-atlas/transport/grpc/handlers/health"
	"github.com/altessa-s/go-atlas/transport/grpc/handlers/serviceinfo"
	"github.com/altessa-s/go-atlas/transport/grpc/interceptors"

	"github.com/kitdoo/my-business-crm-go/internal/pkg/appconfig"
	"github.com/kitdoo/my-business-crm-go/internal/rbac"
	grpcauth "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/interceptors/auth"
	grpcclientkey "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/interceptors/clientkey"
	grpcrbac "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/interceptors/rbac"

	notificationsvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/notification/v1"

	coreerrs "github.com/altessa-s/go-atlas/core/errors"
	obshealth "github.com/altessa-s/go-atlas/observability/health"
	slogx "github.com/altessa-s/go-atlas/observability/slog"
	tlsproviders "github.com/altessa-s/go-atlas/security/tlsutils/providers"
	atlasid "github.com/altessa-s/go-atlas/service/id"
	grpcserver "github.com/altessa-s/go-atlas/transport/grpc/server"
	grpcfactory "github.com/altessa-s/go-atlas/transport/grpc/server/factory"
	httpserver "github.com/altessa-s/go-atlas/transport/http/server"
	httpfactory "github.com/altessa-s/go-atlas/transport/http/server/factory"
	httphealth "github.com/altessa-s/go-atlas/transport/http/server/handlers/health"
	httpping "github.com/altessa-s/go-atlas/transport/http/server/handlers/ping"
	httpwriter "github.com/altessa-s/go-atlas/transport/http/server/writer"
)

// httpInternalPrefix mirrors the default internal prefix used by the
// go-atlas HTTP factory for probes/metrics/pprof.
const httpInternalPrefix = "/internal"

// TransportsModule wires the gRPC and HTTP servers and their handlers.
func TransportsModule() fx.Option {
	return fx.Module("transports",
		fx.Provide(newGRPCServer),
		fx.Provide(newHTTPServer),

		fx.Provide(AsGRPCHandler(newServiceInfoHandler)),
		fx.Provide(AsGRPCHandler(health.New)),

		fx.Provide(AsGRPCInterceptor(grpcauth.New)),
		fx.Provide(newRBACTable),
		fx.Provide(AsGRPCInterceptor(grpcrbac.New)),
		fx.Provide(AsGRPCInterceptor(newClientKeyInterceptor)),

		fx.Invoke(fx.Annotate(registerGRPCHandlers, fx.ParamTags(`group:"grpc-interceptors"`, `group:"grpc-handlers"`))),

		// Force HTTP server construction even when no other consumer depends on it.
		fx.Invoke(func(*httpserver.Server) {}),
	)
}

func registerGRPCHandlers(intrs []interceptors.ServerInterceptor, handlers []grpcserver.Handler, srv *grpcserver.Server) error {
	if err := srv.RegisterInterceptors(intrs...); err != nil {
		return coreerrs.WrapOperation(err, "register gRPC interceptors")
	}
	srv.RegisterHandlers(handlers...)
	return nil
}

// AsGRPCHandler annotates a constructor function to provide a gRPC handler
// into the `grpc-handlers` group consumed by [registerGRPCHandlers].
func AsGRPCHandler(f any) any {
	return fx.Annotate(f, fx.As(new(grpcserver.Handler)),
		fx.ResultTags(`group:"grpc-handlers"`))
}

// AsGRPCInterceptor annotates a constructor function to provide a gRPC
// server interceptor into the `grpc-interceptors` group consumed by
// [registerGRPCHandlers].
func AsGRPCInterceptor(f any) any {
	return fx.Annotate(f, fx.As(new(interceptors.ServerInterceptor)),
		fx.ResultTags(`group:"grpc-interceptors"`))
}

func newGRPCServer(
	cfg *appconfig.Config,
	providers *tlsproviders.Providers,
	limiter limiters.Limiter,
	idp idempotency.Idempotency,
	tracer tracing.Tracer,
	hc *obshealth.Coordinator,
	lc fx.Lifecycle,
) (*grpcserver.Server, error) {
	logger := slog.Default().With(slogx.Module("grpc"))

	srv, err := grpcfactory.New(cfg.Grpc).
		UseLogger(logger).
		UseTlsProviders(providers).
		UseLimiter(limiter).
		UseIdempotency(idp).
		UseTracer(tracer).
		UseHealthChecker(&healthAdapter{hc}).
		WithInterceptors().
		Build()
	if err != nil {
		return nil, coreerrs.WrapOperation(err, "create gRPC server")
	}

	lc.Append(fx.StartHook(srv.Start))
	lc.Append(fx.StopHook(srv.Shutdown))
	return srv, nil
}

func newHTTPServer(
	cfg *appconfig.Config,
	providers *tlsproviders.Providers,
	limiter limiters.Limiter,
	idp idempotency.Idempotency,
	tracer tracing.Tracer,
	hc *obshealth.Coordinator,
	lc fx.Lifecycle,
) (*httpserver.Server, error) {
	if cfg.Http == nil {
		// appconfig.Config.Validate requires Http on any config produced by
		// appconfig.Load; this guard only protects callers (tests, DI graph
		// validation) that build a *Config by hand without going through
		// Load/Validate.
		return nil, nil //nolint:nilnil
	}

	builder := httpfactory.New(cfg.Http).
		UseLogger(slog.Default().With(slogx.Module("http"))).
		UseTlsProviders(providers).
		UseLimiter(limiter).
		UseIdempotency(idp).
		UseTracer(tracer).
		WithMiddlewares().
		WithoutBuiltinHandlers()

	if cfg.Http.Pprof != nil && cfg.Http.Pprof.Enabled {
		builder = builder.WithPprof()
	}

	srv, err := builder.Build()
	if err != nil {
		return nil, coreerrs.WrapOperation(err, "create HTTP server")
	}

	registerInternalHandlers(srv, hc)

	lc.Append(fx.StartHook(srv.Start))
	lc.Append(fx.StopHook(srv.Shutdown))
	return srv, nil
}

// registerInternalHandlers wires /internal/{ping,healthz,readyz,metrics} for
// liveness, readiness, and Prometheus scraping.
func registerInternalHandlers(srv *httpserver.Server, hc *obshealth.Coordinator) {
	srv.Handle(httpInternalPrefix+"/ping", httpping.Handler).Methods(http.MethodGet)
	srv.Handle(httpInternalPrefix+"/healthz", httphealth.K8sHealtz).Methods(http.MethodGet)
	srv.Handle(httpInternalPrefix+"/readyz", httphealth.Readyz(hc)).Methods(http.MethodGet)
	srv.Handle(httpInternalPrefix+"/healthz/details", httphealth.Detailed(hc)).Methods(http.MethodGet)
	srv.Handle(httpInternalPrefix+"/metrics/", prometheusMetrics).Methods(http.MethodGet)
}

func prometheusMetrics(rw httpwriter.ReadWriter) {
	promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{DisableCompression: true},
	).ServeHTTP(rw.ResponseWriter(), rw.Request())
}

func newServiceInfoHandler(sid *atlasid.Service) *serviceinfo.Handler {
	return serviceinfo.New(serviceinfo.WithServiceID(sid.ID()))
}

// newRBACTable resolves the role -> permission grants from cfg.CRM.RBAC.
// cfg.CRM is required by appconfig.Config.Validate on any config produced
// by appconfig.Load; the nil check only protects a hand-built *Config
// that bypassed it (e.g. in tests). An absent/empty rbac section yields a
// nil Table, under which every non-wildcard role is denied every
// permission-gated method (see rbac.Table.Allowed) — fail closed.
func newRBACTable(cfg *appconfig.Config) rbac.Table {
	if cfg.CRM == nil {
		return nil
	}
	return rbac.Table(cfg.CRM.RBAC)
}

// newClientKeyInterceptor resolves the approved API keys from
// cfg.CRM.NotificationClients (inverted: key -> client name) and scopes
// the interceptor to NotificationsService.Send, the one RPC that is
// exempt from user auth (see grpcauth.New) but still must be restricted
// to approved frontends. An absent/empty section denies every caller —
// fail closed, same as an absent/empty RBAC table.
func newClientKeyInterceptor(cfg *appconfig.Config) interceptors.ServerInterceptor {
	var keys map[string]string
	if cfg.CRM != nil {
		keys = make(map[string]string, len(cfg.CRM.NotificationClients))
		for name, key := range cfg.CRM.NotificationClients {
			keys[key.Expose()] = name
		}
	}
	return grpcclientkey.New(keys, notificationsvcpb.NotificationsService_Send_FullMethodName)
}
