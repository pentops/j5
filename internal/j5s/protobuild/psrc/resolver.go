package psrc

import (
	"fmt"
)

type Resolver interface {
	FindFileByPath(filename string) (*File, error)
	ListPackageFiles(pkgName string) ([]string, error)
}

func ChainResolver(deps Resolver) (Resolver, error) {
	builtinResolver := NewBuiltinResolver()
	resolver := newResolverCache(builtinResolver, deps)
	return resolver, nil
}

var ErrFileNotFound = fmt.Errorf("file not found")
var errPackageNotFound = fmt.Errorf("package not found")

type resolverCache struct {
	cache   map[string]*File
	sources []Resolver
}

func newResolverCache(sources ...Resolver) *resolverCache {
	return &resolverCache{
		cache:   make(map[string]*File),
		sources: sources,
	}
}

func (rc *resolverCache) ListPackageFiles(pkgName string) ([]string, error) {
	for _, source := range rc.sources {
		files, err := source.ListPackageFiles(pkgName)
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

func (rc *resolverCache) FindFileByPath(filename string) (*File, error) {
	if res, ok := rc.cache[filename]; ok {
		return res, nil
	}

	for _, source := range rc.sources {
		file, err := source.FindFileByPath(filename)
		if err != nil {
			if err == ErrFileNotFound {
				continue // try next source
			}
			return nil, err
		}
		rc.cache[filename] = file
		return file, nil
	}

	return nil, ErrFileNotFound
}
