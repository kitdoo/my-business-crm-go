package fx

import (
	"context"
	"errors"
	"log/slog"

	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"

	"github.com/altessa-s/go-atlas/config"
	"github.com/altessa-s/go-atlas/data/cache"
	cacheredis "github.com/altessa-s/go-atlas/data/cache/providers/redis"
	"github.com/altessa-s/go-atlas/data/idempotency"
	"github.com/altessa-s/go-atlas/data/limiters"
	"github.com/altessa-s/go-atlas/observability/health"
	"github.com/altessa-s/go-atlas/observability/metrics"
	"github.com/altessa-s/go-atlas/observability/tracing"

	"github.com/kitdoo/my-business-crm-go/internal/pkg/appconfig"

	coreerrs "github.com/altessa-s/go-atlas/core/errors"
	idempotencyfactory "github.com/altessa-s/go-atlas/data/idempotency/factory"
	limiterfactory "github.com/altessa-s/go-atlas/data/limiters/tokenbucket/factory"
	datamongo "github.com/altessa-s/go-atlas/data/mongo"
	mongofactory "github.com/altessa-s/go-atlas/infrastructure/mongo/factory"
	redisfactory "github.com/altessa-s/go-atlas/infrastructure/redis/factory"
	metricsfactory "github.com/altessa-s/go-atlas/observability/metrics/factory"
	slogx "github.com/altessa-s/go-atlas/observability/slog"
	tracingfactory "github.com/altessa-s/go-atlas/observability/tracing/factory"
	tlsfactory "github.com/altessa-s/go-atlas/security/tlsutils/factory"
	tlsproviders "github.com/altessa-s/go-atlas/security/tlsutils/providers"
)

// InfrastructureModule provides core infrastructure components: MongoDB,
// Redis, TLS, metrics, tracing, health, rate limiting, and idempotency.
//
// Constructors that depend on an optional configuration section return
// (nil, nil) when the section is absent — downstream consumers MUST nil-check
// before using these values.
func InfrastructureModule() fx.Option {
	return fx.Module("infrastructure",
		fx.Provide(newMongo),
		fx.Provide(newRedis),
		fx.Provide(newRedisCache),
		fx.Provide(newTLSProviders),
		fx.Provide(newMetricsCollector),
		fx.Provide(newTracer),
		fx.Provide(newHealthCoordinator),
		fx.Provide(fx.Annotate(newLimiter, fx.As(new(limiters.Limiter)))),
		fx.Provide(fx.Annotate(newIdempotency, fx.As(new(idempotency.Idempotency)))),
	)
}

func newMongo(cfg *appconfig.Config, collector metrics.Collector, lc fx.Lifecycle) (*datamongo.Mongo, error) {
	initCtx, cancel := initContext()
	defer cancel()

	m, err := mongofactory.New(cfg.MongoDB).
		UseLogger(slog.Default().With(slogx.Module("mongodb"))).
		UseCollector(collector).
		Build(initCtx)
	if err != nil {
		return nil, coreerrs.WrapOperation(err, "create MongoDB client")
	}

	if err = m.Connect(initCtx); err != nil {
		return nil, coreerrs.WrapOperation(err, "connect to MongoDB")
	}

	lc.Append(fx.StopHook(m.Close))
	return m, nil
}

func newRedis(cfg *appconfig.Config, hc *health.Coordinator, lc fx.Lifecycle) (redis.UniversalClient, error) {
	if cfg.Redis == nil {
		return nil, nil //nolint:nilnil // Redis is optional
	}

	initCtx, cancel := initContext()
	defer cancel()

	client, err := redisfactory.New(cfg.Redis).
		UseLogger(slog.Default().With(slogx.Module("redis"))).
		UseHealthCoordinator(hc).
		Build(initCtx)
	if err != nil {
		return nil, coreerrs.WrapOperation(err, "create Redis client")
	}

	lc.Append(fx.StopHook(client.Close))
	return client, nil
}

// newRedisCache adapts the shared Redis client into a go-atlas cache.Cache
// for cache-aside reads (currently only Inventory.Get uses it). Falls back
// to a no-op cache when Redis is not configured, so callers never nil-check.
func newRedisCache(client redis.UniversalClient) *cache.Cache {
	if client == nil {
		return cache.NewNoop()
	}
	return cache.New(cacheredis.New(client))
}

func newTLSProviders(cfg *appconfig.Config, lc fx.Lifecycle) (*tlsproviders.Providers, error) {
	if !cfg.IsTlsProviderConfigured() {
		return nil, nil //nolint:nilnil // TLS provider is optional
	}

	s, err := tlsfactory.New(cfg.TlsProvider).
		UseLogger(slog.Default().With(slogx.Module("tls"))).
		Build()
	if err != nil {
		return nil, coreerrs.WrapOperation(err, "create tls providers")
	}

	lc.Append(fx.StopHook(func(ctx context.Context) {
		s.Close(ctx, nil)
	}))
	return s, nil
}

func newMetricsCollector(cfg *appconfig.Config, lc fx.Lifecycle) (metrics.Collector, error) {
	var metricsCfg *config.Metrics
	if cfg.Observability != nil {
		metricsCfg = cfg.Observability.Metrics
	}

	collector, err := metricsfactory.New(metricsCfg).Build()
	if err != nil {
		return nil, coreerrs.WrapOperation(err, "create metrics collector")
	}

	lc.Append(fx.StopHook(collector.Shutdown))
	return collector, nil
}

func newTracer(cfg *appconfig.Config, lc fx.Lifecycle) (tracing.Tracer, error) {
	var tracingCfg *config.Tracing
	if cfg.Observability != nil {
		tracingCfg = cfg.Observability.Tracing
	}

	initCtx, cancel := initContext()
	defer cancel()

	tracer, err := tracingfactory.New(tracingCfg).Build(initCtx)
	if err != nil {
		return nil, coreerrs.WrapOperation(err, "create tracer")
	}

	lc.Append(fx.StopHook(tracer.Shutdown))
	return tracer, nil
}

func newHealthCoordinator(collector metrics.Collector) *health.Coordinator {
	return health.New(health.WithCollector(collector))
}

func newLimiter(cfg *appconfig.Config, collector metrics.Collector) (limiters.Limiter, error) {
	if cfg.Limiter == nil {
		return nil, nil //nolint:nilnil // shared limiter is optional
	}

	l, err := limiterfactory.New(cfg.Limiter).
		UseLogger(slog.Default().With(slogx.Module("limiter"))).
		UseCollector(collector).
		Build()
	if err != nil {
		return nil, coreerrs.WrapOperation(err, "create limiter")
	}
	return l, nil
}

func newIdempotency(cfg *appconfig.Config, collector metrics.Collector) (idempotency.Idempotency, error) {
	if cfg.Idempotency == nil {
		return nil, nil //nolint:nilnil // shared idempotency is optional
	}

	k, err := idempotencyfactory.New(cfg.Idempotency).
		UseLogger(slog.Default().With(slogx.Module("idempotency"))).
		UseCollector(collector).
		Build()
	if err != nil {
		return nil, coreerrs.WrapOperation(err, "create idempotency keeper")
	}
	return k, nil
}

var errServiceNotReady = errors.New("service is not ready")

// healthAdapter adapts [health.Coordinator] to the gRPC server's health.Health interface.
type healthAdapter struct {
	coordinator *health.Coordinator
}

// Health reports the current readiness of the service.
func (a *healthAdapter) Health(ctx context.Context) error {
	if a.coordinator.CheckHealth(ctx) != health.StatusServing {
		return errServiceNotReady
	}
	return nil
}
