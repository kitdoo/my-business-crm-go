package run

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/altessa-s/go-atlas/core/runtime/appinfo"
	"github.com/altessa-s/go-atlas/observability/health"
	"github.com/altessa-s/go-atlas/service/id"

	"github.com/kitdoo/my-business-crm-go/internal/pkg/appconfig"

	coreerrs "github.com/altessa-s/go-atlas/core/errors"
	slogx "github.com/altessa-s/go-atlas/observability/slog"
	loggerfactory "github.com/altessa-s/go-atlas/observability/slog/factory"
	fxmodules "github.com/kitdoo/my-business-crm-go/internal/fx"
)

const (
	// startTimeout bounds the time we allow Fx to construct and start all dependencies.
	startTimeout = 30 * time.Second
	// shutdownTimeout bounds the time we allow Fx to gracefully stop dependencies.
	shutdownTimeout = 30 * time.Second
)

// App holds the runtime state needed to bootstrap and run the service.
// It owns configuration loading, Uber Fx dependency injection setup, and
// signal-driven graceful shutdown.
//
// Create with [NewApp], then call [App.Run] exactly once.
type App struct {
	configPath string
	config     *appconfig.Config

	healthCoordinator *health.Coordinator

	serviceID *id.Service
}

// NewApp allocates an [App] with a generated service ID persisted to disk.
// Call [App.Run] to start.
func NewApp() *App {
	return &App{
		serviceID: id.MustNew(
			path.Join(appinfo.LibDir(), strings.ToLower(appinfo.Name)+".sid"),
		),
	}
}

// Run resolves the config path from the --config flag or the CONFIG_FILE
// environment variable, loads configuration, boots all Fx modules, and
// blocks until SIGINT or SIGTERM is received.
func (srv *App) Run(cmd *cobra.Command, _ []string) error {
	configFile, _ := cmd.Parent().PersistentFlags().GetString("config") //nolint:errcheck // flag registered as String
	if cf := appinfo.GetEnvVar("CONFIG_FILE"); cf != "" {
		configFile = cf
	}

	if configFile != "" {
		if _, err := os.Stat(configFile); err != nil {
			return fmt.Errorf("configuration path does not exist: %s", configFile)
		}
		srv.configPath = configFile
	}

	return srv.run(context.Background())
}

// run performs the actual server startup.
func (srv *App) run(ctx context.Context) error {
	if err := srv.loadConfig(); err != nil {
		return coreerrs.WrapOperationWithContext(err, "load config", srv.configPath)
	}

	if srv.config.Node != nil && srv.config.Node.Id != nil {
		srv.serviceID = id.MustNewStaticProvider(*srv.config.Node.Id)
	}

	logger, err := loggerfactory.New(srv.config.Logger).
		WithServiceId(srv.serviceID.ID()).
		WithAppVersion(appinfo.Version).
		WithEnableMasking().
		Build()
	if err != nil {
		return coreerrs.WrapOperation(err, "create logger")
	}
	slog.SetDefault(logger)

	slog.Default().Info("starting service...",
		slog.String("service", appinfo.Name),
		slog.String("version", appinfo.Version),
		slog.String("sid", srv.serviceID.ID()),
		slog.Group("dirs",
			"bin", appinfo.BinDir(),
			"etc", appinfo.EtcDir(),
			"lib", appinfo.LibDir(),
			"var", appinfo.VarDir(),
		),
	)

	if err := appinfo.MakeAllDirs(); err != nil {
		return coreerrs.WrapOperation(err, "create service directories")
	}

	startCtx, startCancel := context.WithTimeout(ctx, startTimeout)
	defer startCancel()

	di := fx.New(
		fx.NopLogger,
		fx.RecoverFromPanics(),
		fx.Supply(srv.config),
		fx.Supply(srv.serviceID),
		fx.Supply(slog.Default()),

		fxmodules.InfrastructureModule(),
		fxmodules.TransportsModule(),

		fx.Populate(&srv.healthCoordinator),
	)

	if di.Err() != nil {
		return coreerrs.WrapOperation(di.Err(), "initialize dependencies")
	}

	if err := di.Start(startCtx); err != nil {
		return coreerrs.WrapOperation(err, "start dependencies")
	}

	return srv.gracefulShutdown(ctx, di)
}

// gracefulShutdown handles signal-based shutdown with proper cleanup.
func (srv *App) gracefulShutdown(ctx context.Context, di *fx.App) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		slog.Default().Info("received shutdown signal", slog.String("signal", sig.String()))
	case <-di.Wait():
		slog.Default().Info("application stopped unexpectedly")
	}

	// Flip readiness probes off before notifying push-side watchers so a probe
	// arriving in this narrow window sees NotServing instead of healthy.
	srv.healthCoordinator.SetOverallStatus(health.StatusNotServing)
	srv.healthCoordinator.BroadcastStatus(health.StatusNotServing)

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeout)
	defer shutdownCancel()

	slog.Default().Info("initiating graceful shutdown...")

	if err := di.Stop(shutdownCtx); err != nil {
		slog.Default().Error("error during graceful shutdown", slogx.Error(err))
		return coreerrs.WrapOperation(err, "graceful shutdown")
	}

	slog.Default().Info("service shutdown completed successfully")
	return nil
}

func (srv *App) loadConfig() error {
	var err error
	srv.config, err = appconfig.Load(srv.configPath)
	return err
}
