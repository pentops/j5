package source

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"google.golang.org/protobuf/types/pluginpb"
)

// BundleSource is the input directory in a local repository with source files
// and config.
type BundleSource interface {
	J5Config() (*config_j5pb.BundleConfigFile, error)
	Input
}

// Input is any bundle, local or remote.
type Input interface {
	ProtoCodeGeneratorRequest(ctx context.Context) (*pluginpb.CodeGeneratorRequest, error)
	SourceImage(ctx context.Context) (*source_j5pb.SourceImage, error)
	Name() string
}

type bundleSource struct {
	debugName  string
	rootSource *Source
	repo       *repo
	refConfig  *config_j5pb.BundleReference
	config     *config_j5pb.BundleConfigFile
	dirInRepo  string
}

func (b bundleSource) fs() (fs.FS, error) {
	if b.dirInRepo == "" {
		return b.repo.repoRoot, nil
	}
	return fs.Sub(b.repo.repoRoot, b.dirInRepo)
}

func (b bundleSource) Name() string {
	return b.debugName
}

func (b *bundleSource) J5Config() (*config_j5pb.BundleConfigFile, error) {
	if b.config != nil {
		return b.config, nil
	}

	root, err := b.fs()
	if err != nil {
		return nil, fmt.Errorf("bundle %s: %w", b.debugName, err)
	}

	config, err := readBundleConfigFile(root, "j5.yaml")
	if err != nil {
		return nil, fmt.Errorf("config for bundle %s: %w", b.debugName, err)
	}

	b.config = config
	return config, nil
}

func (b *bundleSource) ProtoCodeGeneratorRequest(ctx context.Context) (*pluginpb.CodeGeneratorRequest, error) {
	rr, err := codeGeneratorRequestFromSource(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("codegen for bundle %s: %w", b.debugName, err)
	}

	return rr, nil
}

func (b *bundleSource) SourceImage(ctx context.Context) (*source_j5pb.SourceImage, error) {
	img, err := b.rootSource.readImageFromDir(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("reading source image for %s: %w", b.debugName, err)
	}

	return img, nil
}

type imageBundle struct {
	source  *source_j5pb.SourceImage
	name    string
	version string
}

func (ib imageBundle) Name() string {
	if ib.version == "" {
		return ib.name
	}
	return fmt.Sprintf("%s:%s", ib.name, ib.version)
}

func (ib imageBundle) ProtoCodeGeneratorRequest(ctx context.Context) (*pluginpb.CodeGeneratorRequest, error) {
	return codeGeneratorRequestFromImage(ib.source)
}

func (ib imageBundle) SourceImage(ctx context.Context) (*source_j5pb.SourceImage, error) {
	return ib.source, nil
}
