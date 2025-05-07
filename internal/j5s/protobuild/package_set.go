package protobuild

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/bufbuild/protocompile/linker"
	"github.com/pentops/j5/internal/j5s/protobuild/psrc"
	"github.com/pentops/log.go/log"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/descriptorpb"
)

type PackageSet struct {
	dependencyResolver psrc.Resolver
	sourceResolver     *sourceResolver

	// symbols is reused for the entire package set, all files must be linked
	// using the same symbols instance.
	symbols *linker.Symbols

	Packages map[string]*Package
}

func NewPackageSet(deps map[string]*descriptorpb.FileDescriptorProto, localFiles LocalFileSource) (*PackageSet, error) {
	resolver, err := psrc.ChainResolver(deps)

	if err != nil {
		return nil, fmt.Errorf("dependencyChainResolver: %w", err)
	}

	sourceResolver, err := newSourceResolver(localFiles)
	if err != nil {
		return nil, fmt.Errorf("newSourceResolver: %w", err)
	}

	cc := &PackageSet{
		dependencyResolver: resolver,
		sourceResolver:     sourceResolver,
		Packages:           map[string]*Package{},
		symbols:            &linker.Symbols{},
	}
	return cc, nil
}

func (ps *PackageSet) PackageForLocalFile(filename string) (string, bool, error) {
	return ps.sourceResolver.packageForFile(filename)
}

func (ps *PackageSet) LoadLocalPackage(ctx context.Context, pkgName string) (*Package, error) {
	rb := newResolveBaton()
	pkg, err := ps.loadPackage(ctx, rb, pkgName)
	if err != nil {
		return nil, fmt.Errorf("loadPackage %s: %w", pkgName, err)
	}
	ps.Packages[pkgName] = pkg
	return pkg, nil
}

func (ps *PackageSet) ListLocalPackages() []string {
	return ps.sourceResolver.ListPackages()
}

func (ps *PackageSet) GetLocalFileContent(ctx context.Context, filename string) (string, error) {
	data, err := ps.sourceResolver.getFileContent(ctx, filename)
	if err != nil {
		return "", fmt.Errorf("getFileContent %s: %w", filename, err)
	}
	return string(data), nil
}

func (ps *PackageSet) ListPackageFiles(pkgName string) ([]string, error) {
	// TODO: This skips *local* package files.
	return ps.dependencyResolver.ListPackageFiles(pkgName)
}

func (ps *PackageSet) FindFileByPath(filename string) (*psrc.File, error) {
	if filename == "" {
		return nil, errors.New("empty filename")
	}

	pkgName, isLocal, err := ps.sourceResolver.packageForFile(filename)
	if err != nil {
		return nil, fmt.Errorf("packageForFile: %w", err)
	}

	if !isLocal {
		file, err := ps.dependencyResolver.FindFileByPath(filename)
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

	if ps.sourceResolver.isLocalPackage(name) {
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

	fileNames, err := ps.sourceResolver.listPackageFiles(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("package files for (local) %s: %w", name, err)
	}

	pkg := newPackage(name)

	deps := map[string]struct{}{}
	for _, filename := range fileNames {
		file, err := ps.sourceResolver.getFile(ctx, filename)
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
		results, err := srcFile.toSearchResults(pkg)
		if err != nil {
			return nil, err
		}
		for _, result := range results {
			pkg.Files[result.Filename] = result
		}
	}

	return pkg, nil
}

func (ps *PackageSet) loadExternalPackage(ctx context.Context, rb *resolveBaton, name string) (*Package, error) {

	pkg := newPackage(name)

	filenames, err := ps.dependencyResolver.ListPackageFiles(name)
	if err != nil {
		return nil, fmt.Errorf("package files for (dependency) %s: %w", name, err)
	}

	deps := map[string]struct{}{}
	for _, filename := range filenames {
		file, err := ps.dependencyResolver.FindFileByPath(filename)
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

func (ps *PackageSet) CompilePackage(ctx context.Context, packageName string) ([]*psrc.File, error) {
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

	cc := newLinker(ps, ps.symbols)
	files, err := cc.resolveAll(ctx, filenames)
	if err != nil {
		ps.debugState(os.Stderr)
		return nil, fmt.Errorf("CompilePackage %s: %w", packageName, err)
	}

	return files, nil
}

func (ps *PackageSet) debugState(ww io.Writer) {

	fmt.Fprintln(ww, "PackageSet State:")
	for pkgName, pkg := range ps.Packages {
		fmt.Fprintf(ww, "  Package: %s\n", pkgName)
		for _, result := range pkg.Files {
			fmt.Fprintf(ww, "    psrc.File: %s (%s)\n", result.Summary.SourceFilename, result.SourceType.String())
		}
	}
}
