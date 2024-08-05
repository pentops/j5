package builder

import (
	"context"
	"io/fs"

	"github.com/pentops/envconf.go/envconf"
	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/internal/source"
)

type RemoteResolver interface {
	source.RemoteResolver
}

type ResolverConfig struct {
	RegistryAddress string `env:"J5_REGISTRY"`
	AuthToken       string `env:"J5_REGISTRY_TOKEN"`
}

func NewResolver(ctx context.Context, cfg ResolverConfig) (source.RemoteResolver, error) {
	if err := envconf.Parse(&cfg); err != nil {
		return nil, err
	}
	bufClient, err := source.NewBufClient()
	if err != nil {
		return nil, err
	}

	regClient, err := source.NewRegistryClient(cfg.RegistryAddress, cfg.AuthToken)
	if err != nil {
		return nil, err
	}

	return source.NewResolver(bufClient, regClient)
}

func BundleImageSource(ctx context.Context, fs fs.FS, bundleName string, resolver source.RemoteResolver) (*source_j5pb.SourceImage, *config_j5pb.BundleConfigFile, error) {
	src, err := source.NewFSSource(ctx, fs, resolver)
	if err != nil {
		return nil, nil, err
	}
	return src.BundleImageSource(ctx, bundleName)
}
