package j5reflect

import (
	"fmt"
	"sync"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
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

	value  protoreflect.List
	schema *j5schema.ArrayField
}

func (array *baseArrayField) IsSet() bool {
	return array.value.IsValid()
}

func (array *baseArrayField) ItemSchema() j5schema.FieldSchema {
	return array.schema.Schema
}

func (array *baseArrayField) Length() int {
	return array.value.Len()
}

func (array *baseArrayField) Truncate(newLen int) {
	array.value.Truncate(newLen)
}

func newMessageArrayField(context fieldContext, schema *j5schema.ArrayField, value protoreflect.List, factory messageFieldFactory) (ArrayField, error) {
	base := baseArrayField{
		fieldContext: context,
		schema:       schema,
		value:        value,
	}

	switch schema.Schema.(type) {
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

	default:
		return nil, fmt.Errorf("unsupported array item schema %T", schema.Schema)
	}
}

func newLeafArrayField(context fieldContext, schema *j5schema.ArrayField, value protoreflect.List, factory fieldFactory) (ArrayField, error) {
	if value == nil {
		panic("list value is nil for leaf")
	}

	base := baseArrayField{
		fieldContext: context,
		schema:       schema,
		value:        value,
	}

	switch st := schema.Schema.(type) {

	case *j5schema.ScalarSchema:
		return &arrayOfScalarField{
			leafArrayField: leafArrayField{
				baseArrayField: base,
				factory:        factory,
			},
			itemSchema: schema.Schema.(*j5schema.ScalarSchema),
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
		return nil, fmt.Errorf("unsupported array item schema %T", schema.Schema)
	}

}

type mutableArrayField struct {
	baseArrayField
	lock    sync.Mutex
	factory messageFieldFactory
}

var _ MutableArrayField = (*mutableArrayField)(nil)

func (array *mutableArrayField) NewElement() Field {
	array.lock.Lock()
	idx := array.value.Len()
	elem := array.value.AppendMutable().Message()
	array.lock.Unlock()
	return array.wrapValue(idx, elem)
}

func (array *mutableArrayField) RangeValues(cb RangeArrayCallback) error {
	if !array.value.IsValid() {
		return nil // TODO: return an error? Ranging a nil array means there's certainly nothing to range
	}

	for idx := range array.value.Len() {
		fieldVal := array.wrapValue(idx, array.value.Get(idx).Message())
		err := cb(idx, fieldVal)
		if err != nil {
			return err
		}
	}
	return nil
}

func (array *mutableArrayField) wrapValue(idx int, value protoreflect.Message) Field {
	schemaContext := &arrayContext{
		index:  idx,
		schema: array.schema,
	}

	field := array.factory.buildField(schemaContext, value)
	return field
}

type leafArrayField struct {
	baseArrayField
	lock    sync.Mutex
	factory fieldFactory
}

func (array *leafArrayField) RangeValues(cb RangeArrayCallback) error {
	if !array.value.IsValid() {
		return nil // TODO: return an error? Ranging a nil array means there's certainly nothing to range
	}

	for idx := range array.value.Len() {
		fieldVal := array.wrapValue(idx)
		err := cb(idx, fieldVal)
		if err != nil {
			return err
		}
	}
	return nil
}

func (array *leafArrayField) wrapValue(idx int) Field {
	protoItemContext := &protoListValue{
		list:  array.value,
		index: idx,

		//parentField: array.fieldDescriptor,
	}

	schemaContext := &arrayContext{
		index:  idx,
		schema: array.schema,
	}

	field := array.factory.buildField(schemaContext, protoItemContext)
	return field
}

func (array *leafArrayField) appendProtoValue(value protoreflect.Value) int {
	array.lock.Lock()
	idx := array.value.Len()
	array.value.Append(value)
	array.lock.Unlock()
	return idx
}

// protoListValue wraps a scalar/leaf type array, keeping pointer to the parent
// and the location within the parent where the object exists to make it
// semi-mutable.
type protoListValue struct {
	list protoreflect.List
	//parentField protoreflect.FieldDescriptor
	index int
}

var _ protoContext = (*protoListValue)(nil)

func (plv *protoListValue) isSet() bool {
	_, ok := plv.getValue()
	return ok
}

func (plv *protoListValue) setValue(val protoreflect.Value) error {
	if !val.IsValid() {
		return fmt.Errorf("cannot set a nil value to a list val")
	}
	plv.list.Set(plv.index, val)
	return nil
}

func (plv *protoListValue) getValue() (protoreflect.Value, bool) {
	itemVal := plv.list.Get(plv.index)
	return itemVal, itemVal.IsValid()
}

func (plv *protoListValue) getMutableValue(createIfNotSet bool) (protoreflect.Value, error) {
	return plv.list.Get(plv.index), nil
}

/*
func (plv *protoListValue) fieldDescriptor() protoreflect.FieldDescriptor {
	return plv.parentField
}*/

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

func (c *arrayContext) FieldSchema() schema_j5pb.IsField_Type {
	return c.schema.Schema.ToJ5Field().Type
}

func (c *arrayContext) TypeName() string {
	return c.schema.Schema.TypeName()
}

func (c *arrayContext) FullTypeName() string {
	return fmt.Sprintf("%s[%d] (%s)", c.schema.FullName(), c.index, c.schema.Schema.TypeName())
}

func (c *arrayContext) PropertySchema() *schema_j5pb.ObjectProperty {
	return nil
}

func (c *arrayContext) ProtoPath() []string {
	return []string{fmt.Sprintf("%d", c.index)}
}
