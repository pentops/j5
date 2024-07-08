package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type SchemaRoot struct {
	Description string
	Package     string
	Name        string
}

func (s *SchemaRoot) FullName() string {
	return fmt.Sprintf("%s.%s", s.Package, s.Name)
}

func (s *SchemaRoot) AsRef() *RefSchema {
	return &RefSchema{
		Package: s.Package,
		Schema:  s.Name,
	}
}

type Schema interface {
	ToJ5Proto() (*schema_j5pb.Schema, error)
}

type RootSchema interface {
	Schema
	AsRef() *RefSchema
	FullName() string
}

type RefSchema struct {
	Package string
	Schema  string
	To      RootSchema
}

func (s *RefSchema) FullName() string {
	return fmt.Sprintf("%s.%s", s.Package, s.Schema)
}

//func (s *RefSchema) AsRef() *RefSchema {
//	return s
//}

func (s *RefSchema) ToJ5Proto() (*schema_j5pb.Schema, error) {
	if s == nil {
		return nil, fmt.Errorf("Nil Ref!")
	}
	if s.Schema == "" {
		if s.To == nil {
			return nil, fmt.Errorf("no schema or To link for ref")
		}
		switch t := s.To.(type) {
		case *ObjectSchema:
			s.Schema = t.Name
			s.Package = t.Package
		case *OneofSchema:
			s.Schema = t.Name
			s.Package = t.Package
		case *EnumSchema:
			s.Schema = t.Name
			s.Package = t.Package
		default:
			return nil, fmt.Errorf("unsupported ref type %T", t)
		}
	}
	return &schema_j5pb.Schema{
		Type: &schema_j5pb.Schema_Ref{
			Ref: &schema_j5pb.Ref{
				Package: s.Package,
				Schema:  s.Schema,
			},
		},
	}, nil
}

type ScalarSchema struct {
	// subset of the available schema types, everything excluding ref, oneof
	// wrapper, array, object, map
	Proto *schema_j5pb.Schema

	Kind              protoreflect.Kind
	WellKnownTypeName protoreflect.FullName
}

func (s *ScalarSchema) ToJ5Proto() (*schema_j5pb.Schema, error) {
	return s.Proto, nil
}

type AnySchema struct {
	Description *string
}

func (s *AnySchema) ToJ5Proto() (*schema_j5pb.Schema, error) {
	return &schema_j5pb.Schema{
		Type: &schema_j5pb.Schema_Any{
			Any: &schema_j5pb.Any{},
		},
	}, nil
}

type EnumSchema struct {
	SchemaRoot

	NamePrefix string
	Options    []*schema_j5pb.Enum_Value
}

func (s *EnumSchema) ToJ5Proto() (*schema_j5pb.Schema, error) {
	return &schema_j5pb.Schema{
		Type: &schema_j5pb.Schema_Enum{
			Enum: &schema_j5pb.Enum{
				Name:        s.Name,
				Description: s.Description,
				Options:     s.Options,
				Prefix:      s.NamePrefix,
			},
		},
	}, nil
}

type PropertySet []*ObjectProperty

func (ps PropertySet) ByJSONName(name string) *ObjectProperty {
	for _, prop := range ps {
		if prop.JSONName == name {
			return prop
		}
	}
	return nil
}

type ObjectSchema struct {
	SchemaRoot
	Properties PropertySet
	Rules      *schema_j5pb.Object_Rules
}

func (s *ObjectSchema) ToJ5Proto() (*schema_j5pb.Schema, error) {
	properties := make([]*schema_j5pb.ObjectProperty, 0, len(s.Properties))
	for _, prop := range s.Properties {
		property, err := prop.ToJ5Proto()
		if err != nil {
			return nil, fmt.Errorf("property %s: %w", prop.JSONName, err)
		}
		properties = append(properties, property)
	}
	return &schema_j5pb.Schema{
		Type: &schema_j5pb.Schema_Object{
			Object: &schema_j5pb.Object{
				Description: s.Description,
				Name:        s.Name,
				Properties:  properties,
				Rules:       s.Rules,
			},
		},
	}, nil
}

type OneofSchema struct {
	SchemaRoot
	Properties PropertySet
	Rules      *schema_j5pb.Oneof_Rules
}

func (s *OneofSchema) ToJ5Proto() (*schema_j5pb.Schema, error) {
	properties := make([]*schema_j5pb.ObjectProperty, 0, len(s.Properties))
	for _, prop := range s.Properties {
		property, err := prop.ToJ5Proto()
		if err != nil {
			return nil, fmt.Errorf("property %s: %w", prop.JSONName, err)
		}
		properties = append(properties, property)
	}

	return &schema_j5pb.Schema{
		Type: &schema_j5pb.Schema_Oneof{
			Oneof: &schema_j5pb.Oneof{
				Description: s.Description,
				Name:        s.Name,
				Properties:  properties,
				Rules:       s.Rules,
			},
		},
	}, nil
}

type ObjectProperty struct {
	Schema Schema

	ProtoField []protoreflect.FieldNumber

	JSONName string

	Required           bool
	ReadOnly           bool
	WriteOnly          bool
	ExplicitlyOptional bool

	Description string
}

func (prop *ObjectProperty) ToJ5Proto() (*schema_j5pb.ObjectProperty, error) {
	proto, err := prop.Schema.ToJ5Proto()
	if err != nil {
		return nil, fmt.Errorf("property %s: %w", prop.JSONName, err)
	}
	fieldPath := make([]int32, len(prop.ProtoField))
	for idx, field := range prop.ProtoField {
		fieldPath[idx] = int32(field)
	}
	return &schema_j5pb.ObjectProperty{
		Schema:             proto,
		Name:               prop.JSONName,
		Required:           prop.Required,
		ExplicitlyOptional: prop.ExplicitlyOptional,
		ReadOnly:           prop.ReadOnly,
		WriteOnly:          prop.WriteOnly,
		Description:        prop.Description,
		ProtoField:         fieldPath,
	}, nil

}

type MapSchema struct {
	Schema Schema
	Rules  *schema_j5pb.Map_Rules
}

func (s *MapSchema) ToJ5Proto() (*schema_j5pb.Schema, error) {
	item, err := s.Schema.ToJ5Proto()
	if err != nil {
		return nil, fmt.Errorf("map item: %w", err)
	}

	return &schema_j5pb.Schema{
		Type: &schema_j5pb.Schema_Map{
			Map: &schema_j5pb.Map{
				ItemSchema: item,
				Rules:      s.Rules,
			},
		},
	}, nil
}

type ArraySchema struct {
	Schema Schema
	Rules  *schema_j5pb.Array_Rules
}

func (s *ArraySchema) ToJ5Proto() (*schema_j5pb.Schema, error) {

	item, err := s.Schema.ToJ5Proto()
	if err != nil {
		return nil, fmt.Errorf("array item: %w", err)
	}

	return &schema_j5pb.Schema{
		Type: &schema_j5pb.Schema_Array{
			Array: &schema_j5pb.Array{
				Items: item,
				Rules: s.Rules,
			},
		},
	}, nil
}
