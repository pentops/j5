package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pentops/grpc.go/grpcbind"
	"github.com/pentops/j5/internal/registry/anyfs"
	"github.com/pentops/j5/internal/registry/buildwrap"
	"github.com/pentops/j5/internal/registry/github"
	"github.com/pentops/j5/internal/registry/gomodproxy"
	"github.com/pentops/j5/internal/registry/packagestore"
	"github.com/pentops/j5/internal/registry/service"
	"github.com/pentops/log.go/log"
	"github.com/pentops/runner"
	"github.com/pentops/runner/commander"
	"github.com/pentops/sqrlx.go/pgenv"
	"github.com/pentops/sqrlx.go/sqrlx"
	"github.com/pressly/goose"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/pentops/o5-messaging/outbox"
)

var Version = "0.0.0"

func main() {
	cmdGroup := commander.NewCommandSet()

	cmdGroup.Add("serve", commander.NewCommand(runCombinedServer))
	cmdGroup.Add("readonly", commander.NewCommand(runReadonlyServer))
	cmdGroup.Add("migrate", commander.NewCommand(runMigrate))

	cmdGroup.RunMain("registry", Version)
}

func runMigrate(ctx context.Context, cfg struct {
	MigrationsDir string `env:"MIGRATIONS_DIR" default:"./ext/db"`
	DB            pgenv.DatabaseConfig
}) error {

	db, err := cfg.DB.OpenPostgres(ctx)
	if err != nil {
		return err
	}

	return goose.Up(db, cfg.MigrationsDir)
}

type HTTPConfig struct {
	HTTPBind string `env:"HTTP_BIND" default:":8081"`
}

func (srv *HTTPConfig) LisrtenAndServe(ctx context.Context, handler http.Handler) error {
	httpServer := &http.Server{
		Addr:    srv.HTTPBind,
		Handler: handler,
	}
	log.WithField(ctx, "bind", srv.HTTPBind).Info("Begin HTTP Server")

	go func() {
		<-ctx.Done()
		httpServer.Shutdown(ctx) // nolint:errcheck
	}()

	return httpServer.ListenAndServe()
}

func runReadonlyServer(ctx context.Context, cfg struct {
	HTTP         HTTPConfig
	GRPC         grpcbind.EnvConfig
	DB           pgenv.DatabaseConfig
	PackageStore PackageStoreConfig
}) error {
	dbConn, err := cfg.DB.OpenPostgres(ctx)
	if err != nil {
		return err
	}

	db := sqrlx.NewPostgres(dbConn)

	pkgStore, err := cfg.PackageStore.OpenPackageStore(ctx, db)
	if err != nil {
		return err
	}

	registryDownloadService := service.NewRegistryService(pkgStore)

	runGroup := runner.NewGroup(runner.WithName("main"), runner.WithCancelOnSignals())

	runGroup.Add("httpServer", func(ctx context.Context) error {
		handler := gomodproxy.Handler(pkgStore)
		return cfg.HTTP.LisrtenAndServe(ctx, handler)
	})

	runGroup.Add("grpcServer", func(ctx context.Context) error {
		grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
			service.GRPCMiddleware(Version)...,
		))

		registryDownloadService.RegisterGRPC(grpcServer)

		reflection.Register(grpcServer)

		return cfg.GRPC.ListenAndServe(ctx, grpcServer)
	})

	return runGroup.Run(ctx)
}

type PackageStoreConfig struct {
	Storage string `env:"REGISTRY_STORAGE"`
}

func (cfg PackageStoreConfig) OpenPackageStore(ctx context.Context, db sqrlx.Transactor) (*packagestore.PackageStore, error) {

	fs, err := anyfs.NewEnvFS(ctx, cfg.Storage)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	pkgStore, err := packagestore.NewPackageStore(db, fs)
	if err != nil {
		return nil, err
	}

	return pkgStore, nil
}

func runCombinedServer(ctx context.Context, cfg struct {
	HTTP         HTTPConfig
	GRPC         grpcbind.EnvConfig
	PackageStore PackageStoreConfig
	DB           pgenv.DatabaseConfig
}) error {
	db, err := cfg.DB.OpenPostgresTransactor(ctx)
	if err != nil {
		return err
	}

	pkgStore, err := cfg.PackageStore.OpenPackageStore(ctx, db)
	if err != nil {
		return err
	}

	githubClient, err := github.NewEnvClient(ctx)
	if err != nil {
		return err
	}

	dbPublisher, err := outbox.NewDirectPublisher(db, outbox.DefaultSender)
	if err != nil {
		return err
	}

	regWrap := buildwrap.NewRegistryClient(pkgStore)
	j5Builder, err := buildwrap.NewBuilder(regWrap)
	if err != nil {
		return err
	}

	buildWorker := buildwrap.NewBuildWorker(j5Builder, githubClient, pkgStore, dbPublisher)

	registryDownloadService := service.NewRegistryService(pkgStore)

	runGroup := runner.NewGroup(runner.WithName("main"), runner.WithCancelOnSignals())

	runGroup.Add("httpServer", func(ctx context.Context) error {
		handler := gomodproxy.Handler(pkgStore)
		return cfg.HTTP.LisrtenAndServe(ctx, handler)
	})

	runGroup.Add("grpcServer", func(ctx context.Context) error {
		grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
			service.GRPCMiddleware(Version)...,
		))

		buildWorker.RegisterGRPC(grpcServer)
		registryDownloadService.RegisterGRPC(grpcServer)

		reflection.Register(grpcServer)

		return cfg.GRPC.ListenAndServe(ctx, grpcServer)
	})

	return runGroup.Run(ctx)
}
