package j5schema

import (
	"fmt"

	"github.com/pentops/j5/lib/patherr"
)

type RootSet interface {
	refTo(pkg, schema string) (*RefSchema, bool)
	referencePackage(name string) *Package
}

type SchemaSet struct {
	*packageSet
}

func newSchemaSet() *SchemaSet {
	return &SchemaSet{
		packageSet: newPackageSet(),
	}
}

func (ss *SchemaSet) SchemaByName(packageName, name string) (RootSchema, error) {
	ref, ok := ss.getSchema(packageName, name)
	if !ok {
		return nil, fmt.Errorf("schema %q not found in package %q", name, packageName)
	}
	return ref.To, nil
}

func (ss *SchemaSet) IteratePackages(yield func(string, *Package) bool) {
	ss.packages.iterate(yield)
}

func (ss *SchemaSet) GetPackage(name string) (*Package, bool) {
	return ss.getPackage(name)
}

func (ps *SchemaSet) Package(name string) *Package {
	return ps.referencePackage(name)
}

type Package struct {
	Name       string
	Schemas    *cacheMap[RefSchema]
	PackageSet RootSet
}

func NewPackage(name string, set RootSet) *Package {
	return &Package{
		Name:       name,
		Schemas:    newCacheMap[RefSchema](),
		PackageSet: set,
	}
}

func (pkg *Package) IterateSchemas(yield func(string, *RefSchema) bool) {
	pkg.Schemas.iterate(yield)
}

func (pkg *Package) assertRefsLink() error {

	var seenSchemas = map[string]struct{}{}

	var walkRootSchema func(RootSchema) error
	var walkFieldSchema func(FieldSchema) error

	walkRootSchema = func(root RootSchema) error {
		name := root.FullName()
		if _, ok := seenSchemas[name]; ok {
			return nil
		}
		seenSchemas[name] = struct{}{}

		switch tt := root.(type) {
		case *ObjectSchema:
			for _, field := range tt.Properties {
				if err := walkFieldSchema(field.Schema); err != nil {
					return patherr.Wrap(err, field.JSONName)
				}
			}

		case *OneofSchema:
			for _, field := range tt.Properties {
				if err := walkFieldSchema(field.Schema); err != nil {
					return patherr.Wrap(err, field.JSONName)
				}
			}

		case *EnumSchema, *PolymorphSchema:
			// nothing to do, enum has no children with references

		default:
			return fmt.Errorf("unexpected schema type %T", root)
		}

		return nil
	}

	walkRef := func(tt *RefSchema) error {
		if tt.To == nil {
			return fmt.Errorf("unresolved reference to %s", tt.FullName())
		}
		if err := walkRootSchema(tt.To); err != nil {
			return err
		}
		return nil
	}

	walkFieldSchema = func(root FieldSchema) error {

		switch tt := root.(type) {

		case *ObjectField:
			return walkRef(tt.Ref)

		case *OneofField:
			return walkRef(tt.Ref)

		case *EnumField:
			return walkRef(tt.Ref)

		case *MapField:
			if err := walkFieldSchema(tt.ItemSchema); err != nil {
				return err
			}

		case *ArrayField:
			if err := walkFieldSchema(tt.ItemSchema); err != nil {
				return err
			}
		}

		return nil

	}

	for schemaName, ref := range pkg.Schemas.iterate {
		if err := walkRef(ref); err != nil {
			return patherr.Wrap(err, "schemas", schemaName)
		}
	}

	return nil
}
