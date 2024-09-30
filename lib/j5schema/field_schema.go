package j5schema

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/list/v1/list_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type ScalarSchema struct {
	fieldContext

	// subset of the available schema types, everything excluding ref, oneof
	// wrapper, array, object, map
	Proto *schema_j5pb.Field

	Kind              protoreflect.Kind
	WellKnownTypeName protoreflect.FullName
}

var _ FieldSchema = (*ScalarSchema)(nil)

func (s *ScalarSchema) ToJ5Field() *schema_j5pb.Field {
	return s.Proto
}

func (s *ScalarSchema) Mutable() bool {
	return false
}

func (s *ScalarSchema) AsContainer() (Container, bool) {
	return nil, false
}

func (s *ScalarSchema) TypeName() string {
	return baseTypeName(s.Proto.Type)
}

type AnyField struct {
	fieldContext

	OnlyDefined bool
	Types       []protoreflect.FullName
}

var _ FieldSchema = (*AnyField)(nil)

func (s *AnyField) ToJ5Field() *schema_j5pb.Field {
	return &schema_j5pb.Field{
		Type: &schema_j5pb.Field_Any{
			Any: &schema_j5pb.AnyField{
				OnlyDefined: s.OnlyDefined,
				Types:       stringSliceConvert[protoreflect.FullName, string](s.Types),
			},
		},
	}
}

func (s *AnyField) TypeName() string {
	return "any"
}

func (s *AnyField) AsContainer() (Container, bool) {
	return nil, false
}

func (s *AnyField) Mutable() bool {
	return true
}

type EnumField struct {
	fieldContext
	Ref       *RefSchema
	Rules     *schema_j5pb.EnumField_Rules
	ListRules *list_j5pb.EnumRules
	Ext       *schema_j5pb.EnumField_Ext
}

var _ FieldSchema = (*EnumField)(nil)

func (s *EnumField) Mutable() bool {
	return false
}

func (s *EnumField) AsContainer() (Container, bool) {
	return nil, false
}

func (s *EnumField) Schema() *EnumSchema {
	return s.Ref.To.(*EnumSchema)
}

func (s *EnumField) TypeName() string {
	return fmt.Sprintf("enum(%s)", s.Ref.FullName())
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
				Ext:       s.Ext,
			},
		},
	}
}

type ObjectField struct {
	fieldContext
	Ref     *RefSchema
	Flatten bool
	Rules   *schema_j5pb.ObjectField_Rules
	Ext     *schema_j5pb.ObjectField_Ext
}

var _ FieldSchema = (*ObjectField)(nil)

func (s *ObjectField) TypeName() string {
	return fmt.Sprintf("object(%s)", s.Ref.FullName())
}

func (s *ObjectField) AsContainer() (Container, bool) {
	return s.Ref.To.(*ObjectSchema).Properties, true
}

func (s *ObjectField) Mutable() bool {
	return true
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
				Flatten: s.Flatten,
				Rules:   s.Rules,
				Ext:     s.Ext,
			},
		},
	}
}

type OneofField struct {
	fieldContext
	Ref       *RefSchema
	Rules     *schema_j5pb.OneofField_Rules
	ListRules *list_j5pb.OneofRules
	Ext       *schema_j5pb.OneofField_Ext
}

var _ FieldSchema = (*OneofField)(nil)

func (s *OneofField) Mutable() bool {
	return true
}

func (s *OneofField) AsContainer() (Container, bool) {
	return s.Ref.To.(*OneofSchema).Properties, true
}

func (s *OneofField) TypeName() string {
	return fmt.Sprintf("oneof(%s)", s.Ref.FullName())
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
				Ext:       s.Ext,
			},
		},
	}
}

type MapField struct {
	fieldContext
	Schema FieldSchema
	Rules  *schema_j5pb.MapField_Rules
	Ext    *schema_j5pb.MapField_Ext
}

var _ FieldSchema = (*MapField)(nil)

func (s *MapField) AsContainer() (Container, bool) {
	return nil, false
}

func (s *MapField) Mutable() bool {
	return true
}

func (s *MapField) TypeName() string {
	return fmt.Sprintf("map(string,%s)", s.Schema.TypeName())
}

func (s *MapField) ToJ5Field() *schema_j5pb.Field {
	item := s.Schema.ToJ5Field()

	return &schema_j5pb.Field{
		Type: &schema_j5pb.Field_Map{
			Map: &schema_j5pb.MapField{
				ItemSchema: item,
				KeySchema:  &schema_j5pb.Field{Type: &schema_j5pb.Field_String_{}},
				Rules:      s.Rules,
				Ext:        s.Ext,
			},
		},
	}
}

type ArrayField struct {
	fieldContext
	Schema FieldSchema
	Rules  *schema_j5pb.ArrayField_Rules
	Ext    *schema_j5pb.ArrayField_Ext
}

var _ FieldSchema = (*ArrayField)(nil)

func (s *ArrayField) Mutable() bool {
	return true
}

func (s *ArrayField) AsContainer() (Container, bool) {
	return nil, false
}

func (s *ArrayField) TypeName() string {
	return fmt.Sprintf("array(%s)", s.Schema.TypeName())
}

func (s *ArrayField) ToJ5Field() *schema_j5pb.Field {
	item := s.Schema.ToJ5Field()
	return &schema_j5pb.Field{
		Type: &schema_j5pb.Field_Array{
			Array: &schema_j5pb.ArrayField{
				Items: item,
				Rules: s.Rules,
				Ext:   s.Ext,
			},
		},
	}
}
