package source

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/internal/j5s/j5convert"
	"github.com/pentops/j5/internal/j5s/protobuild"
	"github.com/pentops/j5/internal/j5s/protobuild/protomod"
	"github.com/pentops/j5/internal/j5s/protobuild/psrc"
	"github.com/pentops/log.go/log"
	"google.golang.org/protobuf/types/descriptorpb"
)

type Bundle interface {
	DebugName() string
	J5Config() (*config_j5pb.BundleConfigFile, error)
	SourceImage(ctx context.Context, resolver InputSource) (*source_j5pb.SourceImage, error)
	DirInRepo() string
	FS() fs.FS
	GetDependencies(ctx context.Context, resolver InputSource) (map[string]*descriptorpb.FileDescriptorProto, error)

	FileSource() (protobuild.LocalFileSource, error)
}

type bundleSource struct {
	nameInRepo *string
	debugName  string
	fs         fs.FS
	refConfig  *config_j5pb.BundleReference
	config     *config_j5pb.BundleConfigFile
	dirInRepo  string
}

func (bs bundleSource) NameInRepo() *string {
	return bs.nameInRepo
}

func (bs bundleSource) DirInRepo() string {
	return bs.dirInRepo
}

func (b bundleSource) FS() fs.FS {
	return b.fs
}

func (b bundleSource) DebugName() string {
	return b.debugName
}

func (b bundleSource) localDependencies() []string {
	deps := make([]string, 0)
	for _, dep := range b.config.Dependencies {
		local, ok := dep.Type.(*config_j5pb.Input_Local)
		if !ok {
			continue
		}
		deps = append(deps, local.Local)
	}

	for _, include := range b.config.Includes {
		local, ok := include.Input.Type.(*config_j5pb.Input_Local)
		if !ok {
			continue
		}
		deps = append(deps, local.Local)
	}
	return deps
}

func (b *bundleSource) J5Config() (*config_j5pb.BundleConfigFile, error) {
	return b.config, nil
}

func (b *bundleSource) SourceImage(ctx context.Context, resolver InputSource) (*source_j5pb.SourceImage, error) {
	img, err := b.readImageFromDir(ctx, resolver)
	if err != nil {
		return nil, fmt.Errorf("reading source image for %s: %w", b.debugName, err)
	}

	if img.SourceName == "" {
		img.SourceName = b.debugName
	}
	return img, nil
}

func (bundle *bundleSource) GetDependencies(ctx context.Context, resolver InputSource) (map[string]*descriptorpb.FileDescriptorProto, error) {
	ds, err := bundle.getDependencyFiles(ctx, resolver)
	if err != nil {
		return nil, err
	}

	return ds.primary, nil

}

func (bundle *bundleSource) getDependencyFiles(ctx context.Context, resolver InputSource) (*imageFiles, error) {
	ctx = log.WithField(ctx, "bundleDeps", bundle.DebugName())

	j5Config, err := bundle.J5Config()
	if err != nil {
		return nil, err
	}
	dependencies := make([]*source_j5pb.SourceImage, 0, len(j5Config.Dependencies))
	for _, dep := range j5Config.Dependencies {
		img, err := resolver.GetSourceImage(ctx, dep)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, img)
	}
	for _, dep := range j5Config.Includes {
		img, err := resolver.GetSourceImage(ctx, dep.Input)
		if err != nil {
			return nil, err
		}
		img.SourceName = inputName(dep.Input)
		dependencies = append(dependencies, img)
	}
	return combineSourceImages(dependencies)
}

func inputName(input *config_j5pb.Input) string {
	switch it := input.Type.(type) {
	case *config_j5pb.Input_Local:
		return it.Local
	case *config_j5pb.Input_Registry_:
		return fmt.Sprintf("%s/%s", it.Registry.Owner, it.Registry.Name)
	}

	return "<unknown>"
}

func (bundle *bundleSource) readImageFromDir(ctx context.Context, resolver InputSource) (*source_j5pb.SourceImage, error) {
	j5Config, err := bundle.J5Config()
	if err != nil {
		return nil, err
	}

	deps, err := bundle.getDependencyFiles(ctx, resolver)
	if err != nil {
		return nil, err
	}

	localFiles, err := bundle.FileSource()
	if err != nil {
		return nil, err
	}

	depResolver := psrc.DescriptorFiles(deps.primary)
	compiler, err := protobuild.NewPackageSet(depResolver, localFiles)
	if err != nil {
		return nil, err
	}

	img := newImageBuilder(deps)

	pkgNames := make([]string, 0, len(bundle.config.Packages))
	for _, pkg := range bundle.config.Packages {
		img.addPackage(&source_j5pb.Package{
			Name:  pkg.Name,
			Prose: pkg.Prose,
			Label: pkg.Label,
		})
		pkgNames = append(pkgNames, pkg.Name)
	}

	built, err := compiler.BuildPackages(ctx, pkgNames)
	if err != nil {
		return nil, err
	}

	for _, pkg := range built {
		err = img.addBuilt(pkg)
		if err != nil {
			return nil, err
		}
	}

	for _, spec := range j5Config.Includes {
		switch st := spec.Input.Type.(type) {
		case *config_j5pb.Input_Local:
			// include the local image in the source, since it will always have
			// the same commit hash
			localInclude, err := resolver.GetSourceImage(ctx, spec.Input)
			if err != nil {
				return nil, err
			}
			localInclude.SourceName = inputName(spec.Input)

			err = img.include(localInclude)
			if err != nil {
				return nil, err
			}

		case *config_j5pb.Input_Registry_:
			// reference the include spec in the output image so it can be
			// included on-the-fly at read time, taking the latest version
			img.img.Includes = append(img.img.Includes, &source_j5pb.Include{
				Name:      st.Registry.Name,
				Owner:     st.Registry.Owner,
				Version:   st.Registry.Version,
				Reference: st.Registry.Reference,
			})

		default:
			return nil, fmt.Errorf("unsupported include type %T", st)
		}
	}

	err = protomod.MutateImageWithMods(img.img, bundle.config.Mods)
	if err != nil {
		return nil, fmt.Errorf("MutateImageWithMods: %w", err)
	}

	return img.img, nil
}

func (bundle *bundleSource) FileSource() (protobuild.LocalFileSource, error) {
	bundleDir := bundle.DirInRepo()

	packages := []string{}
	packageMap := make(map[string]struct{})
	for _, pkg := range bundle.config.Packages {
		packages = append(packages, pkg.Name)
		packageMap[pkg.Name] = struct{}{}
	}

	sourcePackages, err := listPackages(bundle.fs, bundleDir)
	if err != nil {
		return nil, err
	}
	for _, pkg := range sourcePackages {
		if _, ok := packageMap[pkg]; ok {
			continue
		}
		packages = append(packages, pkg)
	}

	localFiles := &fileReader{
		fs:       bundle.fs,
		fsName:   bundleDir,
		packages: packages,
	}

	return localFiles, nil

}

// listPackages lists all of the directories in the root of the bundle
func listPackages(root fs.FS, pkgRoot string) ([]string, error) {
	packages := map[string]struct{}{}
	err := fs.WalkDir(root, ".", func(path string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if dirEntry.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".j5s" && ext != ".proto" {
			return nil
		}
		pkgName, _, err := j5convert.SplitPackageFromFilename(path)
		if err != nil {
			return fmt.Errorf("split package from filename %s: %w", path, err)
		}
		packages[pkgName] = struct{}{}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", pkgRoot, err)
	}
	pkgNames := make([]string, 0, len(packages))
	for k := range packages {
		if k == "." {
			continue
		}
		pkgNames = append(pkgNames, k)
	}

	sort.Strings(pkgNames)
	return pkgNames, nil
}

type fileReader struct {
	fs       fs.FS
	fsName   string
	packages []string
}

func (rr *fileReader) GetLocalFile(ctx context.Context, filename string) ([]byte, error) {
	return fs.ReadFile(rr.fs, filename)
}

func (rr *fileReader) ListPackages() []string {
	return rr.packages
}

func (rr *fileReader) ProseFiles(pkgName string) ([]*source_j5pb.ProseFile, error) {
	pkgRoot := strings.ReplaceAll(pkgName, ".", "/")
	proseFiles := []*source_j5pb.ProseFile{}
	err := fs.WalkDir(rr.fs, pkgRoot, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		ext := strings.ToLower(filepath.Ext(path))

		if ext != ".md" {
			return nil
		}

		data, err := fs.ReadFile(rr.fs, path)
		if err != nil {
			return err
		}
		proseFiles = append(proseFiles, &source_j5pb.ProseFile{
			Path:    path,
			Content: data,
		})
		return nil

	})
	if err != nil {
		return nil, err
	}
	return proseFiles, nil
}

func (rr *fileReader) ListSourceFiles(ctx context.Context, pkgName string) ([]string, error) {
	pkgRoot := strings.ReplaceAll(pkgName, ".", "/")

	files := make([]string, 0)
	err := fs.WalkDir(rr.fs, pkgRoot, func(path string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if dirEntry.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".j5s.proto") {
			return nil
		}
		if strings.HasSuffix(path, ".proto") || strings.HasSuffix(path, ".j5s") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", rr.fsName, err)
	}
	return files, nil
}

func (rr *fileReader) ListJ5Files(ctx context.Context) ([]string, error) {
	files := make([]string, 0)
	err := fs.WalkDir(rr.fs, ".", func(path string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if dirEntry.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".j5s") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil

}
