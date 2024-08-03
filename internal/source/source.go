package source

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/bufbuild/protoyaml-go"
	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

type Source struct {
	thisRepo *repo

	bufCache *BufCache

	regClient *registryClient

	lockWriter *lockWriter
	locks      *config_j5pb.LockFile

	j5Cache *j5Cache
}

type lockWriter struct {
	filename string
}

func (lw *lockWriter) write(lock *config_j5pb.LockFile) error {
	data, err := protoyaml.MarshalOptions{}.Marshal(lock)
	if err != nil {
		return err
	}
	return os.WriteFile(lw.filename, data, 0666)
}

func NewSource(ctx context.Context, rootDir string) (*Source, error) {
	fsRoot := os.DirFS(rootDir)
	fsSource, err := NewFSSource(ctx, fsRoot)
	if err != nil {
		return nil, err
	}

	fsSource.lockWriter = &lockWriter{
		filename: filepath.Join(rootDir, "j5-lock.yaml"),
	}

	if fsSource.locks == nil {
		fsSource.locks = &config_j5pb.LockFile{}
	}

	fsSource.j5Cache, err = newJ5Cache()
	if err != nil {
		return nil, err
	}
	return fsSource, nil
}

func NewFSSource(ctx context.Context, root fs.FS) (*Source, error) {

	bufCache, err := NewBufCache()
	if err != nil {
		return nil, err
	}

	regClient := envRegistryClient()

	src := &Source{
		regClient: regClient,
		bufCache:  bufCache,
	}

	thisRepo, err := src.newRepo(".", root)
	if err != nil {
		return nil, err
	}
	src.thisRepo = thisRepo

	lock, err := readLockFile(root, "j5-lock.yaml")
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("reading lock file: %w", err)
		}
	} else {
		src.locks = lock
	}

	return src, nil
}

type repo struct {
	repoRoot fs.FS
	bundles  map[string]*bundleSource
	config   *config_j5pb.RepoConfigFile
}

func (src *Source) newRepo(debugName string, repoRoot fs.FS) (*repo, error) {

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
		bundles:  map[string]*bundleSource{},
	}

	for _, refConfig := range config.Bundles {
		thisRepo.bundles[refConfig.Name] = &bundleSource{
			rootSource: src,
			debugName:  fmt.Sprintf("%s/%s", debugName, refConfig.Dir),
			repo:       thisRepo,
			dirInRepo:  refConfig.Dir,
			refConfig:  refConfig,
		}
	}

	if len(config.Packages) > 0 || len(config.Publish) > 0 || config.Registry != nil {
		// Inline Bundle
		thisRepo.bundles[""] = &bundleSource{
			debugName:  debugName,
			repo:       thisRepo,
			rootSource: src,
			dirInRepo:  "",
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
func (src Source) RepoConfig() *config_j5pb.RepoConfigFile {
	return src.thisRepo.config
}

func (src Source) AllBundles() []BundleSource {
	out := make([]BundleSource, 0, len(src.thisRepo.bundles))
	for _, bundle := range src.thisRepo.bundles {
		out = append(out, bundle)
	}
	return out
}

func (src *Source) CombinedInput(ctx context.Context, inputs []*config_j5pb.Input) (Input, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("no inputs")
	}
	if len(inputs) == 1 {
		return src.GetInput(ctx, inputs[0])
	}

	allFiles := map[string]string{}
	fullImage := &source_j5pb.SourceImage{
		Options: &config_j5pb.PackageOptions{},
	}
	for _, input := range inputs {
		bundle, err := src.GetInput(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("input %v: %w", input, err)
		}
		img, err := bundle.SourceImage(ctx)
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

	return &imageBundle{
		source: fullImage,
		name:   "combined",
	}, nil
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
func (src *Source) BundleSource(name string) (BundleSource, error) {
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
