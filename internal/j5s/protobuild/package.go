package protobuild

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/pentops/j5/gen/j5/sourcedef/v1/sourcedef_j5pb"
	"github.com/pentops/j5/internal/j5s/j5convert"
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

type SourceFile struct {
	Filename string
	Summary  *j5convert.FileSummary

	J5Source  *sourcedef_j5pb.SourceFile
	RawSource []byte

	Result *SearchResult
}

type Package struct {
	Name        string
	SourceFiles []*SourceFile

	Files              map[string]*SearchResult
	DirectDependencies map[string]*Package
	Exports            map[string]*j5convert.TypeRef
}

func newPackage(name string) *Package {
	pkg := &Package{
		Name:               name,
		DirectDependencies: map[string]*Package{},
		Exports:            map[string]*j5convert.TypeRef{},
		Files:              map[string]*SearchResult{},
	}
	return pkg
}

func (pkg *Package) includeIO(summary *j5convert.FileSummary, deps map[string]struct{}) {
	for _, exp := range summary.Exports {
		pkg.Exports[exp.Name] = exp
	}

	for _, ref := range summary.TypeDependencies {
		deps[ref.Package] = struct{}{}
	}
	for _, file := range summary.FileDependencies {
		dependsOn := j5convert.PackageFromFilename(file)
		deps[dependsOn] = struct{}{}
	}
}

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
	errs  *ErrCollector
}

func newResolveBaton() *resolveBaton {
	return &resolveBaton{
		chain: []string{},
		errs:  &ErrCollector{},
	}
}

func (rb *resolveBaton) cloneFor(name string) (*resolveBaton, error) {
	for _, ancestor := range rb.chain {
		if name == ancestor {
			return nil, NewCircularDependencyError(rb.chain, name)
		}
	}

	return &resolveBaton{
		chain: append(slices.Clone(rb.chain), name),
		errs:  rb.errs,
	}, nil
}

type PackageSrc interface {
	fileSource
	PackageForLocalFile(filename string) (string, bool, error)
	LoadLocalPackage(ctx context.Context, pkgName string) (*Package, *ErrCollector, error)
	ListLocalPackages() []string
	GetLocalFileContent(ctx context.Context, filename string) (string, error)
}
