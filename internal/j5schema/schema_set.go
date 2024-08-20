package j5schema

import (
	"fmt"

	"github.com/pentops/j5/internal/patherr"
)

type RootSet interface {
	refTo(pkg, schema string) (*RefSchema, bool)
	referencePackage(name string) *Package
}

type SchemaSet struct {
	Packages map[string]*Package
}

func newSchemaSet() *SchemaSet {
	return &SchemaSet{
		Packages: map[string]*Package{},
	}
}

func (ss *SchemaSet) SchemaByName(packageName, name string) (RootSchema, error) {
	pkg, ok := ss.Packages[packageName]
	if !ok {
		return nil, fmt.Errorf("package %q not found", packageName)
	}
	ref, ok := pkg.Schemas[name]
	if !ok {
		return nil, fmt.Errorf("schema %q not found in package %q", name, packageName)
	}
	return ref.To, nil
}

func (ss *SchemaSet) refTo(pkg, schema string) (*RefSchema, bool) {
	refPackage := ss.Package(pkg)
	if existing, ok := refPackage.Schemas[schema]; ok {
		return existing, true
	}

	refSchema := &RefSchema{
		Package: refPackage,
		Schema:  schema,
	}
	refPackage.Schemas[schema] = refSchema

	return refSchema, false
}

func (ps *SchemaSet) referencePackage(name string) *Package {
	if pkg, ok := ps.Packages[name]; ok {
		return pkg
	}
	pkg := &Package{
		Name:       name,
		PackageSet: ps,
		Schemas:    map[string]*RefSchema{},
	}
	ps.Packages[name] = pkg
	return pkg
}

func (ps *SchemaSet) Package(name string) *Package {
	return ps.referencePackage(name)
}

type Package struct {
	Name       string
	Schemas    map[string]*RefSchema
	PackageSet RootSet
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

		case *EnumSchema:
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
			if err := walkFieldSchema(tt.Schema); err != nil {
				return err
			}

		case *ArrayField:
			if err := walkFieldSchema(tt.Schema); err != nil {
				return err
			}
		}

		return nil

	}

	for schemaName, ref := range pkg.Schemas {
		if err := walkRef(ref); err != nil {
			return patherr.Wrap(err, "schemas", schemaName)
		}
	}

	return nil
}
