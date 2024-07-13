package j5reflect

import "fmt"

type WalkProperty struct {
	*ObjectProperty
	Path []string
}

type WalkCallback func(schema WalkProperty) error

func WalkSchemaFields(root RootSchema, callback WalkCallback) error {
	return walkSchemaFields(root, callback, nil)
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
			return fmt.Errorf("callback: %w", err)
		}

		switch st := prop.Schema.(type) {
		case *ObjectField:
			if err := walkSchemaFields(st.Ref.To, callback, propPath); err != nil {
				return fmt.Errorf("walk object as field: %w", err)
			}
		case *OneofField:
			if err := walkSchemaFields(st.Ref.To, callback, propPath); err != nil {
				return fmt.Errorf("walk oneof as field: %w", err)
			}
		}
	}

	return nil
}
