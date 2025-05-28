package protobuild

import (
	"fmt"
	"slices"
	"strings"

	"github.com/pentops/j5/internal/j5s/j5convert"
	"github.com/pentops/j5/internal/j5s/protobuild/psrc"
)

func NewCircularDependencyError(chain []string, dep string) error {
	return &CircularDependencyError{
		Chain: chain,
		Dep:   dep,
	}
}

type CircularDependencyError struct {
	Chain []string
	Dep   string
}

func (e *CircularDependencyError) Error() string {
	return fmt.Sprintf("circular dependency detected: %s -> %s", strings.Join(e.Chain, " -> "), e.Dep)
}

type Package struct {
	Name        string
	SourceFiles []*SourceFile

	Files              map[string]*psrc.File
	DirectDependencies map[string]*Package
	Exports            map[string]*j5convert.TypeRef

	pkgDeps map[string]struct{}

	Built *BuiltPackage
}

func (ps *PackageSet) newPackage(name string) *Package {
	if _, ok := ps.Packages[name]; ok {
		panic(fmt.Sprintf("package %s already exists", name))
	}

	pkg := &Package{
		Name:               name,
		DirectDependencies: map[string]*Package{},
		Exports:            map[string]*j5convert.TypeRef{},
		Files:              map[string]*psrc.File{},
		pkgDeps:            map[string]struct{}{},
	}
	ps.Packages[name] = pkg
	return pkg
}

func (pkg *Package) includeIO(summary *j5convert.FileSummary) {
	for _, exp := range summary.Exports {
		pkg.Exports[exp.Name] = exp
	}

	for _, ref := range summary.TypeDependencies {
		pkg.pkgDeps[ref.Package] = struct{}{}
	}

	for _, file := range summary.FileDependencies {
		dependsOn, _, err := j5convert.SplitPackageFromFilename(file)
		if err != nil {
			// fallback to full name
			dependsOn = j5convert.PackageFromFilename(file)
		}
		pkg.pkgDeps[dependsOn] = struct{}{}
	}
}

// ResolveType implements j5convert.TypeResolver interface.
func (pkg *Package) ResolveType(pkgName string, name string) (*j5convert.TypeRef, error) {
	if pkgName == pkg.Name {
		gotType, ok := pkg.Exports[name]
		if ok {
			return gotType, nil
		}
		return nil, &j5convert.TypeNotFoundError{
			// no package, is own package.
			Name: name,
		}
	}

	pkg, ok := pkg.DirectDependencies[pkgName]
	if !ok {
		return nil, fmt.Errorf("ResolveType: package %s not loaded", pkgName)
	}

	gotType, ok := pkg.Exports[name]
	if ok {
		return gotType, nil
	}

	return nil, &j5convert.TypeNotFoundError{
		Package: pkgName,
		Name:    name,
	}
}

type resolveBaton struct {
	chain []string
}

func newResolveBaton() *resolveBaton {
	return &resolveBaton{
		chain: []string{},
	}
}

func (rb *resolveBaton) cloneFor(name string) (*resolveBaton, error) {
	if slices.Contains(rb.chain, name) {
		return nil, NewCircularDependencyError(rb.chain, name)
	}

	return &resolveBaton{
		chain: append(slices.Clone(rb.chain), name),
	}, nil
}
