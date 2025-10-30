package j5schema

import (
	"fmt"
	"maps"
	"strings"

	"github.com/pentops/j5/lib/patherr"
)

type WalkProperty struct {
	*ObjectProperty
	Path []string
}

type WalkCallback func(schema WalkProperty) error

func WalkSchemaFields(root RootSchema, asClient bool, callback WalkCallback) error {
	err := walkSchemaFields(root, asClient, callback, nil, map[string]struct{}{})
	if err != nil {
		return fmt.Errorf("walk from schema %s: %w", root.FullName(), err)
	}
	return nil
}

func walkSchemaFields(root RootSchema, asClient bool, callback WalkCallback, path []string, seen map[string]struct{}) error {
	subSeen := maps.Clone(seen)
	fullName := root.FullName()
	if _, ok := subSeen[fullName]; ok {
		// We've already seen this schema in this walk,
		return fmt.Errorf("%s circular reference (%s)", strings.Join(path, "."), fullName)
	}
	subSeen[fullName] = struct{}{}

	var properties PropertySet
	switch rt := root.(type) {
	case *ObjectSchema:
		if asClient {
			properties = rt.ClientProperties()
		} else {
			properties = rt.Properties
		}
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
			if err := walkSchemaFields(st.Ref.To, asClient, callback, propPath, maps.Clone(subSeen)); err != nil {
				return err // not wrapped, the path is already in the error above
			}
		case *OneofField:
			if err := walkSchemaFields(st.Ref.To, asClient, callback, propPath, maps.Clone(subSeen)); err != nil {
				return err // not wrapped, the path is already in the error above
			}
		}
	}

	return nil
}
