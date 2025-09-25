package protobuild

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/bufbuild/protocompile/linker"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/internal/bcl/errpos"
	"github.com/pentops/j5/internal/dag"
	"github.com/pentops/j5/internal/j5s/protobuild/psrc"
	"github.com/pentops/log.go/log"
)

type LocalSourceResolver interface {
	ListPackages() []string
	PackageForFile(filename string) (string, bool, error)
	IsLocalPackage(name string) bool
	PackageSourceFiles(ctx context.Context, pkgName string) ([]*SourceFile, error)
	PackageProseFiles(pkgName string) ([]*source_j5pb.ProseFile, error)
}

type PackageSet struct {
	dependencyResolver psrc.Resolver
	sourceResolver     LocalSourceResolver

	// symbols is reused for the entire package set, all files must be linked
	// using the same symbols instance.
	symbols *linker.Symbols

	Packages map[string]*Package
}

func NewPackageSet(deps psrc.Resolver, sourceResolver LocalSourceResolver) (*PackageSet, error) {
	dependencyResolver, err := psrc.ChainResolver(deps)
	if err != nil {
		return nil, fmt.Errorf("dependencyChainResolver: %w", err)
	}

	cc := &PackageSet{
		dependencyResolver: dependencyResolver,
		sourceResolver:     sourceResolver,
		Packages:           map[string]*Package{},
		symbols:            &linker.Symbols{},
	}
	return cc, nil
}

func (ps *PackageSet) PackageForLocalFile(filename string) (string, bool, error) {
	return ps.sourceResolver.PackageForFile(filename)
}

func (ps *PackageSet) ListLocalPackages() []string {
	return ps.sourceResolver.ListPackages()
}

func (ps *PackageSet) ListPackageFiles(pkgName string) ([]string, error) {
	// TODO: This skips *local* package files.
	return ps.dependencyResolver.ListPackageFiles(pkgName)
}

func (ps *PackageSet) FindFileByPath(filename string) (*psrc.File, error) {
	if filename == "" {
		return nil, errors.New("empty filename")
	}

	pkgName, isLocal, err := ps.sourceResolver.PackageForFile(filename)
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

	return nil, fmt.Errorf("file %s not found in package %s", filename, pkgName)
}

func (ps *PackageSet) loadPackage(ctx context.Context, rb *resolveBaton, name string) (*Package, error) {
	ctx = log.WithField(ctx, "loadPackage", name)
	log.Debug(ctx, "Compiler: Load Package")
	rb, err := rb.cloneFor(name)
	if err != nil {
		return nil, err
	}

	isLocal := ps.sourceResolver.IsLocalPackage(name)

	pkg, ok := ps.Packages[name]
	if ok {
		if pkg.Built == nil && isLocal {
			err = ps.buildLocalPackage(ctx, pkg)
			if err != nil {
				return nil, fmt.Errorf("buildLocalPackage %s: %w", name, err)
			}
		}
		return pkg, nil
	}

	if isLocal {
		log.Debug(ctx, "Loading Local Package")
		pkg, err = ps.loadLocalPackage(ctx, rb, name)
		if err != nil {
			return nil, err
		}

	} else {
		log.Debug(ctx, "Loading External Package")
		pkg, err = ps.loadExternalPackage(ctx, rb, name)
		if err != nil {
			return nil, err
		}
	}
	log.Debug(ctx, "Loaded Package")

	return pkg, nil
}

func (ps *PackageSet) resolveDependencies(ctx context.Context, rb *resolveBaton, pkg *Package) error {
	pkg.DirectDependencies = map[string]*Package{}
	for dep := range pkg.pkgDeps {
		log.WithField(ctx, "dep", dep).Debug("resolve dependency")
		depPkg, err := ps.loadPackage(ctx, rb, dep)
		if err != nil {
			return fmt.Errorf("loadPackage %s, required for %s: %w", dep, pkg.Name, err)
		}
		pkg.DirectDependencies[dep] = depPkg
	}
	return nil
}

func (ps *PackageSet) localPackageIO(ctx context.Context, name string) (*Package, error) {

	pkg := ps.newPackage(name)

	files, err := ps.sourceResolver.PackageSourceFiles(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("sourceFiles for %s: %w", name, err)
	}
	for _, file := range files {
		pkg.SourceFiles = append(pkg.SourceFiles, file)
		pkg.includeIO(file.Summary)
	}

	delete(pkg.pkgDeps, pkg.Name)

	return pkg, nil

}

func (pkg *Package) buildSearchResults() error {
	for _, srcFile := range pkg.SourceFiles {
		results, err := srcFile.toSearchResults(pkg)
		if err != nil {
			return err
		}
		for _, result := range results {
			pkg.Files[result.Filename] = result
		}
	}
	return nil
}

func (ps *PackageSet) loadLocalPackage(ctx context.Context, rb *resolveBaton, name string) (*Package, error) {

	pkg, err := ps.localPackageIO(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("localPackageIO %s: %w", name, err)
	}

	err = ps.resolveDependencies(ctx, rb, pkg)
	if err != nil {
		return nil, fmt.Errorf("resolveDependencies for %s: %w", name, err)
	}

	err = ps.buildLocalPackage(ctx, pkg)
	if err != nil {
		return nil, fmt.Errorf("buildLocalPackage for %s: %w", name, err)
	}
	return pkg, nil
}

func (ps *PackageSet) buildLocalPackage(ctx context.Context, pkg *Package) error {
	err := pkg.buildSearchResults()
	if err != nil {
		return fmt.Errorf("buildSearchResults for %s: %w", pkg.Name, err)
	}

	filenames := make([]string, 0)
	for filename := range pkg.Files {
		filenames = append(filenames, filename)
	}

	sort.Strings(filenames) // for consistent error ordering

	cc := newLinker(ps, ps.symbols)
	files, err := cc.resolveAll(ctx, filenames)
	if err != nil {
		return fmt.Errorf("resolveAll files for %s: %w", pkg.Name, err)
	}

	prose, err := ps.sourceResolver.PackageProseFiles(pkg.Name)
	if err != nil {
		return err
	}

	packageDeps := make([]*BuiltPackage, 0, len(pkg.DirectDependencies))
	for _, dep := range pkg.DirectDependencies {
		if dep.Built == nil {
			continue
		}
		packageDeps = append(packageDeps, dep.Built)
	}

	pkg.Built = &BuiltPackage{
		Name:         pkg.Name,
		Proto:        files,
		Prose:        prose,
		Dependencies: packageDeps,
	}

	return nil
}

func (ps *PackageSet) loadExternalPackage(ctx context.Context, rb *resolveBaton, name string) (*Package, error) {

	pkg := ps.newPackage(name)

	filenames, err := ps.dependencyResolver.ListPackageFiles(name)
	if err != nil {
		return nil, fmt.Errorf("package files for (dependency) %s: %w", name, err)
	}

	for _, filename := range filenames {
		file, err := ps.dependencyResolver.FindFileByPath(filename)
		if err != nil {
			return nil, fmt.Errorf("findFileByPath %s: %w", filename, err)
		}

		pkg.Files[file.Summary.SourceFilename] = file
		pkg.includeIO(file.Summary)
	}
	delete(pkg.pkgDeps, pkg.Name)

	err = ps.resolveDependencies(ctx, rb, pkg)
	if err != nil {
		return nil, fmt.Errorf("resolveDependencies for %s: %w", name, err)
	}

	return pkg, nil
}

type BuiltPackage struct {
	Name         string
	Proto        []*psrc.File
	Prose        []*source_j5pb.ProseFile
	Dependencies []*BuiltPackage
}

func (ps *PackageSet) CompilePackage(ctx context.Context, packageName string) (*BuiltPackage, error) {

	ctx = log.WithField(ctx, "CompilePackage", packageName)
	log.Debug(ctx, "Compiler: Load")
	rb := newResolveBaton()

	pkg, err := ps.loadPackage(ctx, rb, packageName)
	if err != nil {
		return nil, fmt.Errorf("loadPackage %s: %w", packageName, err)
	}

	return pkg.Built, nil
}

func (ps *PackageSet) BuildPackages(ctx context.Context, pkgNames []string) ([]*BuiltPackage, error) {
	var packageNodes []dag.Node

	packages := make(map[string]*Package)

	globalErrors := &errpos.ErrorsWithSource{}

	// IO Summary for all packages
	for _, pkgName := range pkgNames {
		if !ps.sourceResolver.IsLocalPackage(pkgName) {
			return nil, fmt.Errorf("package %s is not a local package", pkgName)
		}

		pkg, err := ps.localPackageIO(ctx, pkgName)
		if err != nil {
			if ep, ok := errpos.AsErrorsWithSource(err); ok {
				globalErrors.Append(ep)
				continue
			} else {
				return nil, fmt.Errorf("localPackageIO %s: %w", pkgName, err)
			}
		}
		packages[pkgName] = pkg
	}

	if len(globalErrors.Errors) > 0 {
		return nil, globalErrors
	}

	// Identify dependencies within provided packages
	for _, pkg := range packages {
		var deps []string
		for dep := range pkg.pkgDeps {
			if _, ok := packages[dep]; ok {
				deps = append(deps, dep)
			}
		}

		packageNodes = append(packageNodes, dag.Node{
			Name:          pkg.Name,
			IncomingEdges: deps,
		})
	}

	// sort the packages in topological order
	sortedPackages, err := dag.SortDAG(packageNodes)
	if err != nil {
		return nil, fmt.Errorf("sortDAG: %w", err)
	}

	done := map[string]struct{}{}
	out := []*BuiltPackage{}

	var addPkg func(pkg *BuiltPackage) error
	addPkg = func(pkg *BuiltPackage) error {
		if _, ok := done[pkg.Name]; ok {
			return nil
		}
		done[pkg.Name] = struct{}{}

		for _, dep := range pkg.Dependencies {
			if err := addPkg(dep); err != nil {
				return fmt.Errorf("addPkg %s: %w", dep.Name, err)
			}
		}
		out = append(out, pkg)
		return nil
	}

	for _, pkgName := range sortedPackages {
		pkg := packages[pkgName]

		rb := newResolveBaton()
		err = ps.resolveDependencies(ctx, rb, pkg)
		if err != nil {
			return nil, fmt.Errorf("resolveDependencies for %s: %w", pkg.Name, err)
		}

		err := ps.buildLocalPackage(ctx, pkg)
		if err != nil {
			return nil, fmt.Errorf("buildLocalPackage %s: %w", pkg.Name, err)
		}

		if err := addPkg(pkg.Built); err != nil {
			return nil, fmt.Errorf("addPkg %s: %w", pkgName, err)
		}
	}

	return out, nil
}
