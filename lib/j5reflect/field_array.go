package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/internal/j5schema"
)

type baseArrayField struct {
	fieldDefaults
	fieldInParent *realProtoMessageField
	schema        *j5schema.ArrayField
	factory       fieldFactory
}

func (field *baseArrayField) IsSet() bool {
	return field.fieldInParent.isSet()
}

func (field *baseArrayField) Type() FieldType {
	return FieldTypeArray
}

func (field *baseArrayField) ItemSchema() j5schema.FieldSchema {
	return field.schema.Schema
}

func (field *baseArrayField) SetDefault() error {
	field.fieldInParent.getOrCreateMutable().List()
	return nil
}

func (field *baseArrayField) Range(cb func(Field) error) error {
	if !field.fieldInParent.isSet() {
		return nil
	}
	list := field.fieldInParent.getValue().List()

	for i := 0; i < list.Len(); i++ {
		val := list.Get(i)
		wrapped := &protoListItem{
			protoValueWrapper: protoValueWrapper{
				value:   val,
				prField: field.fieldInParent.fieldInParent,
			},
			prList: list,
			idx:    i,
		}
		context := &arrayContext{
			index:  i,
			schema: field.schema,
		}
		property := field.factory.buildField(context, wrapped)

		err := cb(property)
		if err != nil {
			return err
		}
	}
	return nil
}

func newArrayField(context fieldContext, schema *j5schema.ArrayField, value *realProtoMessageField) (ArrayField, error) {
	if !value.fieldInParent.IsList() {
		return nil, fmt.Errorf("ArrayField is not a list")
	}

	factory, err := newFieldFactory(schema.Schema, value.fieldInParent)
	if err != nil {
		return nil, err
	}

	base := baseArrayField{
		fieldDefaults: fieldDefaults{
			fieldType: FieldTypeArray,
			context:   context,
		},
		fieldInParent: value,
		schema:        schema,
		factory:       factory,
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
}

var _ MutableArrayField = (*mutableArrayField)(nil)

func (field *mutableArrayField) NewElement() Field {
	list := field.fieldInParent.getOrCreateMutable().List()
	idx := list.Len()
	elem := list.AppendMutable()
	element := &protoListItem{
		protoValueWrapper: protoValueWrapper{
			prField: field.fieldInParent.fieldInParent,
			value:   elem,
		},
		prList: list,
		idx:    idx,
	}
	context := &arrayContext{
		index:  idx,
		schema: field.schema,
	}
	property := field.factory.buildField(context, element)
	return property
}

type leafArrayField struct {
	baseArrayField
}

func (field *leafArrayField) AppendGoValue(value interface{}) error {
	list := field.fieldInParent.getOrCreateMutable().List()
	reflectValue, err := scalarReflectFromGo(field.schema.Schema.ToJ5Field(), value)
	if err != nil {
		return err
	}
	list.Append(reflectValue)
	return nil
}

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