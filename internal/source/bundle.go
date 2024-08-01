package source

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"google.golang.org/protobuf/types/pluginpb"
)

type Input interface {
	J5Config() (*config_j5pb.BundleConfigFile, error)
	ProtoCodeGeneratorRequest(ctx context.Context) (*pluginpb.CodeGeneratorRequest, error)
	SourceImage(ctx context.Context) (*source_j5pb.SourceImage, error)
	Name() string
}

type repo struct {
	repoRoot fs.FS
	bundles  map[string]*bundle
	config   *config_j5pb.RepoConfigFile
}

func newRepo(debugName string, repoRoot fs.FS) (*repo, error) {

	config, err := readDirConfigs(repoRoot)
	if err != nil {
		return nil, err
	}

	if err := resolveConfigReferences(config); err != nil {
		return nil, fmt.Errorf("resolving config references: %w", err)
	}

	thisRepo := &repo{
		config: config,
		//commitInfo: commitInfo,
		repoRoot: repoRoot,
		bundles:  map[string]*bundle{},
	}

	for _, refConfig := range config.Bundles {
		thisRepo.bundles[refConfig.Name] = &bundle{
			debugName: fmt.Sprintf("%s/%s", debugName, refConfig.Dir),
			repo:      thisRepo,
			dirInRepo: refConfig.Dir,
			refConfig: refConfig,
		}
	}

	if len(config.Packages) > 0 || len(config.Publish) > 0 || config.Registry != nil {
		// Inline Bundle
		thisRepo.bundles[""] = &bundle{
			debugName: debugName,
			repo:      thisRepo,
			dirInRepo: "",
			refConfig: &config_j5pb.BundleReference{
				Name: "",
				Dir:  "",
			},
			config: &config_j5pb.BundleConfigFile{
				Registry: config.Registry,
				Publish:  config.Publish,
				Packages: config.Packages,
				Options:  config.Options,
			},
		}
	}

	return thisRepo, nil

}

type bundle struct {
	debugName string
	repo      *repo
	refConfig *config_j5pb.BundleReference
	config    *config_j5pb.BundleConfigFile
	dirInRepo string
}

func (b bundle) fs() (fs.FS, error) {
	if b.dirInRepo == "" {
		return b.repo.repoRoot, nil
	}
	return fs.Sub(b.repo.repoRoot, b.dirInRepo)
}

func (b bundle) Name() string {
	return b.debugName
}

func (b *bundle) J5Config() (*config_j5pb.BundleConfigFile, error) {
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

func (b *bundle) ProtoCodeGeneratorRequest(ctx context.Context) (*pluginpb.CodeGeneratorRequest, error) {
	rr, err := codeGeneratorRequestFromSource(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("codegen for bundle %s: %w", b.debugName, err)
	}

	return rr, nil
}

func (b *bundle) SourceImage(ctx context.Context) (*source_j5pb.SourceImage, error) {
	img, err := readImageFromDir(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("reading source image for %s: %w", b.debugName, err)
	}

	return img, nil
}

type combinedBundle struct {
	source *source_j5pb.SourceImage
}

func (cb combinedBundle) Name() string {
	return "combined"
}

func (cb combinedBundle) J5Config() (*config_j5pb.BundleConfigFile, error) {
	return nil, nil
}

func (cb combinedBundle) ProtoCodeGeneratorRequest(ctx context.Context) (*pluginpb.CodeGeneratorRequest, error) {
	return codeGeneratorRequestFromImage(cb.source)
}

func (cb combinedBundle) SourceImage(ctx context.Context) (*source_j5pb.SourceImage, error) {
	return cb.source, nil
}

func (cb combinedBundle) CommitInfo(context.Context) (*source_j5pb.CommitInfo, error) {
	return nil, nil
}

type imageBundle struct {
	source *source_j5pb.SourceImage
	repo   *config_j5pb.RegistryConfig
}

func (ib imageBundle) Name() string {
	return fmt.Sprintf("%s/%s", ib.repo.Organization, ib.repo.Name)
}

func (ib imageBundle) J5Config() (*config_j5pb.BundleConfigFile, error) {

	return &config_j5pb.BundleConfigFile{
		Registry: ib.repo,
		Packages: ib.source.Packages,
	}, nil
}

func (ib imageBundle) ProtoCodeGeneratorRequest(ctx context.Context) (*pluginpb.CodeGeneratorRequest, error) {
	return codeGeneratorRequestFromImage(ib.source)
}

func (ib imageBundle) SourceImage(ctx context.Context) (*source_j5pb.SourceImage, error) {
	return ib.source, nil
}

func (ib imageBundle) CommitInfo(context.Context) (*source_j5pb.CommitInfo, error) {
	return nil, nil
}
