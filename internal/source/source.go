package source

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"

	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// A repo has:
// - Config, with:
//   - Generate
//   - Plugins
// - Bundles, with:
//   - Dependencies
//   - Publish
// - Lock File

// Registry codebase uses builder.RunPublishBuild
// Builder's two build methods take:
//   - SourceImage.
//   - GenerateConfig or PublishConfig
//     Both of which take a list of plugins.
//     Publish also contains output format / destination.

// Build:
// - Fetch Repo
// - Parse Config
// - Identify one or more Bundles as input
// - Get a SourceImage for each bundle
//   - Resolve Dependencies
//   - Proto Parse -> Source Image
// - Merge SourceImages
// - Run Plugins
// - (Format Output)
// - Write / Upload Output

// We need:
// A set of bundles, accessing the code and config
// A dependency resolver to download the SourceImage for dependencies
// Config manager for the plugins

// RemoteResolver fetches, locks and caches dependencies from buf and j5
type RemoteResolver interface {
	GetRemoteDependency(ctx context.Context, input *config_j5pb.Input, locks *config_j5pb.LockFile) (*source_j5pb.SourceImage, error)
	LatestLocks(ctx context.Context, deps []*config_j5pb.Input) (*config_j5pb.LockFile, error)
}

type InputSource interface {
	GetSourceImage(ctx context.Context, input *config_j5pb.Input) (*source_j5pb.SourceImage, error)
}

type Source struct {
	thisRepo *repo
	resolver RemoteResolver
}

func NewFSSource(ctx context.Context, root fs.FS, resolver RemoteResolver) (*Source, error) {
	src := &Source{
		resolver: resolver,
	}

	thisRepo, err := src.newRepo(".", root)
	if err != nil {
		return nil, err
	}
	src.thisRepo = thisRepo

	return src, nil
}

func (src *Source) ListAllDependencies() ([]*config_j5pb.Input, error) {
	allDeps := []*config_j5pb.Input{}
	for _, bundle := range src.thisRepo.bundles {
		cfg, err := bundle.J5Config()
		if err != nil {
			return nil, fmt.Errorf("bundle %q: %w", bundle.Name(), err)
		}
		allDeps = append(allDeps, cfg.Dependencies...)
	}
	return allDeps, nil
}

func (src *Source) GetSourceImage(ctx context.Context, input *config_j5pb.Input) (*source_j5pb.SourceImage, error) {
	if local, ok := input.Type.(*config_j5pb.Input_Local); ok {
		bundle, ok := src.thisRepo.bundles[local.Local]
		if !ok {
			return nil, fmt.Errorf("bundle %q not found", local.Local)
		}
		return bundle.SourceImage(ctx, src)
	}

	return src.resolver.GetRemoteDependency(ctx, input, src.thisRepo.lockFile)
}

type repo struct {
	repoRoot fs.FS
	bundles  map[string]*bundleSource
	config   *config_j5pb.RepoConfigFile
	lockFile *config_j5pb.LockFile
}

func (src *Source) newRepo(debugName string, repoRoot fs.FS) (*repo, error) {

	config, err := readDirConfigs(repoRoot)
	if err != nil {
		return nil, err
	}

	if err := resolveConfigReferences(config); err != nil {
		return nil, fmt.Errorf("resolving config references: %w", err)
	}

	lockFile, err := readLockFile(repoRoot, "j5-lock.yaml")
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("reading lock file: %w", err)
		}
		lockFile = &config_j5pb.LockFile{}
	}

	thisRepo := &repo{
		config:   config,
		repoRoot: repoRoot,
		lockFile: lockFile,
		bundles:  map[string]*bundleSource{},
	}

	for _, refConfig := range config.Bundles {
		bundleRoot := repoRoot
		if refConfig.Dir != "" {
			bundleRoot, err = fs.Sub(bundleRoot, refConfig.Dir)
			if err != nil {
				return nil, fmt.Errorf("subdir %q: %w", refConfig.Dir, err)
			}
		}

		thisRepo.bundles[refConfig.Name] = &bundleSource{
			debugName: fmt.Sprintf("%s/%s", debugName, refConfig.Dir),
			fs:        bundleRoot,
			dirInRepo: refConfig.Dir,
			refConfig: refConfig,
		}
	}

	if len(config.Packages) > 0 || len(config.Publish) > 0 || config.Registry != nil {
		// Inline Bundle
		thisRepo.bundles[""] = &bundleSource{
			debugName: debugName,
			fs:        repoRoot,
			refConfig: &config_j5pb.BundleReference{
				Name: "",
				Dir:  "",
			},
			config: &config_j5pb.BundleConfigFile{
				Registry:     config.Registry,
				Publish:      config.Publish,
				Packages:     config.Packages,
				Options:      config.Options,
				Dependencies: config.Dependencies,
			},
		}
	}

	return thisRepo, nil
}

func (src Source) RepoConfig() *config_j5pb.RepoConfigFile {
	return src.thisRepo.config
}

func (src Source) AllBundles() []*bundleSource {
	out := make([]*bundleSource, 0, len(src.thisRepo.bundles))
	for _, bundle := range src.thisRepo.bundles {
		out = append(out, bundle)
	}
	return out
}

func (src *Source) CombinedSourceImage(ctx context.Context, inputs []*config_j5pb.Input) (*source_j5pb.SourceImage, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("no inputs")
	}
	if len(inputs) == 1 {
		return src.GetSourceImage(ctx, inputs[0])
	}

	allFiles := map[string]string{}
	fullImage := &source_j5pb.SourceImage{
		Options: &config_j5pb.PackageOptions{},
	}
	for _, input := range inputs {
		img, err := src.GetSourceImage(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("input %v: %w", input, err)
		}

		wantFiles := map[string]struct{}{}
		for _, file := range img.SourceFilenames {
			wantFiles[file] = struct{}{}
		}

		for _, file := range img.File {
			hash, err := hashFile(file)
			if err != nil {
				return nil, fmt.Errorf("file %q: %w", *file.Name, err)
			}
			if existing, ok := allFiles[*file.Name]; ok {
				if existing != hash {
					return nil, fmt.Errorf("file %q has conflicting content", *file.Name)
				}
			} else {
				allFiles[*file.Name] = hash
				fullImage.File = append(fullImage.File, file)
				if _, ok := wantFiles[*file.Name]; ok {
					fullImage.SourceFilenames = append(fullImage.SourceFilenames, *file.Name)
				}
			}
		}

		if img.Options != nil {
			for _, subPkg := range img.Options.SubPackages {
				found := false
				for _, existing := range fullImage.Options.SubPackages {
					if existing.Name == subPkg.Name {
						// no config other than name for now.
						found = true
						break
					}
				}
				if !found {
					fullImage.Options.SubPackages = append(fullImage.Options.SubPackages, subPkg)
				}
			}
		}

		fullImage.Packages = append(fullImage.Packages, img.Packages...)
	}

	return fullImage, nil
}

func hashFile(file *descriptorpb.FileDescriptorProto) (string, error) {
	sh := sha256.New()
	fileContent, err := proto.Marshal(file)
	if err != nil {
		return "", err
	}
	sh.Write(fileContent)
	return base64.StdEncoding.EncodeToString(sh.Sum(nil)), nil

}
func (src *Source) BundleImageSource(ctx context.Context, name string) (*source_j5pb.SourceImage, *config_j5pb.BundleConfigFile, error) {
	bundleSource, err := src.BundleSource(name)
	if err != nil {
		return nil, nil, err
	}

	img, err := bundleSource.SourceImage(ctx, src)
	if err != nil {
		return nil, nil, err
	}

	cfg, err := bundleSource.J5Config()
	if err != nil {
		return nil, nil, err
	}

	return img, cfg, nil
}

func (src *Source) BundleSource(name string) (*bundleSource, error) {
	if name != "" {
		if bundle, ok := src.thisRepo.bundles[name]; ok {
			return bundle, nil
		}
		return nil, fmt.Errorf("bundle %q not found", name)
	}
	if len(src.thisRepo.bundles) == 0 {
		return nil, fmt.Errorf("no bundles found")
	}
	if len(src.thisRepo.bundles) > 1 {
		return nil, fmt.Errorf("multiple bundles found, must specify a name")
	}

	for _, bundle := range src.thisRepo.bundles {
		return bundle, nil
	}

	return nil, fmt.Errorf("no bundles found")

}

func (src *Source) SourceFile(ctx context.Context, filename string) ([]byte, error) {
	return fs.ReadFile(src.thisRepo.repoRoot, filename)
}
