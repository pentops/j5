package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/lib/j5reflect/protoval"
	"github.com/pentops/j5/lib/j5schema"
)

/*** Interface ***/

type RangeMapCallback func(string, Field) error

type MapField interface {
	Field
	ItemSchema() j5schema.FieldSchema
	Range(RangeMapCallback) error
	NewElement(key string) (Field, error)
	GetOrCreateElement(key string) (Field, error)
	GetElement(key string) (Field, bool, error)
}

type MutableMapField interface {
	MapField
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
	value protoval.MapValue
	//fieldDescriptor protoreflect.FieldDescriptor
	schema *j5schema.MapField

	factory fieldFactory
}

func (mapField *baseMapField) IsSet() bool {
	return mapField.value.IsSet()
}

func (mapField *baseMapField) ItemSchema() j5schema.FieldSchema {
	return mapField.schema.ItemSchema
}

func (mapField *baseMapField) SetDefaultValue() error {
	return nil

}
func (mapField *baseMapField) GetElement(key string) (Field, bool, error) {
	value := mapField.value.ValueAt(key)
	return mapField.wrapValue(key, value), value.IsSet(), nil
}

func (mapField *baseMapField) GetOrCreateElement(key string) (Field, error) {
	value := mapField.value.ValueAt(key)
	if !value.IsSet() {
		value.Create()
	}
	value.Create()
	return mapField.wrapValue(key, value), nil
}

func (mapField *baseMapField) Range(cb RangeMapCallback) error {
	if !mapField.value.IsSet() {
		return nil // empty map, probably invalid anyway, but has no keys.
	}

	var outerErr error

	mapField.value.RangeValues(func(key string, val protoval.Value) bool {
		field := mapField.wrapValue(key, val)
		outerErr = cb(key, field)
		return outerErr == nil
	})
	return outerErr
}

func (mapField *baseMapField) wrapValue(key string, value protoval.Value) Field {
	context := &mapContext{
		name:   key,
		schema: mapField.schema,
	}
	return mapField.factory.buildField(context, value)
}

func (mapField *baseMapField) NewElement(key string) (Field, error) {
	value := mapField.value.ValueAt(key)
	if value.IsSet() {
		return nil, fmt.Errorf("key %q already exists in map", key)
	}
	value.Create()
	return mapField.wrapValue(key, value), nil
}

func newMessageMapField(context fieldContext, schema *j5schema.MapField, value protoval.MapValue, factory fieldFactory) (MutableMapField, error) {

	base := baseMapField{
		fieldContext: context,
		value:        value,
		schema:       schema,
		factory:      factory,
	}

	switch schema.ItemSchema.(type) {
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
	default:
		return nil, fmt.Errorf("unsupported schema type %T", schema.ItemSchema)
	}
}

func newLeafMapField(context fieldContext, schema *j5schema.MapField, value protoval.MapValue, factory fieldFactory) (LeafMapField, error) {

	base := baseMapField{
		fieldContext: context,
		value:        value,
		schema:       schema,
		factory:      factory,
	}

	switch st := schema.ItemSchema.(type) {
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
				factory:      factory,
			},
			itemSchema: st.Schema(),
		}, nil

	default:
		return nil, fmt.Errorf("unsupported schema type %T", schema.ItemSchema)
	}

}

type mutableMapField struct {
	baseMapField
}

var _ MutableMapField = (*mutableMapField)(nil)

func (mapField *mutableMapField) AsMap() (MapField, bool) {
	return mapField, true
}

type leafMapField struct {
	baseMapField
	factory fieldFactory
}

var _ LeafMapField = (*leafMapField)(nil)

func (mapField *leafMapField) AsMap() (MapField, bool) {
	return mapField, true
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

func (c *mapContext) FieldSchema() j5schema.FieldSchema {
	return c.schema.ItemSchema
}

func (c *mapContext) TypeName() string {
	return c.schema.ItemSchema.TypeName()
}
func (c *mapContext) PropertySchema() *j5schema.ObjectProperty {
	return nil
}

func (c *mapContext) ProtoPath() []string {
	return []string{c.name}
}

func (c *mapContext) FullTypeName() string {
	return fmt.Sprintf("%s.{}%s", c.schema.FullName(), c.schema.ItemSchema.TypeName())
}
