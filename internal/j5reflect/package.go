package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/internal/patherr"
)

type PackageSet struct {
	Packages map[string]*Package
}

func NewPackageSet() *PackageSet {
	return &PackageSet{
		Packages: map[string]*Package{},
	}
}

func (ps *PackageSet) Package(name string) *Package {
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

type Package struct {
	Name       string
	Schemas    map[string]*RefSchema
	PackageSet *PackageSet
}

func (pkg *Package) assertAllRefsLink() error {

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
			return fmt.Errorf("unresolved reference %s", tt.FullName())
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

	for _, ref := range pkg.Schemas {
		if err := walkRootSchema(ref.To); err != nil {
			return patherr.Wrap(err, ref.FullName())
		}
	}

	return nil
}
