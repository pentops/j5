package j5schema

import (
	"fmt"
	"strings"
	"sync"
)

type packageSet struct {
	packages *cacheMap[Package]
}

func newPackageSet() *packageSet {
	return &packageSet{
		packages: newCacheMap[Package](),
	}
}

func (ps *packageSet) SchemaByName(name string) (RootSchema, error) {
	parts := strings.SplitN(name, ".", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid schema name %q", name)
	}
	packageName, nameInPackage := strings.Join(parts[0:len(parts)-1], "."), parts[len(parts)-1]
	pkg, ok := ps.getPackage(packageName)
	if !ok {
		return nil, fmt.Errorf("package %q not found", packageName)
	}
	if schema, ok := pkg.Schemas.get(nameInPackage); ok {
		return schema.To, nil
	}
	return nil, fmt.Errorf("schema %q not found in package %q", nameInPackage, packageName)
}

func (ps *packageSet) refTo(pkg, schema string) (*RefSchema, bool) {
	refPackage := ps.referencePackage(pkg)
	return refPackage.Schemas.getOrCreate(schema, func() *RefSchema {
		return &RefSchema{
			Package: refPackage,
			Schema:  schema,
		}
	})
}

func (ps *packageSet) getPackage(name string) (*Package, bool) {
	return ps.packages.get(name)
}

func (ps *packageSet) referencePackage(name string) *Package {
	pkg, _ := ps.packages.getOrCreate(name, func() *Package {
		return NewPackage(name, ps)
	})
	return pkg
}

func (ps *packageSet) getSchema(packageName, schemaName string) (*RefSchema, bool) {
	pkg, ok := ps.getPackage(packageName)
	if !ok {
		return nil, false
	}
	return pkg.Schemas.get(schemaName)
}

type cacheMap[T any] struct {
	store map[string]*T
	lock  sync.RWMutex
}

func newCacheMap[T any]() *cacheMap[T] {
	return &cacheMap[T]{
		store: make(map[string]*T),
	}
}

func (cm *cacheMap[T]) get(key string) (*T, bool) {
	cm.lock.RLock()
	defer cm.lock.RUnlock()
	value, ok := cm.store[key]
	return value, ok
}

func (cm *cacheMap[T]) set(key string, value *T) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	cm.store[key] = value
}

func (cm *cacheMap[T]) getOrCreate(key string, constructor func() *T) (*T, bool) {
	existing, ok := cm.get(key)
	if ok {
		return existing, true
	}
	cm.lock.Lock()
	defer cm.lock.Unlock()
	// may be created between locks
	if existing, ok := cm.store[key]; ok {
		return existing, true
	}
	value := constructor()
	cm.store[key] = value
	return value, false
}

func (cm *cacheMap[T]) iterate(fn func(key string, value *T) bool) {
	cm.lock.RLock()
	defer cm.lock.RUnlock()
	for key, value := range cm.store {
		if !fn(key, value) {
			break
		}
	}
}
