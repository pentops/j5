package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/lib/j5reflect/protoval"
	"github.com/pentops/j5/lib/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

/*** Interface ***/

type RangeArrayCallback func(int, Field) error

type ArrayField interface {
	Field
	ItemSchema() j5schema.FieldSchema
	RangeValues(RangeArrayCallback) error
	Length() int
	Truncate(int)
}

type MutableArrayField interface {
	ArrayField
	NewElement() Field
}

type ArrayOfContainerField interface {
	MutableArrayField
	NewContainerElement() (ContainerField, int)
	RangeContainers(func(int, ContainerField) error) error
}

/*** Implementation ***/

type baseArrayField struct {
	fieldDefaults
	fieldContext

	value  protoval.ListValue
	schema *j5schema.ArrayField
}

func (array *baseArrayField) IsSet() bool {
	return array.value.IsSet()
}

func (array *baseArrayField) ItemSchema() j5schema.FieldSchema {
	return array.schema.ItemSchema
}

func (array *baseArrayField) Length() int {
	return array.value.Len()
}

func (array *baseArrayField) Truncate(newLen int) {
	array.value.Truncate(newLen)
}

func (array *baseArrayField) SetDefaultValue() error {
	return nil
}

func newMessageArrayField(context fieldContext, schema *j5schema.ArrayField, value protoval.ListValue, factory fieldFactory) (ArrayField, error) {
	base := baseArrayField{
		fieldContext: context,
		schema:       schema,
		value:        value,
	}

	switch schema.ItemSchema.(type) {
	case *j5schema.ObjectField:
		return &arrayOfObjectField{
			mutableArrayField: mutableArrayField{
				baseArrayField: base,
				factory:        factory,
			},
		}, nil

	case *j5schema.OneofField:
		return &arrayOfOneofField{
			mutableArrayField: mutableArrayField{
				baseArrayField: base,
				factory:        factory,
			},
		}, nil

	case *j5schema.AnyField:
		return &arrayOfAnyField{
			mutableArrayField: mutableArrayField{
				baseArrayField: base,
				factory:        factory,
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported (message) array item schema %T", schema.ItemSchema)
	}
}

func newLeafArrayField(context fieldContext, schema *j5schema.ArrayField, value protoval.ListValue, factory fieldFactory) (ArrayField, error) {
	if value == nil {
		panic("list value is nil for leaf")
	}

	base := baseArrayField{
		fieldContext: context,
		schema:       schema,
		value:        value,
	}

	switch st := schema.ItemSchema.(type) {

	case *j5schema.ScalarSchema:
		return &arrayOfScalarField{
			leafArrayField: leafArrayField{
				baseArrayField: base,
				factory:        factory,
			},
			itemSchema: schema.ItemSchema.(*j5schema.ScalarSchema),
		}, nil

	case *j5schema.EnumField:
		return &arrayOfEnumField{
			leafArrayField: leafArrayField{
				baseArrayField: base,
				factory:        factory,
			},
			itemSchema: st.Schema(),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported (leaf) array item schema %T", schema.ItemSchema)
	}

}

type mutableArrayField struct {
	baseArrayField
	factory fieldFactory
}

var _ MutableArrayField = (*mutableArrayField)(nil)

func (array *mutableArrayField) NewElement() Field {
	elem, idx := array.value.AppendMessage()
	return array.wrapValue(idx, elem)
}

func (array *mutableArrayField) RangeValues(cb RangeArrayCallback) error {
	if !array.value.IsSet() {
		return nil // TODO: return an error? Ranging a nil array means there's certainly nothing to range
	}

	for idx := range array.value.Len() {
		val, hasVal := array.value.ValueAt(idx)
		if !hasVal {
			return fmt.Errorf("array value at index %d is not set", idx)
		}

		fieldVal := array.wrapValue(idx, val)
		err := cb(idx, fieldVal)
		if err != nil {
			return err
		}
	}
	return nil
}

func (array *mutableArrayField) wrapValue(idx int, value protoval.Value) Field {
	schemaContext := &arrayContext{
		index:  idx,
		schema: array.schema,
	}

	field := array.factory.buildField(schemaContext, value)
	return field
}

type leafArrayField struct {
	baseArrayField
	factory fieldFactory
}

func (array *leafArrayField) RangeValues(cb RangeArrayCallback) error {
	if !array.value.IsSet() {
		return nil // TODO: return an error? Ranging a nil array means there's certainly nothing to range
	}

	for idx := range array.value.Len() {
		itemValue, ok := array.value.ValueAt(idx)
		if !ok {
			return nil // This should not happen, but if it does, we return nil
		}
		fieldVal := array.wrapValue(idx, itemValue)
		err := cb(idx, fieldVal)
		if err != nil {
			return err
		}
	}
	return nil
}

func (array *leafArrayField) wrapValue(idx int, value protoval.Value) Field {
	schemaContext := &arrayContext{
		index:  idx,
		schema: array.schema,
	}

	field := array.factory.buildField(schemaContext, value)
	return field
}

func (array *leafArrayField) appendProtoValue(value protoreflect.Value) int {
	_, idx := array.value.AppendValue(value)
	return idx
}

type arrayContext struct {
	index  int
	schema *j5schema.ArrayField
}

var _ fieldContext = (*arrayContext)(nil)

func (c *arrayContext) NameInParent() string {
	return fmt.Sprintf("%d", c.index)
}

func (c *arrayContext) IndexInParent() int {
	return c.index
}

func (c *arrayContext) FieldSchema() j5schema.FieldSchema {
	return c.schema.ItemSchema
}

func (c *arrayContext) TypeName() string {
	return c.schema.ItemSchema.TypeName()
}

func (c *arrayContext) FullTypeName() string {
	return fmt.Sprintf("%s[%d] (%s)", c.schema.FullName(), c.index, c.schema.ItemSchema.TypeName())
}

func (c *arrayContext) PropertySchema() *j5schema.ObjectProperty {
	return nil
}

func (c *arrayContext) ProtoPath() []string {
	return []string{fmt.Sprintf("%d", c.index)}
}
