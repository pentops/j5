package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/internal/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type arrayOfObjectField struct {
	MutableArrayField
}

var _ ArrayOfObjectField = (*arrayOfObjectField)(nil)

func (field *arrayOfObjectField) NewObjectElement() (Object, error) {
	of := field.NewElement().(ObjectField)
	return of.Object()
}

type arrayOfOneofField struct {
	MutableArrayField
}

func (field *arrayOfOneofField) NewOneofElement() (Oneof, error) {
	of := field.NewElement().(OneofField)
	return of.Oneof()
}

var _ ArrayOfOneofField = (*arrayOfOneofField)(nil)

type baseArrayField struct {
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
		property := field.factory.buildField(wrapped)

		err := cb(property)
		if err != nil {
			return err
		}
	}
	return nil
}

func newArrayField(schema *j5schema.ArrayField, value *realProtoMessageField) (ArrayField, error) {
	if !value.fieldInParent.IsList() {
		return nil, fmt.Errorf("ArrayField is not a list")
	}

	factory, err := newFieldFactory(schema.Schema, value.fieldInParent)
	if err != nil {
		return nil, err
	}

	base := baseArrayField{
		fieldInParent: value,
		schema:        schema,
		factory:       factory,
	}

	switch st := schema.Schema.(type) {
	case *j5schema.ObjectField:
		return &arrayOfObjectField{
			MutableArrayField: &mutableArrayField{
				baseArrayField: base,
			},
		}, nil

	case *j5schema.OneofField:
		return &arrayOfOneofField{
			MutableArrayField: &mutableArrayField{
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

func (field *mutableArrayField) asProperty(base fieldBase) Property {
	return &arrayProperty{
		field:     field,
		fieldBase: base,
	}
}

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
	property := field.factory.buildField(element)
	return property
}

type leafArrayField struct {
	baseArrayField
}

func (field *leafArrayField) asProperty(base fieldBase) Property {
	return &arrayProperty{
		field:     field,
		fieldBase: base,
	}
}

func (field *leafArrayField) AppendGoValue(value interface{}) error {
	return nil
}

type arrayOfScalarField struct {
	leafArrayField
	itemSchema *j5schema.ScalarSchema
}

var _ ArrayOfScalarField = (*arrayOfScalarField)(nil)

func (field *arrayOfScalarField) AppendGoScalar(val interface{}) error {
	list := field.fieldInParent.getOrCreateMutable().List()
	value, err := scalarReflectFromGo(field.itemSchema.Proto, val)
	if err != nil {
		return err
	}
	list.Append(value)
	return nil
}

type arrayOfEnumField struct {
	leafArrayField
	itemSchema *j5schema.EnumSchema
}

var _ ArrayOfEnumField = (*arrayOfEnumField)(nil)

func (field *arrayOfEnumField) AppendEnumFromString(name string) error {
	option := field.itemSchema.OptionByName(name)
	if option != nil {
		list := field.fieldInParent.getOrCreateMutable().List()
		list.Append(protoreflect.ValueOfEnum(protoreflect.EnumNumber(option.Number)))
		return nil
	}
	return fmt.Errorf("enum value %s not found", name)
}
