package protobuild

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

type fileResolver interface {
	findFileByPath(filename string) (*SearchResult, error)
	listPackageFiles(pkgName string) ([]string, error)
}

func dependencyChainResolver(deps DependencySet) (fileResolver, error) {
	depResolver, err := newDependencyResolver(deps)
	if err != nil {
		return nil, fmt.Errorf("newResolver: %w", err)
	}

	builtinResolver := newBuiltinResolver()
	resolver := newResolverCache(builtinResolver, depResolver)

	return resolver, nil
}

var errFileNotFound = fmt.Errorf("file not found")
var errPackageNotFound = fmt.Errorf("package not found")

type DependencySet interface {
	ListDependencyFiles(root string) []string
	GetDependencyFile(filename string) (*descriptorpb.FileDescriptorProto, error)
}

// dependencyResolver wraps a DependencySet to resolve proto sources
type dependencyResolver struct {
	deps DependencySet
}

func newDependencyResolver(externalDeps DependencySet) (*dependencyResolver, error) {
	rr := &dependencyResolver{
		deps: externalDeps,
	}
	return rr, nil
}

func (rr *dependencyResolver) listPackageFiles(pkgName string) ([]string, error) {
	root := strings.ReplaceAll(pkgName, ".", "/")

	filenames := rr.deps.ListDependencyFiles(root)
	if len(filenames) == 0 {
		return nil, fmt.Errorf("no files for package at %s", root)
	}

	return filenames, nil
}

func (rr *dependencyResolver) findFileByPath(filename string) (*SearchResult, error) {

	ec := &ErrCollector{}

	file, err := rr.deps.GetDependencyFile(filename)
	if err != nil {
		return nil, fmt.Errorf("dependency file: %w", err)
	}

	summary, err := buildSummaryFromDescriptor(file, ec)
	if err != nil {
		return nil, fmt.Errorf("summary for dependency %s: %w", file, err)
	}
	return &SearchResult{
		Summary:    summary,
		Desc:       file,
		SourceType: ExternalProtoSource,
	}, nil
}

var builtinPrefixes = []string{
	"buf/validate/",
	"google/api/",
	"google/protobuf/",
	"j5/auth/v1/",
	"j5/bcl/v1/",
	"j5/client/v1/",
	"j5/ext/v1/",
	"j5/list/v1/",
	"j5/messaging/v1/",
	"j5/schema/v1/",
	"j5/source/v1/",
	"j5/sourcedef/v1/",
	"j5/state/v1/",
	"j5/types/any/v1/",
	"j5/types/date/v1/",
	"j5/types/decimal/v1/",
}

type builtinResolver struct {
}

func newBuiltinResolver() *builtinResolver {
	return &builtinResolver{}
}

func (br *builtinResolver) hasRoot(filename string) bool {
	for _, prefix := range builtinPrefixes {
		if strings.HasPrefix(filename, prefix) {
			return true
		}
	}
	return false
}

func (br *builtinResolver) listPackageFiles(pkgName string) ([]string, error) {
	root := strings.ReplaceAll(pkgName, ".", "/") + "/"
	isBuiltin := br.hasRoot(root)
	if !isBuiltin {
		return nil, errPackageNotFound
	}
	files := []string{}
	protoregistry.GlobalFiles.RangeFilesByPackage(protoreflect.FullName(pkgName), func(refl protoreflect.FileDescriptor) bool {
		files = append(files, refl.Path())
		return true
	})

	return files, nil

}

func (br *builtinResolver) findFileByPath(filename string) (*SearchResult, error) {
	if !br.hasRoot(filename) {
		return nil, errFileNotFound
	}

	refl, err := protoregistry.GlobalFiles.FindFileByPath(filename)
	if err != nil {
		return nil, fmt.Errorf("find builtin file %s: %w", filename, err)
	}

	ec := &ErrCollector{}
	summary, err := buildSummaryFromReflect(refl, ec)
	if err != nil {
		return nil, fmt.Errorf("summary for builtin %s: %w", filename, err)
	}
	return &SearchResult{
		Summary:    summary,
		Refl:       refl,
		SourceType: BuiltInProtoSource,
	}, nil
}

type resolverCache struct {
	cache   map[string]*SearchResult
	sources []fileResolver
}

func newResolverCache(sources ...fileResolver) *resolverCache {
	return &resolverCache{
		cache:   make(map[string]*SearchResult),
		sources: sources,
	}
}

func (rc *resolverCache) listPackageFiles(pkgName string) ([]string, error) {
	for _, source := range rc.sources {
		files, err := source.listPackageFiles(pkgName)
		if err != nil {
			if err == errPackageNotFound {
				continue // try next source
			}
			return nil, err
		}
		return files, nil
	}
	return nil, errPackageNotFound
}

func (rc *resolverCache) findFileByPath(filename string) (*SearchResult, error) {
	if res, ok := rc.cache[filename]; ok {
		return res, nil
	}

	for _, source := range rc.sources {
		file, err := source.findFileByPath(filename)
		if err != nil {
			if err == errFileNotFound {
				continue // try next source
			}
			return nil, err
		}
		rc.cache[filename] = file
		return file, nil
	}

	return nil, errFileNotFound
}
