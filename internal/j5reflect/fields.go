package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/internal/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type FieldType string

const (
	FieldTypeUnknown = FieldType("?")
	FieldTypeObject  = FieldType("object")
	FieldTypeOneof   = FieldType("oneof")
	FieldTypeEnum    = FieldType("enum")
	FieldTypeArray   = FieldType("array")
	FieldTypeMap     = FieldType("map")
	FieldTypeScalar  = FieldType("scalar")
)

type fieldFactory interface {
	buildField(value protoValueContext) Field
}

type objectFieldFactory struct {
	schema *j5schema.ObjectField
}

func (f *objectFieldFactory) buildField(value protoValueContext) Field {
	return newObjectField(f.schema, value)
}

type oneofFieldFactory struct {
	schema *j5schema.OneofField
}

func (f *oneofFieldFactory) buildField(value protoValueContext) Field {
	return newOneofField(f.schema, value)
}

type enumFieldFactory struct {
	schema *j5schema.EnumField
}

func (f *enumFieldFactory) buildField(value protoValueContext) Field {
	return newEnumField(f.schema, value)
}

type scalarFieldFactory struct {
	schema *j5schema.ScalarSchema
}

func (f *scalarFieldFactory) buildField(value protoValueContext) Field {
	return newScalarField(f.schema, value)
}

func newFieldFactory(schema j5schema.FieldSchema, field protoreflect.FieldDescriptor) (fieldFactory, error) {
	switch st := schema.(type) {
	case *j5schema.ObjectField:
		if field.Kind() != protoreflect.MessageKind {
			return nil, fmt.Errorf("ObjectField is kind %s", field.Kind())
		}
		return &objectFieldFactory{schema: st}, nil

	case *j5schema.OneofField:
		if field.Kind() != protoreflect.MessageKind {
			return nil, fmt.Errorf("OneofField is kind %s", field.Kind())
		}
		return &oneofFieldFactory{schema: st}, nil

	case *j5schema.EnumField:
		if field.Kind() != protoreflect.EnumKind {
			return nil, fmt.Errorf("EnumField is kind %s", field.Kind())
		}
		return &enumFieldFactory{schema: st}, nil

	case *j5schema.ScalarSchema:
		if st.WellKnownTypeName != "" {
			if field.Kind() != protoreflect.MessageKind {
				return nil, fmt.Errorf("ScalarField is kind %s, want message for %T", field.Kind(), st.Proto.Type)
			}
			if string(field.Message().FullName()) != string(st.WellKnownTypeName) {
				return nil, fmt.Errorf("ScalarField message is %s, want %s for %T", field.Message().FullName(), st.WellKnownTypeName, st.Proto.Type)
			}
		} else if field.Kind() != st.Kind {
			return nil, fmt.Errorf("ScalarField value is kind %s, want %s for %T", field.Kind(), st.Kind, st.Proto.Type)
		}
		return &scalarFieldFactory{schema: st}, nil

	default:
		return nil, fmt.Errorf("unsupported schema type %T", schema)
	}
}

type objectField struct {
	value   protoValueContext
	schema  *j5schema.ObjectField
	_object *ObjectImpl
}

var _ ObjectField = (*objectField)(nil)

func newObjectField(schema *j5schema.ObjectField, value protoValueContext) *objectField {
	of := &objectField{
		value:  value,
		schema: schema,
	}
	return of
}

func (obj *objectField) asProperty(base fieldBase) Property {
	return &objectProperty{
		field:     obj,
		fieldBase: base,
	}
}

func (obj *objectField) Schema() j5schema.FieldSchema {
	return obj.schema
}

func (obj *objectField) Type() FieldType {
	return FieldTypeObject
}

func (obj *objectField) IsSet() bool {
	return obj.value.isSet()
}

func (obj *objectField) SetDefault() error {
	_ = obj.value.getOrCreateMutable()
	return nil
}

func (obj *objectField) Object() (Object, error) {
	if obj._object == nil {
		msgChild, err := obj.value.getOrCreateChildMessage()
		if err != nil {
			return nil, err
		}
		built, err := newObject(obj.schema.Schema(), msgChild)
		if err != nil {
			return nil, err
		}
		obj._object = built
	}
	return obj._object, nil
}

type oneofField struct {
	value  protoValueContext
	schema *j5schema.OneofField
	_oneof *OneofImpl
}

var _ OneofField = (*oneofField)(nil)

func newOneofField(schema *j5schema.OneofField, value protoValueContext) *oneofField {
	return &oneofField{
		value:  value,
		schema: schema,
	}
}

func (field *oneofField) asProperty(base fieldBase) Property {
	return &oneofProperty{
		field:     field,
		fieldBase: base,
	}
}

func (field *oneofField) IsSet() bool {
	return field.value.isSet()
}

func (field *oneofField) SetDefault() error {
	return fmt.Errorf("cannot set default on oneof fields")
}

func (field *oneofField) Type() FieldType {
	return FieldTypeOneof
}

func (field *oneofField) Oneof() (*OneofImpl, error) {
	if field._oneof == nil {
		msgChild, err := field.value.getOrCreateChildMessage()
		if err != nil {
			return nil, err
		}

		obj, err := newOneof(field.schema.Schema(), msgChild)
		if err != nil {
			return nil, err
		}
		field._oneof = obj
	}

	return field._oneof, nil
}

type enumField struct {
	value  protoValueContext
	schema *j5schema.EnumField
}

var _ EnumField = (*enumField)(nil)

func newEnumField(schema *j5schema.EnumField, value protoValueContext) *enumField {
	return &enumField{
		value:  value,
		schema: schema,
	}
}

func (ef *enumField) asProperty(base fieldBase) Property {
	return &enumProperty{
		field:     ef,
		fieldBase: base,
	}
}

func (ef *enumField) IsSet() bool {
	return ef.value.isSet()
}

func (ef *enumField) SetDefault() error {
	ef.value.getOrCreateMutable()
	return nil
}

func (ef *enumField) Type() FieldType {
	return FieldTypeEnum
}

func (ef *enumField) GetValue() (*j5schema.EnumOption, error) {
	value := int32(ef.value.getValue().Enum())
	for _, val := range ef.schema.Schema().Options {
		if val.Number == value {
			return val, nil
		}
	}
	return nil, fmt.Errorf("enum value %d not found", value)
}

func (ef *enumField) SetFromString(val string) error {
	option := ef.schema.Schema().OptionByName(val)
	if option != nil {
		ef.value.setValue(protoreflect.ValueOfEnum(protoreflect.EnumNumber(option.Number)))
		return nil
	}
	return fmt.Errorf("enum value %s not found", val)
}

/*
	type ArrayItem interface {
		SetGoValue(interface{}) error
	}

	type arrayItem struct {
		field  *arrayField
		schema j5schema.FieldSchema
		idx    int
	}
*/

type scalarField struct {
	field  protoValueContext
	schema *j5schema.ScalarSchema
}

func newScalarField(schema *j5schema.ScalarSchema, value protoValueContext) *scalarField {

	return &scalarField{
		field:  value,
		schema: schema,
	}
}

func (sf *scalarField) asProperty(base fieldBase) Property {
	return &scalarProperty{
		field:     sf,
		fieldBase: base,
	}
}

func (sf *scalarField) IsSet() bool {
	return sf.field.isSet()
}

func (sf *scalarField) SetDefault() error {
	sf.field.getOrCreateMutable()
	return nil
}

func (sf *scalarField) Type() FieldType {
	return FieldTypeScalar
}

func (sf *scalarField) Schema() *j5schema.ScalarSchema {
	return sf.schema
}

func (sf *scalarField) SetASTValue(value ASTValue) error {
	reflectValue, err := scalarReflectFromAST(sf.schema.Proto, value)
	if err != nil {
		return err
	}

	sf.field.setValue(reflectValue)
	return nil
}

func (sf *scalarField) SetGoValue(value interface{}) error {
	reflectValue, err := scalarReflectFromGo(sf.schema.Proto, value)
	if err != nil {
		return err
	}

	sf.field.setValue(reflectValue)
	return nil
}

func (sf *scalarField) ToGoValue() (interface{}, error) {
	return scalarGoFromReflect(sf.schema.Proto, sf.field.getValue())
}
