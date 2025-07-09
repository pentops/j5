package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/pentops/j5/buildlib"
	"github.com/pentops/log.go/log"
	"github.com/pentops/j5/internal/registry/anyfs"
	"github.com/pentops/j5/internal/registry/buildwrap"
	"github.com/pentops/j5/internal/registry/github"
	"github.com/pentops/j5/internal/registry/gomodproxy"
	"github.com/pentops/j5/internal/registry/packagestore"
	"github.com/pentops/j5/internal/registry/service"
	"github.com/pentops/runner"
	"github.com/pentops/runner/commander"
	"github.com/pentops/sqrlx.go/sqrlx"
	"github.com/pressly/goose"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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
	service.DBConfig
}) error {

	db, err := cfg.OpenDatabase(ctx)
	if err != nil {
		return err
	}

	return goose.Up(db, cfg.MigrationsDir)
}

func runReadonlyServer(ctx context.Context, cfg struct {
	HTTPPort int `env:"HTTP_PORT" default:"8081"`
	GRPCPort int `env:"GRPC_PORT" default:"8080"`
	service.DBConfig
	PackageStoreConfig
}) error {
	dbConn, err := cfg.OpenDatabase(ctx)
	if err != nil {
		return err
	}

	db := sqrlx.NewPostgres(dbConn)

	pkgStore, err := cfg.OpenPackageStore(ctx, db)
	if err != nil {
		return err
	}

	registryDownloadService := service.NewRegistryService(pkgStore)

	runGroup := runner.NewGroup(runner.WithName("main"), runner.WithCancelOnSignals())

	runGroup.Add("httpServer", func(ctx context.Context) error {
		handler := gomodproxy.Handler(pkgStore)

		httpServer := &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
			Handler: handler,
		}
		log.WithField(ctx, "port", cfg.HTTPPort).Info("Begin Registry Server")

		go func() {
			<-ctx.Done()
			httpServer.Shutdown(ctx) // nolint:errcheck
		}()

		return httpServer.ListenAndServe()
	})

	runGroup.Add("grpcServer", func(ctx context.Context) error {
		grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
			service.GRPCMiddleware()...,
		))

		registryDownloadService.RegisterGRPC(grpcServer)

		reflection.Register(grpcServer)

		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
		if err != nil {
			return err
		}
		log.WithField(ctx, "port", cfg.GRPCPort).Info("Begin Worker Server")
		go func() {
			<-ctx.Done()
			grpcServer.GracefulStop() // nolint:errcheck
		}()

		return grpcServer.Serve(lis)
	})

	return runGroup.Run(ctx)
}

type PackageStoreConfig struct {
	Storage string `env:"REGISTRY_STORAGE"`
}

func (cfg PackageStoreConfig) OpenPackageStore(ctx context.Context, db sqrlx.Transactor) (*packagestore.PackageStore, error) {
	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	s3Client := s3.NewFromConfig(awsConfig)

	s3fs, err := anyfs.NewS3FS(s3Client, cfg.Storage)
	if err != nil {
		return nil, err
	}

	pkgStore, err := packagestore.NewPackageStore(db, s3fs)
	if err != nil {
		return nil, err
	}

	return pkgStore, nil
}

func runCombinedServer(ctx context.Context, cfg struct {
	HTTPPort int `env:"HTTP_PORT" default:"8081"`
	GRPCPort int `env:"GRPC_PORT" default:"8080"`
	PackageStoreConfig
	service.DBConfig
}) error {
	dbConn, err := cfg.OpenDatabase(ctx)
	if err != nil {
		return err
	}

	db := sqrlx.NewPostgres(dbConn)

	pkgStore, err := cfg.OpenPackageStore(ctx, db)
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
	j5Builder, err := buildlib.NewBuilder(regWrap)
	if err != nil {
		return err
	}

	buildWorker := buildwrap.NewBuildWorker(j5Builder, githubClient, pkgStore, dbPublisher)

	registryDownloadService := service.NewRegistryService(pkgStore)

	runGroup := runner.NewGroup(runner.WithName("main"), runner.WithCancelOnSignals())

	runGroup.Add("httpServer", func(ctx context.Context) error {
		handler := gomodproxy.Handler(pkgStore)

		httpServer := &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
			Handler: handler,
		}
		log.WithField(ctx, "port", cfg.HTTPPort).Info("Begin Registry Server")

		go func() {
			<-ctx.Done()
			httpServer.Shutdown(ctx) // nolint:errcheck
		}()

		return httpServer.ListenAndServe()
	})

	runGroup.Add("grpcServer", func(ctx context.Context) error {
		grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
			service.GRPCMiddleware()...,
		))

		buildWorker.RegisterGRPC(grpcServer)
		registryDownloadService.RegisterGRPC(grpcServer)

		reflection.Register(grpcServer)

		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
		if err != nil {
			return err
		}
		log.WithField(ctx, "port", cfg.GRPCPort).Info("Begin Worker Server")
		go func() {
			<-ctx.Done()
			grpcServer.GracefulStop() // nolint:errcheck
		}()

		return grpcServer.Serve(lis)
	})

	return runGroup.Run(ctx)
}
