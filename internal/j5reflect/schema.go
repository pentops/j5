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

func (s *SchemaRoot) PackageName() string {
	return s.Package
}

func (s *SchemaRoot) AsRef() *RefSchema {
	return &RefSchema{
		Package: s.Package,
		Schema:  s.Name,
	}
}

type FieldSchema interface {
	ToJ5Field() (*schema_j5pb.Schema, error)
}

type RootSchema interface {
	AsRef() *RefSchema
	FullName() string
	PackageName() string
	ToJ5Root() (*schema_j5pb.RootSchema, error)
}

type RefSchema struct {
	Package string
	Schema  string
	To      RootSchema
}

func (s *RefSchema) FullName() string {
	return fmt.Sprintf("%s.%s", s.Package, s.Schema)
}

type ScalarSchema struct {
	// subset of the available schema types, everything excluding ref, oneof
	// wrapper, array, object, map
	Proto *schema_j5pb.Schema

	Kind              protoreflect.Kind
	WellKnownTypeName protoreflect.FullName
}

func (s *ScalarSchema) ToJ5Field() (*schema_j5pb.Schema, error) {
	return s.Proto, nil
}

type AnySchema struct {
	Description *string
}

func (s *AnySchema) ToJ5Field() (*schema_j5pb.Schema, error) {
	return &schema_j5pb.Schema{
		Type: &schema_j5pb.Schema_Any{
			Any: &schema_j5pb.Any{},
		},
	}, nil
}

type EnumAsFieldSchema struct {
	Ref   *RefSchema
	Rules *schema_j5pb.EnumAsField_Rules
}

func (s *EnumAsFieldSchema) Schema() *EnumSchema {
	return s.Ref.To.(*EnumSchema)
}

func (s *EnumAsFieldSchema) ToJ5Field() (*schema_j5pb.Schema, error) {
	return &schema_j5pb.Schema{
		Type: &schema_j5pb.Schema_Enum{
			Enum: &schema_j5pb.EnumAsField{
				Schema: &schema_j5pb.EnumAsField_Ref{
					Ref: &schema_j5pb.Ref{
						Package: s.Ref.Package,
						Schema:  s.Ref.Schema,
					},
				},
				Rules: s.Rules,
			},
		},
	}, nil
}

type EnumSchema struct {
	SchemaRoot

	NamePrefix string
	Options    []*schema_j5pb.Enum_Value
}

func (s *EnumSchema) ToJ5Root() (*schema_j5pb.RootSchema, error) {
	return &schema_j5pb.RootSchema{
		Type: &schema_j5pb.RootSchema_Enum{
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
}

func (s *ObjectSchema) ToJ5Root() (*schema_j5pb.RootSchema, error) {
	properties := make([]*schema_j5pb.ObjectProperty, 0, len(s.Properties))
	for _, prop := range s.Properties {
		property, err := prop.ToJ5Proto()
		if err != nil {
			return nil, fmt.Errorf("property %s: %w", prop.JSONName, err)
		}
		properties = append(properties, property)
	}
	return &schema_j5pb.RootSchema{
		Type: &schema_j5pb.RootSchema_Object{
			Object: &schema_j5pb.Object{
				Description: s.Description,
				Name:        s.Name,
				Properties:  properties,
			},
		},
	}, nil
}

type ObjectAsFieldSchema struct {
	Ref   *RefSchema
	Rules *schema_j5pb.ObjectAsField_Rules
}

func (s *ObjectAsFieldSchema) Schema() *ObjectSchema {
	return s.Ref.To.(*ObjectSchema)
}

func (s *ObjectAsFieldSchema) ToJ5Field() (*schema_j5pb.Schema, error) {
	return &schema_j5pb.Schema{
		Type: &schema_j5pb.Schema_Object{
			Object: &schema_j5pb.ObjectAsField{
				Schema: &schema_j5pb.ObjectAsField_Ref{
					Ref: &schema_j5pb.Ref{
						Package: s.Ref.Package,
						Schema:  s.Ref.Schema,
					},
				},
				Rules: s.Rules,
			},
		},
	}, nil
}

type OneofSchema struct {
	SchemaRoot
	Properties PropertySet
}

func (s *OneofSchema) ToJ5Root() (*schema_j5pb.RootSchema, error) {
	properties := make([]*schema_j5pb.ObjectProperty, 0, len(s.Properties))
	for _, prop := range s.Properties {
		property, err := prop.ToJ5Proto()
		if err != nil {
			return nil, fmt.Errorf("property %s: %w", prop.JSONName, err)
		}
		properties = append(properties, property)
	}

	return &schema_j5pb.RootSchema{
		Type: &schema_j5pb.RootSchema_Oneof{
			Oneof: &schema_j5pb.Oneof{
				Description: s.Description,
				Name:        s.Name,
				Properties:  properties,
			},
		},
	}, nil
}

type OneofAsFieldSchema struct {
	Ref   *RefSchema
	Rules *schema_j5pb.OneofAsField_Rules
}

func (s *OneofAsFieldSchema) Schema() *OneofSchema {
	return s.Ref.To.(*OneofSchema)
}

func (s *OneofAsFieldSchema) ToJ5Field() (*schema_j5pb.Schema, error) {
	return &schema_j5pb.Schema{
		Type: &schema_j5pb.Schema_Oneof{
			Oneof: &schema_j5pb.OneofAsField{
				Schema: &schema_j5pb.OneofAsField_Ref{
					Ref: &schema_j5pb.Ref{
						Package: s.Ref.Package,
						Schema:  s.Ref.Schema,
					},
				},
				Rules: s.Rules,
			},
		},
	}, nil
}

type ObjectProperty struct {
	Schema FieldSchema

	ProtoField []protoreflect.FieldNumber

	JSONName string

	Required           bool
	ReadOnly           bool
	WriteOnly          bool
	ExplicitlyOptional bool

	Description string
}

func (prop *ObjectProperty) ToJ5Proto() (*schema_j5pb.ObjectProperty, error) {
	proto, err := prop.Schema.ToJ5Field()
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
	Schema FieldSchema
	Rules  *schema_j5pb.Map_Rules
}

func (s *MapSchema) ToJ5Field() (*schema_j5pb.Schema, error) {
	item, err := s.Schema.ToJ5Field()
	if err != nil {
		return nil, fmt.Errorf("map item: %w", err)
	}

	return &schema_j5pb.Schema{
		Type: &schema_j5pb.Schema_Map{
			Map: &schema_j5pb.Map{
				ItemSchema: item,
				KeySchema:  &schema_j5pb.Schema{Type: &schema_j5pb.Schema_String_{}},
				Rules:      s.Rules,
			},
		},
	}, nil
}

type ArraySchema struct {
	Schema FieldSchema
	Rules  *schema_j5pb.Array_Rules
}

func (s *ArraySchema) ToJ5Field() (*schema_j5pb.Schema, error) {

	item, err := s.Schema.ToJ5Field()
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
