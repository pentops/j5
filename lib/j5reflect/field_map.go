package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/internal/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type mutableMapField struct {
	baseMapField
}

var _ MutableMapField = (*mutableMapField)(nil)

func (field *mutableMapField) NewValue(key string) Field {
	mapVal := field.fieldInParent.getOrCreateMutable().Map()
	keyVal := protoreflect.ValueOfString(key).MapKey()
	val := mapVal.Get(keyVal)
	wrapped := &protoMapItem{
		protoValueWrapper: protoValueWrapper{
			value:   val,
			prField: field.fieldInParent.fieldInParent.MapValue(),
		},
		key:   keyVal,
		prMap: mapVal,
	}
	context := &mapContext{
		name:   key,
		schema: field.schema,
	}
	property := field.factory.buildField(context, wrapped)
	return property
}

type mapOfObjectField struct {
	MutableMapField
}

func (field *mapOfObjectField) NewObjectValue(key string) (Object, error) {
	of := field.NewValue(key).(ObjectField)
	return of.Object()
}

type MapOfOneofField interface {
	NewOneofValue(key string) (*OneofImpl, error)
}

type mapOfOneofField struct {
	MutableMapField
}

func (field *mapOfOneofField) NewOneofValue(key string) (Oneof, error) {
	of := field.NewValue(key).(OneofField)
	return of.Oneof()
}

type baseMapField struct {
	fieldDefaults
	fieldInParent *realProtoMessageField
	schema        *j5schema.MapField
	factory       fieldFactory
}

func (field *baseMapField) Type() FieldType {
	return FieldTypeMap
}

func (field *baseMapField) ItemSchema() j5schema.FieldSchema {
	return field.schema.Schema
}

func (field *baseMapField) IsSet() bool {
	return field.fieldInParent.isSet()
}

func (field *baseMapField) SetDefault() error {
	field.fieldInParent.getOrCreateMutable().Map()
	return nil
}

func (field *baseMapField) Range(cb func(string, Field) error) error {
	if !field.fieldInParent.isSet() {
		return nil
	}
	mapVal := field.fieldInParent.getValue().Map()
	var outerErr error

	fieldDef := field.fieldInParent.fieldInParent.MapValue()

	mapVal.Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
		keyStr := key.Value().String()
		wrapped := &protoMapItem{
			protoValueWrapper: protoValueWrapper{
				value:   val,
				prField: fieldDef,
			},
			key:   key,
			prMap: mapVal,
		}
		context := &mapContext{
			name:   keyStr,
			schema: field.schema,
		}
		itemField := field.factory.buildField(context, wrapped)
		outerErr = cb(keyStr, itemField)
		return outerErr == nil
	})
	return outerErr
}

func newMapField(context fieldContext, schema *j5schema.MapField, value *realProtoMessageField) (MapField, error) {
	if !value.fieldInParent.IsMap() {
		return nil, fmt.Errorf("MapField is not a map")
	}

	factory, err := newFieldFactory(schema.Schema, value.fieldInParent.MapValue())
	if err != nil {
		return nil, fmt.Errorf("factory for map value: %w", err)
	}

	base := baseMapField{
		fieldDefaults: fieldDefaults{
			fieldType: FieldTypeMap,
			context:   context,
		},
		fieldInParent: value,
		schema:        schema,
		factory:       factory,
	}

	switch st := schema.Schema.(type) {
	case *j5schema.ObjectField:
		return &mapOfObjectField{
			MutableMapField: &mutableMapField{
				baseMapField: base,
			},
		}, nil

	case *j5schema.OneofField:
		return &mapOfOneofField{
			MutableMapField: &mutableMapField{
				baseMapField: base,
			},
		}, nil

	case *j5schema.ScalarSchema:
		return &mapOfScalarField{
			leafMapField: leafMapField{
				baseMapField: base,
			},
			itemSchema: st,
		}, nil

	case *j5schema.EnumField:
		return &mapOfEnumField{
			leafMapField: leafMapField{
				baseMapField: base,
			},
			itemSchema: st.Schema(),
		}, nil
	}

	if schema.Schema.Mutable() {
		return &mutableMapField{
			baseMapField: base,
		}, nil
	} else {
		return &leafMapField{
			baseMapField: base,
		}, nil
	}
}

type leafMapField struct {
	baseMapField
}

var _ MapField = (*leafMapField)(nil)

func (field *leafMapField) setKey(key protoreflect.MapKey, val protoreflect.Value) {
	mapVal := field.fieldInParent.getOrCreateMutable().Map()
	mapVal.Set(key, val)
}

type mapOfScalarField struct {
	leafMapField
	itemSchema *j5schema.ScalarSchema
}

func (field *mapOfScalarField) SetGoScalar(key string, value interface{}) error {
	reflVal, err := scalarReflectFromGo(field.itemSchema.Proto, value)
	if err != nil {
		return fmt.Errorf("converting value to proto: %w", err)
	}

	field.setKey(protoreflect.ValueOfString(key).MapKey(), reflVal)
	return nil
}

type mapOfEnumField struct {
	leafMapField
	itemSchema *j5schema.EnumSchema
}

func (field *mapOfEnumField) SetEnum(key string, value string) error {
	option := field.itemSchema.OptionByName(value)
	if option == nil {
		return fmt.Errorf("enum value %s not found", value)
	}

	field.setKey(protoreflect.ValueOfString(key).MapKey(), protoreflect.ValueOfEnum(protoreflect.EnumNumber(option.Number())))
	return nil
}

type mapContext struct {
	name   string
	schema *j5schema.MapField
}

var _ fieldContext = (*mapContext)(nil)

func (c *mapContext) nameInParent() string {
	return c.name
}

func (c *mapContext) indexInParent() int {
	return -1
}

func (c *mapContext) fieldSchema() schema_j5pb.IsField_Type {
	return c.schema.Schema.ToJ5Field().Type
}
func (c *mapContext) typeName() string {
	return c.schema.Schema.TypeName()
}
func (c *mapContext) propertySchema() *schema_j5pb.ObjectProperty {
	return nil
}

func (c *mapContext) protoPath() []string {
	return []string{c.name}
}

func (c *mapContext) fullTypeName() string {
	return fmt.Sprintf("%s.{}%s", c.schema.FullName(), c.schema.Schema.TypeName())
}