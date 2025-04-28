package protobuild

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/bufbuild/protocompile/linker"
	"github.com/pentops/j5/internal/j5s/j5convert"
	"github.com/pentops/log.go/log"
	"golang.org/x/exp/maps"
)

// ps.Packages[pkgName] = pkg
var _ PackageSrc = &PackageSet{}

type PackageSet struct {
	dependencyResolver fileSource
	localResolver      *sourceResolver

	Packages map[string]*Package
}

func NewPackageSet(deps DependencySet, localFiles LocalFileSource) (*PackageSet, error) {
	resolver, err := dependencyChainResolver(deps)

	if err != nil {
		return nil, fmt.Errorf("dependencyChainResolver: %w", err)
	}

	sourceResolver, err := newSourceResolver(localFiles)
	if err != nil {
		return nil, fmt.Errorf("newSourceResolver: %w", err)
	}

	cc := &PackageSet{
		dependencyResolver: resolver,
		localResolver:      sourceResolver,
		Packages:           map[string]*Package{},
	}
	return cc, nil
}

func (ps *PackageSet) PackageForLocalFile(filename string) (string, bool, error) {
	return ps.localResolver.packageForFile(filename)
}

func (ps *PackageSet) LoadLocalPackage(ctx context.Context, pkgName string) (*Package, *ErrCollector, error) {
	rb := newResolveBaton()
	pkg, err := ps.loadPackage(ctx, rb, pkgName)
	if err != nil {
		return nil, rb.errs, fmt.Errorf("loadPackage %s: %w", pkgName, err)
	}
	ps.Packages[pkgName] = pkg
	return pkg, rb.errs, nil
}

func (ps *PackageSet) ListLocalPackages() []string {
	return ps.localResolver.ListPackages()
}

func (ps *PackageSet) GetLocalFileContent(ctx context.Context, filename string) (string, error) {
	data, err := ps.localResolver.getFileContent(ctx, filename)
	if err != nil {
		return "", fmt.Errorf("getFileContent %s: %w", filename, err)
	}
	return string(data), nil
}

func (ps *PackageSet) packageFiles(pkgName string) ([]string, error) {
	// TODO: This skips *local* package files.
	return ps.dependencyResolver.packageFiles(pkgName)
}

func (ps *PackageSet) findFileByPath(filename string) (*SearchResult, error) {
	if filename == "" {
		return nil, errors.New("empty filename")
	}

	pkgName, isLocal, err := ps.localResolver.packageForFile(filename)
	if err != nil {
		return nil, fmt.Errorf("packageForFile: %w", err)
	}

	if !isLocal {
		file, err := ps.dependencyResolver.findFileByPath(filename)
		if err != nil {
			return nil, fmt.Errorf("readFile: %w", err)
		}
		return file, nil
	}

	pkg, ok := ps.Packages[pkgName]
	if !ok {
		return nil, fmt.Errorf("package %s not found for file %q", pkgName, filename)
	}

	res, ok := pkg.Files[filename]
	if ok {
		return res, nil
	}

	return nil, fmt.Errorf("file %s not found in package %s, have %s", filename, pkgName, strings.Join(maps.Keys(pkg.Files), ", "))
}

func (ps *PackageSet) loadPackage(ctx context.Context, rb *resolveBaton, name string) (*Package, error) {
	ctx = log.WithField(ctx, "loadPackage", name)
	rb, err := rb.cloneFor(name)
	if err != nil {
		return nil, fmt.Errorf("cloneFor %s: %w", name, err)
	}

	pkg, ok := ps.Packages[name]
	if ok {
		return pkg, nil
	}

	if ps.localResolver.isLocalPackage(name) {
		log.Debug(ctx, "Loading Local Package")
		pkg, err = ps.loadLocalPackage(ctx, rb, name)
		if err != nil {
			return nil, fmt.Errorf("loadLocalPackage %s: %w", name, err)
		}

	} else {
		log.Debug(ctx, "Loading External Package")
		pkg, err = ps.loadExternalPackage(ctx, rb, name)
		if err != nil {
			return nil, fmt.Errorf("loadExternalPackage %s: %w", name, err)
		}
	}
	log.Debug(ctx, "Loaded Package")
	ps.Packages[name] = pkg

	return pkg, nil
}

func (ps *PackageSet) resolveDependencies(ctx context.Context, rb *resolveBaton, pkg *Package, deps map[string]struct{}) error {
	delete(deps, pkg.Name)
	pkg.DirectDependencies = map[string]*Package{}
	for dep := range deps {
		depPkg, err := ps.loadPackage(ctx, rb, dep)
		if err != nil {
			return fmt.Errorf("loadPackage %s: %w", dep, err)
		}
		pkg.DirectDependencies[dep] = depPkg
	}
	return nil
}

func (ps *PackageSet) loadLocalPackage(ctx context.Context, rb *resolveBaton, name string) (*Package, error) {

	fileNames, err := ps.localResolver.listPackageFiles(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("package files for (local) %s: %w", name, err)
	}

	pkg := newPackage(name)

	deps := map[string]struct{}{}
	for _, filename := range fileNames {
		file, err := ps.localResolver.getFile(ctx, filename, rb.errs)
		if err != nil {
			return nil, fmt.Errorf("GetLocalFile %s: %w", filename, err)
		}
		pkg.SourceFiles = append(pkg.SourceFiles, file)
		pkg.includeIO(file.Summary, deps)
	}

	err = ps.resolveDependencies(ctx, rb, pkg, deps)
	if err != nil {
		return nil, fmt.Errorf("resolveDependencies for %s: %w", name, err)
	}

	for _, srcFile := range pkg.SourceFiles {
		if srcFile.Result != nil {
			pkg.Files[srcFile.Filename] = srcFile.Result
		} else if srcFile.J5Source != nil {
			descs, err := j5convert.ConvertJ5File(pkg, srcFile.J5Source)
			if err != nil {
				return nil, fmt.Errorf("convertJ5File %s: %w", srcFile.Filename, err)
			}

			for _, desc := range descs {
				pkg.Files[desc.GetName()] = &SearchResult{
					Summary:    srcFile.Summary,
					Desc:       desc,
					SourceType: LocalJ5Source,
				}
			}
		} else {
			return nil, fmt.Errorf("source file %s has no result and is not j5s", srcFile.Filename)
		}
	}

	return pkg, nil
}

func (ps *PackageSet) loadExternalPackage(ctx context.Context, rb *resolveBaton, name string) (*Package, error) {

	pkg := newPackage(name)

	filenames, err := ps.dependencyResolver.packageFiles(name)
	if err != nil {
		return nil, fmt.Errorf("package files for (dependency) %s: %w", name, err)
	}

	deps := map[string]struct{}{}
	for _, filename := range filenames {
		file, err := ps.dependencyResolver.findFileByPath(filename)
		if err != nil {
			return nil, fmt.Errorf("findFileByPath %s: %w", filename, err)
		}

		pkg.Files[file.Summary.SourceFilename] = file
		pkg.includeIO(file.Summary, deps)
	}

	err = ps.resolveDependencies(ctx, rb, pkg, deps)
	if err != nil {
		return nil, fmt.Errorf("resolveDependencies for %s: %w", name, err)
	}

	return pkg, nil
}

func (ps *PackageSet) CompilePackage(ctx context.Context, packageName string) (linker.Files, error) {
	ctx = log.WithField(ctx, "CompilePackage", packageName)
	log.Debug(ctx, "Compiler: Load")
	rb := newResolveBaton()

	pkg, err := ps.loadPackage(ctx, rb, packageName)
	if err != nil {
		return nil, fmt.Errorf("loadPackage %s: %w", packageName, err)
	}

	filenames := make([]string, 0)
	for filename := range pkg.Files {
		filenames = append(filenames, filename)
	}

	sort.Strings(filenames) // for consistent error ordering

	log.Debug(ctx, "Compiler: Link")

	errs := &ErrCollector{}
	cc := newLinker(ps, errs)
	files, err := cc.resolveAll(ctx, filenames)
	if err != nil {
		return nil, fmt.Errorf("CompilePackage %s: %w", packageName, err)
	}

	return files, nil
}
