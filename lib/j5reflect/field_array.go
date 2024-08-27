package j5reflect

import (
	"fmt"
	"sync"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/internal/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

/*** Interface ***/

type RangeArrayCallback func(int, Field) error

type ArrayField interface {
	Field
	RangeValues(RangeArrayCallback) error
}

type MutableArrayField interface {
	ArrayField
	NewElement() Field
}

/*** Implementation ***/

type baseArrayField struct {
	fieldDefaults
	value protoreflect.List
	//fieldDescriptor protoreflect.FieldDescriptor
	schema  *j5schema.ArrayField
	factory fieldFactory
}

func (array *baseArrayField) Type() FieldType {
	return FieldTypeArray
}

func (array *baseArrayField) IsSet() bool {
	return array.value.IsValid()
}

func (array *baseArrayField) ItemSchema() j5schema.FieldSchema {
	return array.schema.Schema
}

func (array *baseArrayField) RangeValues(cb RangeArrayCallback) error {
	if !array.value.IsValid() {
		return nil // TODO: return an error? Ranging a nil array means there's certainly nothing to range
	}

	for idx := 0; idx < array.value.Len(); idx++ {
		fieldVal := array.wrapValue(idx, array.value.Get(idx))
		err := cb(idx, fieldVal)
		if err != nil {
			return err
		}
	}
	return nil
}

func (array *baseArrayField) wrapValue(idx int, value protoreflect.Value) Field {
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

func newArrayField(context fieldContext, schema *j5schema.ArrayField, value protoreflect.List, factory fieldFactory) (ArrayField, error) {

	base := baseArrayField{
		fieldDefaults: fieldDefaults{
			fieldType: FieldTypeArray,
			context:   context,
		},
		schema:  schema,
		value:   value,
		factory: factory,
		//fieldDescriptor: value.fieldInParent,
	}

	switch st := schema.Schema.(type) {
	case *j5schema.ObjectField:
		return &arrayOfObjectField{
			mutableArrayField: mutableArrayField{
				baseArrayField: base,
			},
		}, nil

	case *j5schema.OneofField:
		return &arrayOfOneofField{
			mutableArrayField: mutableArrayField{
				baseArrayField: base,
			},
		}, nil

	case *j5schema.ScalarSchema:
		return &arrayOfScalarField{
			leafArrayField: leafArrayField{
				baseArrayField: base,
			},
			itemSchema: schema.Schema.(*j5schema.ScalarSchema),
		}, nil

	case *j5schema.EnumField:
		return &arrayOfEnumField{
			leafArrayField: leafArrayField{
				baseArrayField: base,
			},
			itemSchema: st.Schema(),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported array item schema %T", schema.Schema)
	}

}

type mutableArrayField struct {
	baseArrayField
	value protoreflect.List
	lock  sync.Mutex
}

var _ MutableArrayField = (*mutableArrayField)(nil)

func (array *mutableArrayField) NewElement() Field {
	array.lock.Lock()
	idx := array.value.Len()
	elem := array.value.AppendMutable()
	array.lock.Unlock()
	return array.wrapValue(idx, elem)
}

type leafArrayField struct {
	baseArrayField
	lock sync.Mutex
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

func (c *arrayContext) nameInParent() string {
	return fmt.Sprintf("%d", c.index)
}

func (c *arrayContext) indexInParent() int {
	return c.index
}

func (c *arrayContext) fieldSchema() schema_j5pb.IsField_Type {
	return c.schema.Schema.ToJ5Field().Type
}

func (c *arrayContext) typeName() string {
	return c.schema.Schema.TypeName()
}

func (c *arrayContext) fullTypeName() string {
	return fmt.Sprintf("%s.[]%s", c.schema.FullName(), c.schema.Schema.TypeName())
}

func (c *arrayContext) propertySchema() *schema_j5pb.ObjectProperty {
	return nil
}

func (c *arrayContext) protoPath() []string {
	return []string{fmt.Sprintf("%d", c.index)}
}
