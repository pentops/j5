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
	"github.com/pentops/j5/internal/j5s/protobuild"
	"github.com/pentops/j5/internal/protosrc"
	"github.com/pentops/log.go/log"
	"google.golang.org/protobuf/reflect/protodesc"
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
	debugName string
	fs        fs.FS
	refConfig *config_j5pb.BundleReference
	config    *config_j5pb.BundleConfigFile
	dirInRepo string
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
		dependencies = append(dependencies, img)
	}
	return combineSourceImages(dependencies)
}

// getIncludes returns the images corresponding to the inputs. The returned
// slice will have the same indexes as the input.
func (bundle *bundleSource) getIncludes(ctx context.Context, resolver InputSource) ([]*source_j5pb.SourceImage, error) {
	j5Config, err := bundle.J5Config()
	if err != nil {
		return nil, err
	}
	dependencies := make([]*source_j5pb.SourceImage, len(j5Config.Includes))
	for idx, spec := range j5Config.Includes {
		img, err := resolver.GetSourceImage(ctx, spec.Input)
		if err != nil {
			return nil, err
		}

		dependencies[idx] = img
	}

	return dependencies, nil
}

func (bundle *bundleSource) readImageFromDir(ctx context.Context, resolver InputSource) (*source_j5pb.SourceImage, error) {

	deps, err := bundle.getDependencyFiles(ctx, resolver)
	if err != nil {
		return nil, err
	}

	localFiles, err := bundle.FileSource()
	if err != nil {
		return nil, err
	}

	compiler, err := protobuild.NewPackageSet(deps.primary, localFiles)
	if err != nil {
		return nil, err
	}

	img := newImageBuilder(deps)

	for _, pkg := range bundle.config.Packages {
		img.addPackage(&source_j5pb.PackageInfo{
			Name:  pkg.Name,
			Prose: pkg.Prose,
			Label: pkg.Label,
		})

		built, err := compiler.CompilePackage(ctx, pkg.Name)
		if err != nil {
			return nil, fmt.Errorf("compile package %s: %w", pkg.Name, err)
		}

		descriptors := make([]*descriptorpb.FileDescriptorProto, 0, len(built.Proto))
		for _, file := range built.Proto {
			descriptor := protodesc.ToFileDescriptorProto(file.Linked)
			descriptors = append(descriptors, descriptor)

		}

		sorted, err := protosrc.SortByDependency(descriptors)
		if err != nil {
			return nil, fmt.Errorf("sort by dependency: %w", err)
		}
		for _, descriptor := range sorted {
			if err := img.addFile(descriptor, true); err != nil {
				return nil, fmt.Errorf("add file %s: %w", descriptor.GetName(), err)
			}
		}
		for _, file := range built.Prose {
			img.addProseFile(file)
		}
	}

	includeImages, err := bundle.getIncludes(ctx, resolver)
	if err != nil {
		return nil, err
	}

	for _, included := range includeImages {
		img.include(included)
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
		pkgPath := filepath.Dir(path)
		packages[pkgPath] = struct{}{}
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
		pkgName := strings.ReplaceAll(k, "/", ".")
		pkgNames = append(pkgNames, pkgName)
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
