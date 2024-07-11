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
			Any: &schema_j5pb.AnyField{},
		},
	}, nil
}

type EnumFieldSchema struct {
	Ref   *RefSchema
	Rules *schema_j5pb.EnumField_Rules
}

func (s *EnumFieldSchema) Schema() *EnumSchema {
	return s.Ref.To.(*EnumSchema)
}

func (s *EnumFieldSchema) ToJ5Field() (*schema_j5pb.Schema, error) {
	return &schema_j5pb.Schema{
		Type: &schema_j5pb.Schema_Enum{
			Enum: &schema_j5pb.EnumField{
				Schema: &schema_j5pb.EnumField_Ref{
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

type ObjectFieldSchema struct {
	Ref   *RefSchema
	Rules *schema_j5pb.ObjectField_Rules
}

func (s *ObjectFieldSchema) Schema() *ObjectSchema {
	return s.Ref.To.(*ObjectSchema)
}

func (s *ObjectFieldSchema) ToJ5Field() (*schema_j5pb.Schema, error) {
	return &schema_j5pb.Schema{
		Type: &schema_j5pb.Schema_Object{
			Object: &schema_j5pb.ObjectField{
				Schema: &schema_j5pb.ObjectField_Ref{
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

type OneofFieldSchema struct {
	Ref   *RefSchema
	Rules *schema_j5pb.OneofField_Rules
}

func (s *OneofFieldSchema) Schema() *OneofSchema {
	return s.Ref.To.(*OneofSchema)
}

func (s *OneofFieldSchema) ToJ5Field() (*schema_j5pb.Schema, error) {
	return &schema_j5pb.Schema{
		Type: &schema_j5pb.Schema_Oneof{
			Oneof: &schema_j5pb.OneofField{
				Schema: &schema_j5pb.OneofField_Ref{
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

type MapSchema struct {
	Schema FieldSchema
	Rules  *schema_j5pb.MapField_Rules
}

func (s *MapSchema) ToJ5Field() (*schema_j5pb.Schema, error) {
	item, err := s.Schema.ToJ5Field()
	if err != nil {
		return nil, fmt.Errorf("map item: %w", err)
	}

	return &schema_j5pb.Schema{
		Type: &schema_j5pb.Schema_Map{
			Map: &schema_j5pb.MapField{
				ItemSchema: item,
				KeySchema:  &schema_j5pb.Schema{Type: &schema_j5pb.Schema_String_{}},
				Rules:      s.Rules,
			},
		},
	}, nil
}

type ArraySchema struct {
	Schema FieldSchema
	Rules  *schema_j5pb.ArrayField_Rules
}

func (s *ArraySchema) ToJ5Field() (*schema_j5pb.Schema, error) {

	item, err := s.Schema.ToJ5Field()
	if err != nil {
		return nil, fmt.Errorf("array item: %w", err)
	}

	return &schema_j5pb.Schema{
		Type: &schema_j5pb.Schema_Array{
			Array: &schema_j5pb.ArrayField{
				Items: item,
				Rules: s.Rules,
			},
		},
	}, nil
}
