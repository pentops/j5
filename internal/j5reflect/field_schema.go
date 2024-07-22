package j5reflect

import (
	"github.com/pentops/j5/gen/j5/list/v1/list_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type FieldSchema interface {
	ToJ5Field() *schema_j5pb.Field
}

type ScalarSchema struct {
	// subset of the available schema types, everything excluding ref, oneof
	// wrapper, array, object, map
	Proto *schema_j5pb.Field

	Kind              protoreflect.Kind
	WellKnownTypeName protoreflect.FullName
}

func (s *ScalarSchema) ToJ5Field() *schema_j5pb.Field {
	return s.Proto
}

type AnyField struct {
	Description *string
}

func (s *AnyField) ToJ5Field() *schema_j5pb.Field {
	return &schema_j5pb.Field{
		Type: &schema_j5pb.Field_Any{
			Any: &schema_j5pb.AnyField{},
		},
	}
}

type EnumField struct {
	Ref       *RefSchema
	Rules     *schema_j5pb.EnumField_Rules
	ListRules *list_j5pb.EnumRules
}

func (s *EnumField) Schema() *EnumSchema {
	return s.Ref.To.(*EnumSchema)
}

func (s *EnumField) ToJ5Field() *schema_j5pb.Field {
	return &schema_j5pb.Field{
		Type: &schema_j5pb.Field_Enum{
			Enum: &schema_j5pb.EnumField{
				Schema: &schema_j5pb.EnumField_Ref{
					Ref: &schema_j5pb.Ref{
						Package: s.Ref.Package.Name,
						Schema:  s.Ref.Schema,
					},
				},
				Rules:     s.Rules,
				ListRules: s.ListRules,
			},
		},
	}
}

type ObjectField struct {
	Ref   *RefSchema
	Rules *schema_j5pb.ObjectField_Rules
}

func (s *ObjectField) Schema() *ObjectSchema {
	return s.Ref.To.(*ObjectSchema)
}

func (s *ObjectField) ToJ5Field() *schema_j5pb.Field {
	return &schema_j5pb.Field{
		Type: &schema_j5pb.Field_Object{
			Object: &schema_j5pb.ObjectField{
				Schema: &schema_j5pb.ObjectField_Ref{
					Ref: &schema_j5pb.Ref{
						Package: s.Ref.Package.Name,
						Schema:  s.Ref.Schema,
					},
				},
				Rules: s.Rules,
			},
		},
	}
}

type OneofField struct {
	Ref       *RefSchema
	Rules     *schema_j5pb.OneofField_Rules
	ListRules *list_j5pb.OneofRules
}

func (s *OneofField) Schema() *OneofSchema {
	return s.Ref.To.(*OneofSchema)
}

func (s *OneofField) ToJ5Field() *schema_j5pb.Field {
	return &schema_j5pb.Field{
		Type: &schema_j5pb.Field_Oneof{
			Oneof: &schema_j5pb.OneofField{
				Schema: &schema_j5pb.OneofField_Ref{
					Ref: &schema_j5pb.Ref{
						Package: s.Ref.Package.Name,
						Schema:  s.Ref.Schema,
					},
				},
				Rules:     s.Rules,
				ListRules: s.ListRules,
			},
		},
	}
}

type MapField struct {
	Schema FieldSchema
	Rules  *schema_j5pb.MapField_Rules
}

func (s *MapField) ToJ5Field() *schema_j5pb.Field {
	item := s.Schema.ToJ5Field()

	return &schema_j5pb.Field{
		Type: &schema_j5pb.Field_Map{
			Map: &schema_j5pb.MapField{
				ItemSchema: item,
				KeySchema:  &schema_j5pb.Field{Type: &schema_j5pb.Field_String_{}},
				Rules:      s.Rules,
			},
		},
	}
}

type ArrayField struct {
	Schema FieldSchema
	Rules  *schema_j5pb.ArrayField_Rules
}

func (s *ArrayField) ToJ5Field() *schema_j5pb.Field {
	item := s.Schema.ToJ5Field()
	return &schema_j5pb.Field{
		Type: &schema_j5pb.Field_Array{
			Array: &schema_j5pb.ArrayField{
				Items: item,
				Rules: s.Rules,
			},
		},
	}
}
