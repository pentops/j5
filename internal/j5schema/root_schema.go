package j5schema

import (
	"fmt"
	"strings"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type RootSchema interface {
	AsRef() *RefSchema
	FullName() string
	Name() string
	PackageName() string
	Package() *Package
	ToJ5Root() *schema_j5pb.RootSchema
	ToJ5ClientRoot() *schema_j5pb.RootSchema
	Description() string
}

type RefSchema struct {
	Package *Package
	Schema  string
	To      RootSchema
}

func (s *RefSchema) FullName() string {
	return fmt.Sprintf("%s.%s", s.Package.Name, s.Schema)
}

type rootSchema struct {
	description string
	pkg         *Package
	name        string
}

func (s *rootSchema) FullName() string {
	return fmt.Sprintf("%s.%s", s.pkg.Name, s.name)
}

func (s *rootSchema) Name() string {
	if s == nil {
		return "<nil>"
	}
	return s.name
}

func (s *rootSchema) PackageName() string {
	return s.pkg.Name
}

func (s *rootSchema) Package() *Package {
	return s.pkg
}

func (s *rootSchema) AsRef() *RefSchema {
	return &RefSchema{
		Package: s.pkg,
		Schema:  s.name,
	}
}

func (s *rootSchema) Description() string {
	return s.description
}

type EnumSchema struct {
	rootSchema

	NamePrefix string
	Options    []*schema_j5pb.Enum_Value
}

func (s *EnumSchema) OptionByName(name string) *schema_j5pb.Enum_Value {
	shortName := strings.TrimPrefix(name, s.NamePrefix)
	for _, opt := range s.Options {
		if opt.Name == shortName {
			return opt
		}
	}
	return nil
}

func (s *EnumSchema) OptionByNumber(num int32) *schema_j5pb.Enum_Value {
	for _, opt := range s.Options {
		if opt.Number == num {
			return opt
		}
	}
	return nil
}

func (s *EnumSchema) ToJ5Root() *schema_j5pb.RootSchema {
	return &schema_j5pb.RootSchema{
		Type: &schema_j5pb.RootSchema_Enum{
			Enum: &schema_j5pb.Enum{
				Name:        s.name,
				Description: s.description,
				Options:     s.Options,
				Prefix:      s.NamePrefix,
			},
		},
	}
}

func (s *EnumSchema) ToJ5ClientRoot() *schema_j5pb.RootSchema {
	return s.ToJ5Root()
}

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
	Entity     *schema_j5pb.EntityObject
	Properties PropertySet
}

func (s *ObjectSchema) Clone() *ObjectSchema {
	properties := make(PropertySet, len(s.Properties))
	copy(properties, s.Properties)
	return &ObjectSchema{
		rootSchema: rootSchema{
			description: s.description,
			pkg:         s.pkg,
			name:        s.name,
		},
		Properties: properties,
	}
}

func (s *ObjectSchema) ToJ5Object() *schema_j5pb.Object {
	properties := make([]*schema_j5pb.ObjectProperty, 0, len(s.Properties))
	for _, prop := range s.Properties {
		property := prop.ToJ5Proto()
		properties = append(properties, property)
	}
	return &schema_j5pb.Object{
		Description: s.description,
		Name:        s.name,
		Properties:  properties,
		Entity:      s.Entity,
	}
}

func (s *ObjectSchema) ClientProperties() []*ObjectProperty {
	properties := make([]*ObjectProperty, 0, len(s.Properties))
	for _, prop := range s.Properties {
		switch propType := prop.Schema.(type) {
		case *ObjectField:
			if propType.Flatten {

				children := propType.Schema().ClientProperties()
				for _, child := range children {
					child := child.nestedClone(prop.ProtoField)
					properties = append(properties, child)
				}

				continue
			}
		}
		properties = append(properties, prop)
	}
	return properties
}

func (s *ObjectSchema) ToJ5ClientRoot() *schema_j5pb.RootSchema {
	clientProperties := s.ClientProperties()
	properties := make([]*schema_j5pb.ObjectProperty, 0, len(clientProperties))
	for _, prop := range s.Properties {
		property := prop.ToJ5Proto()
		properties = append(properties, property)
	}
	return &schema_j5pb.RootSchema{
		Type: &schema_j5pb.RootSchema_Object{
			Object: &schema_j5pb.Object{
				Description: s.description,
				Name:        s.name,
				Properties:  properties,
				Entity:      s.Entity,
			},
		},
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
				Description: s.description,
				Name:        s.name,
				Properties:  properties,
			},
		},
	}
}

func (s *OneofSchema) ToJ5ClientRoot() *schema_j5pb.RootSchema {
	return s.ToJ5Root()
}

type PropertySet []*ObjectProperty

type ObjectProperty struct {
	Parent RootSchema
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

func (prop *ObjectProperty) nestedClone(inParent []protoreflect.FieldNumber) *ObjectProperty {
	protoField := append(inParent, prop.ProtoField...)
	return &ObjectProperty{
		Parent:             prop.Parent,
		Schema:             prop.Schema,
		ProtoField:         protoField,
		JSONName:           prop.JSONName,
		Required:           prop.Required,
		ReadOnly:           prop.ReadOnly,
		WriteOnly:          prop.WriteOnly,
		ExplicitlyOptional: prop.ExplicitlyOptional,
		Description:        prop.Description,
	}
}
