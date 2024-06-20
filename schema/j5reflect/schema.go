package j5reflect

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Schema struct {
	objectItem *ObjectSchema
	oneofItem  *OneofSchema
	arrayItem  *ArraySchema
	mapItem    *MapSchema
	enumItem   *EnumSchema
	scalarItem *ScalarSchema
	anyItem    *AnySchema
	refItem    *RefSchema
}

type schemaItem interface {
	ToJ5Proto() (*schema_j5pb.Schema, error)
}

func (s *Schema) ResolvedType() schemaItem {
	if s.refItem != nil {
		return s.refItem.To.ResolvedType()
	}
	return s.Type()
}

func (s *Schema) Type() schemaItem {

	if s.objectItem != nil {
		return s.objectItem
	} else if s.scalarItem != nil {
		return s.scalarItem
	} else if s.oneofItem != nil {
		return s.oneofItem
	} else if s.mapItem != nil {
		return s.mapItem
	} else if s.arrayItem != nil {
		return s.arrayItem
	} else if s.anyItem != nil {
		return s.anyItem
	} else if s.enumItem != nil {
		return s.enumItem
	} else if s.refItem != nil {
		return s.refItem
	}
	return nil

}

func (s *Schema) ToJ5Proto() (*schema_j5pb.Schema, error) {
	item := s.Type()
	if item == nil {
		return nil, fmt.Errorf("no schema type set")
	}
	return item.ToJ5Proto()
}

type RefSchema struct {
	Name protoreflect.FullName
	To   *Schema
}

func (s *RefSchema) ToJ5Proto() (*schema_j5pb.Schema, error) {
	return &schema_j5pb.Schema{
		Type: &schema_j5pb.Schema_Ref{
			Ref: string(s.Name),
		},
	}, nil
}

type ScalarSchema struct {
	// subset of the available schema types, everything excluding ref, oneof
	// wrapper, array, object, map
	proto *schema_j5pb.Schema

	Kind              protoreflect.Kind
	WellKnownTypeName protoreflect.FullName
}

func (s *ScalarSchema) ToJ5Proto() (*schema_j5pb.Schema, error) {
	return s.proto, nil
}

type AnySchema struct {
	Description string
}

func (s *AnySchema) ToJ5Proto() (*schema_j5pb.Schema, error) {
	return &schema_j5pb.Schema{
		Description: s.Description,
		Type: &schema_j5pb.Schema_Any{
			Any: &schema_j5pb.AnySchemaItem{},
		},
	}, nil
}

type EnumSchema struct {
	Description string
	NamePrefix  string
	Descriptor  protoreflect.EnumDescriptor
	Options     []*schema_j5pb.EnumItem_Value
}

func (s *EnumSchema) ToJ5Proto() (*schema_j5pb.Schema, error) {
	return &schema_j5pb.Schema{
		Description: s.Description,
		Type: &schema_j5pb.Schema_EnumItem{
			EnumItem: &schema_j5pb.EnumItem{
				Options: s.Options,
			},
		},
	}, nil
}

func walkName(src protoreflect.Descriptor) string {
	goTypeName := string(src.Name())
	parent := src.Parent()
	// if parent is a message
	msg, ok := parent.(protoreflect.MessageDescriptor)
	if !ok {
		return goTypeName
	}

	return fmt.Sprintf("%s_%s", walkName(msg), goTypeName)
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
	Description string
	Properties  PropertySet
	Rules       *schema_j5pb.ObjectRules

	ProtoMessage protoreflect.MessageDescriptor
}

func (s *ObjectSchema) GoTypeName() string {
	return walkName(s.ProtoMessage)
}

func (s *ObjectSchema) GrpcPackageName() string {
	return string(s.ProtoMessage.ParentFile().Package())
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
		Description: s.Description,
		Type: &schema_j5pb.Schema_ObjectItem{
			ObjectItem: &schema_j5pb.ObjectItem{
				Properties:       properties,
				Rules:            s.Rules,
				ProtoMessageName: string(s.ProtoMessage.Name()),
				ProtoFullName:    string(s.ProtoMessage.FullName()),
			},
		},
	}, nil
}

type OneofSchema struct {
	Description string
	Properties  PropertySet
	Rules       *schema_j5pb.OneofRules

	// optional, only for proto types, otherwise is a oneof in the outer message
	ProtoMessage protoreflect.MessageDescriptor

	OneofDescriptor protoreflect.OneofDescriptor
}

func (s *OneofSchema) GoTypeName() string {
	if s.ProtoMessage != nil {
		return walkName(s.ProtoMessage)
	}
	if s.OneofDescriptor != nil {
		return walkName(s.OneofDescriptor)
	}
	panic("invalid oneof, no message or descriptor set")
}

func (s *OneofSchema) GrpcPackageName() string {
	if s.ProtoMessage != nil {
		return string(s.ProtoMessage.ParentFile().Package())
	}
	if s.OneofDescriptor != nil {
		return string(s.OneofDescriptor.ParentFile().Package())
	}
	panic("invalid oneof, no message or descriptor set")
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
	wrapperItem := &schema_j5pb.OneofWrapperItem{
		Properties: properties,
		Rules:      s.Rules,
	}

	if s.ProtoMessage != nil {
		wrapperItem.ProtoMessageName = string(s.ProtoMessage.Name())
		wrapperItem.ProtoFullName = string(s.ProtoMessage.FullName())
	}
	return &schema_j5pb.Schema{
		Description: s.Description,
		Type: &schema_j5pb.Schema_OneofWrapper{
			OneofWrapper: wrapperItem,
		},
	}, nil
}

type ObjectProperty struct {
	Schema *Schema

	ProtoField []protoreflect.FieldDescriptor

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
		fieldPath[idx] = int32(field.Number())
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

func (prop *ObjectProperty) GoFieldName() string {
	return strcase.ToCamel(prop.JSONName)
}

func (prop *ObjectProperty) ProtoName() (string, error) {
	if len(prop.ProtoField) != 1 {
		return "", fmt.Errorf("invalid property for proto name")
	}
	return string(prop.ProtoField[0].Name()), nil
}

type MapSchema struct {
	Description string
	Schema      *Schema
	Rules       *schema_j5pb.MapRules
}

func (s *MapSchema) ToJ5Proto() (*schema_j5pb.Schema, error) {
	item, err := s.Schema.ToJ5Proto()
	if err != nil {
		return nil, fmt.Errorf("map item: %w", err)
	}

	return &schema_j5pb.Schema{
		Description: s.Description,
		Type: &schema_j5pb.Schema_MapItem{
			MapItem: &schema_j5pb.MapItem{
				ItemSchema: item,
				Rules:      s.Rules,
			},
		},
	}, nil
}

type ArraySchema struct {
	Description string
	Schema      *Schema
	Rules       *schema_j5pb.ArrayRules
}

func (s *ArraySchema) ToJ5Proto() (*schema_j5pb.Schema, error) {

	item, err := s.Schema.ToJ5Proto()
	if err != nil {
		return nil, fmt.Errorf("array item: %w", err)
	}

	return &schema_j5pb.Schema{
		Description: s.Description,
		Type: &schema_j5pb.Schema_ArrayItem{
			ArrayItem: &schema_j5pb.ArrayItem{
				Items: item,
				Rules: s.Rules,
			},
		},
	}, nil
}
