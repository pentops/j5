package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/lib/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

/*** Interface ***/

type RangeMapCallback func(string, Field) error

type MapField interface {
	Field
	ItemSchema() j5schema.FieldSchema
	Range(RangeMapCallback) error
}

type MutableMapField interface {
	MapField
	NewElement(key string) (Field, error)
}

type LeafMapField interface {
	MapField
}

type MapOfContainerField interface {
	MutableMapField
	NewContainerElement(key string) (ContainerField, error)
}

/*** Implementation ***/

type baseMapField struct {
	fieldDefaults
	fieldContext
	value protoreflect.Map
	//fieldDescriptor protoreflect.FieldDescriptor
	schema *j5schema.MapField
}

func (mapField *baseMapField) IsSet() bool {
	return mapField.value.IsValid()
}

func (mapField *baseMapField) ItemSchema() j5schema.FieldSchema {
	return mapField.schema.Schema
}

func newMessageMapField(context fieldContext, schema *j5schema.MapField, value protoreflect.Map, factory messageFieldFactory) (MutableMapField, error) {

	base := baseMapField{
		fieldContext: context,
		value:        value,
		schema:       schema,
	}

	switch schema.Schema.(type) {
	case *j5schema.ObjectField:
		return &mapOfObjectField{
			MutableMapField: &mutableMapField{
				baseMapField: base,
				factory:      factory,
			},
		}, nil

	case *j5schema.OneofField:
		return &mapOfOneofField{
			MutableMapField: &mutableMapField{
				baseMapField: base,
				factory:      factory,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported schema type %T", schema.Schema)
	}
}

func newLeafMapField(context fieldContext, schema *j5schema.MapField, value protoreflect.Map, factory fieldFactory) (LeafMapField, error) {

	base := baseMapField{
		fieldContext: context,
		value:        value,
		schema:       schema,
	}

	switch st := schema.Schema.(type) {
	case *j5schema.ScalarSchema:
		return &mapOfScalarField{
			leafMapField: leafMapField{
				baseMapField: base,
				factory:      factory,
			},
			itemSchema: st,
		}, nil

	case *j5schema.EnumField:
		return &mapOfEnumField{
			leafMapField: leafMapField{
				baseMapField: base,
				factory:      factory,
			},
			itemSchema: st.Schema(),
		}, nil

	default:
		return nil, fmt.Errorf("unsupported schema type %T", schema.Schema)
	}

}

type mutableMapField struct {
	baseMapField
	factory messageFieldFactory
}

var _ MutableMapField = (*mutableMapField)(nil)

func (mapField *mutableMapField) Range(cb RangeMapCallback) error {
	if !mapField.value.IsValid() {
		return nil // empty map, probably invalid anyway, but has no keys.
	}

	var outerErr error

	mapField.value.Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
		keyStr := key.Value().String()
		field := mapField.wrapValue(keyStr, val)
		outerErr = cb(keyStr, field)
		return outerErr == nil
	})
	return outerErr
}

func (mapField *mutableMapField) wrapValue(key string, value protoreflect.Value) Field {
	context := &mapContext{
		name:   key,
		schema: mapField.schema,
	}
	return mapField.factory.buildField(context, value.Message())
}

func (mapField *mutableMapField) NewElement(key string) (Field, error) {
	mapKey := protoreflect.ValueOfString(key).MapKey()
	if mapField.value.Has(mapKey) {
		return nil, fmt.Errorf("key %q already exists in map", key)
	}
	itemVal := mapField.value.Mutable(mapKey)
	return mapField.wrapValue(key, itemVal), nil
}

type leafMapField struct {
	baseMapField
	factory fieldFactory
}

var _ LeafMapField = (*leafMapField)(nil)

func (mapField *leafMapField) Range(cb RangeMapCallback) error {
	if !mapField.value.IsValid() {
		return nil // empty map, probably invalid anyway, but has no keys.
	}

	var outerErr error

	mapField.value.Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
		keyStr := key.Value().String()
		field := mapField.wrapValue(keyStr)
		outerErr = cb(keyStr, field)
		return outerErr == nil
	})
	return outerErr
}

func (mapField *leafMapField) wrapValue(key string) Field {
	wrapped := &protoMapValue{
		mapVal: mapField.value,
		key:    protoreflect.ValueOfString(key).MapKey(),
	}
	context := &mapContext{
		name:   key,
		schema: mapField.schema,
	}
	return mapField.factory.buildField(context, wrapped)
}

func (mapField *leafMapField) setKey(key string, val protoreflect.Value) {
	keyVal := protoreflect.ValueOfString(key).MapKey()
	mapField.value.Set(keyVal, val)
}

type protoMapValue struct {
	mapVal protoreflect.Map
	key    protoreflect.MapKey
}

var _ protoContext = (*protoListValue)(nil)

func (pmv *protoMapValue) isSet() bool {
	_, ok := pmv.getValue()
	return ok
}

func (pmv *protoMapValue) setValue(val protoreflect.Value) error {
	if !val.IsValid() {
		pmv.mapVal.Clear(pmv.key)
		return nil
	}
	pmv.mapVal.Set(pmv.key, val)
	return nil
}

func (pmv *protoMapValue) getValue() (protoreflect.Value, bool) {
	itemVal := pmv.mapVal.Get(pmv.key)
	return itemVal, itemVal.IsValid()
}

func (pmv *protoMapValue) getMutableValue(createIfNotSet bool) (protoreflect.Value, error) {
	return pmv.mapVal.Get(pmv.key), nil
}

type mapContext struct {
	name   string
	schema *j5schema.MapField
}

var _ fieldContext = (*mapContext)(nil)

func (c *mapContext) NameInParent() string {
	return c.name
}

func (c *mapContext) IndexInParent() int {
	return -1
}

func (c *mapContext) FieldSchema() schema_j5pb.IsField_Type {
	return c.schema.Schema.ToJ5Field().Type
}

func (c *mapContext) TypeName() string {
	return c.schema.Schema.TypeName()
}
func (c *mapContext) PropertySchema() *schema_j5pb.ObjectProperty {
	return nil
}

func (c *mapContext) ProtoPath() []string {
	return []string{c.name}
}

func (c *mapContext) FullTypeName() string {
	return fmt.Sprintf("%s.{}%s", c.schema.FullName(), c.schema.Schema.TypeName())
}
