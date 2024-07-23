package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/internal/patherr"
)

type WalkProperty struct {
	*ObjectProperty
	Path []string
}

type WalkCallback func(schema WalkProperty) error

func WalkSchemaFields(root RootSchema, callback WalkCallback) error {
	err := walkSchemaFields(root, callback, nil)
	if err != nil {
		return err
	}
	return nil
}

func walkSchemaFields(root RootSchema, callback WalkCallback, path []string) error {

	var properties PropertySet
	switch rt := root.(type) {
	case *ObjectSchema:
		properties = rt.Properties
	case *OneofSchema:
		properties = rt.Properties
	case *EnumSchema:
		// do nothing
	default:
		return fmt.Errorf("unsupported schema type %T", root)
	}

	for _, prop := range properties {
		propPath := append(path, prop.JSONName)
		if err := callback(WalkProperty{
			ObjectProperty: prop,
			Path:           propPath,
		}); err != nil {
			return patherr.New(err, root.FullName(), prop.JSONName)
		}

		switch st := prop.Schema.(type) {
		case *ObjectField:
			if err := walkSchemaFields(st.Ref.To, callback, propPath); err != nil {
				return err // not wrapped, the path is already in the error above
			}
		case *OneofField:
			if err := walkSchemaFields(st.Ref.To, callback, propPath); err != nil {
				return err // not wrapped, the path is already in the error above
			}
		}
	}

	return nil
}
