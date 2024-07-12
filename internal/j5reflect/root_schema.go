package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type RootSchema interface {
	AsRef() *RefSchema
	FullName() string
	PackageName() string
	ToJ5Root() *schema_j5pb.RootSchema
}

type RefSchema struct {
	Package string
	Schema  string
	To      RootSchema
}

func (s *RefSchema) FullName() string {
	return fmt.Sprintf("%s.%s", s.Package, s.Schema)
}

type rootSchema struct {
	Description string
	Package     string
	Name        string
}

func (s *rootSchema) FullName() string {
	return fmt.Sprintf("%s.%s", s.Package, s.Name)
}

func (s *rootSchema) PackageName() string {
	return s.Package
}

func (s *rootSchema) AsRef() *RefSchema {
	return &RefSchema{
		Package: s.Package,
		Schema:  s.Name,
	}
}

type EnumSchema struct {
	rootSchema

	NamePrefix string
	Options    []*schema_j5pb.Enum_Value
}

func (s *EnumSchema) ToJ5Root() *schema_j5pb.RootSchema {
	return &schema_j5pb.RootSchema{
		Type: &schema_j5pb.RootSchema_Enum{
			Enum: &schema_j5pb.Enum{
				Name:        s.Name,
				Description: s.Description,
				Options:     s.Options,
				Prefix:      s.NamePrefix,
			},
		},
	}
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
	rootSchema
	Properties PropertySet
}

func (s *ObjectSchema) ToJ5Object() *schema_j5pb.Object {
	properties := make([]*schema_j5pb.ObjectProperty, 0, len(s.Properties))
	for _, prop := range s.Properties {
		property := prop.ToJ5Proto()
		properties = append(properties, property)
	}
	return &schema_j5pb.Object{
		Description: s.Description,
		Name:        s.Name,
		Properties:  properties,
	}
}

func (s *ObjectSchema) ToJ5Root() *schema_j5pb.RootSchema {
	built := s.ToJ5Object()
	return &schema_j5pb.RootSchema{
		Type: &schema_j5pb.RootSchema_Object{
			Object: built,
		},
	}
}

type OneofSchema struct {
	rootSchema
	Properties PropertySet
}

func (s *OneofSchema) ToJ5Root() *schema_j5pb.RootSchema {
	properties := make([]*schema_j5pb.ObjectProperty, 0, len(s.Properties))
	for _, prop := range s.Properties {
		property := prop.ToJ5Proto()
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
	}
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

func (prop *ObjectProperty) ToJ5Proto() *schema_j5pb.ObjectProperty {
	fieldPath := make([]int32, len(prop.ProtoField))
	for idx, field := range prop.ProtoField {
		fieldPath[idx] = int32(field)
	}
	return &schema_j5pb.ObjectProperty{
		Schema:             prop.Schema.ToJ5Field(),
		Name:               prop.JSONName,
		Required:           prop.Required,
		ExplicitlyOptional: prop.ExplicitlyOptional,
		ReadOnly:           prop.ReadOnly,
		WriteOnly:          prop.WriteOnly,
		Description:        prop.Description,
		ProtoField:         fieldPath,
	}

}
