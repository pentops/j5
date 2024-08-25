package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/internal/j5schema"
)

type Oneof interface {
	PropertySet
	GetOne() (Field, error)
}

type OneofImpl struct {
	schema *j5schema.OneofSchema
	value  *protoMessageWrapper
	*propSet
}

func newOneof(schema *j5schema.OneofSchema, value *protoMessageWrapper) (*OneofImpl, error) {

	props, err := collectProperties(schema.Properties, value)
	if err != nil {
		return nil, err
	}

	fieldset, err := newPropSet(schema.FullName(), props)
	if err != nil {
		return nil, err
	}
	return &OneofImpl{
		schema:  schema,
		value:   value,
		propSet: fieldset,
	}, nil
}

func (fs *OneofImpl) GetOne() (Field, error) {
	var property Field
	for _, prop := range fs.asSlice {
		if prop.IsSet() {
			if property != nil {
				return nil, fmt.Errorf("multiple values set for oneof")
			}
			property = prop
		}
	}
	return property, nil
}

type OneofField interface {
	ContainerField

	Oneof() (Oneof, error)
}

type oneofField struct {
	fieldDefaults
	value  protoValueContext
	schema *j5schema.OneofField
	_oneof *OneofImpl
}

type oneofFieldFactory struct {
	schema *j5schema.OneofField
}

var _ fieldFactory = (*oneofFieldFactory)(nil)

func (f *oneofFieldFactory) buildField(context fieldContext, value protoValueContext) Field {
	return newOneofField(context, f.schema, value)
}

var _ OneofField = (*oneofField)(nil)

func newOneofField(context fieldContext, schema *j5schema.OneofField, value protoValueContext) *oneofField {
	return &oneofField{
		fieldDefaults: fieldDefaults{
			fieldType: FieldTypeOneof,
			context:   context,
		},
		value:  value,
		schema: schema,
	}
}

func (field *oneofField) AsContainer() (ContainerField, bool) {
	return field, true
}

func (field *oneofField) GetOrCreateContainer() (PropertySet, error) {
	oneof, err := field.Oneof()
	if err != nil {
		return nil, err
	}
	return oneof, nil
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

func (field *oneofField) Oneof() (Oneof, error) {
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

type arrayOfOneofField struct {
	mutableArrayField
}

func (field *arrayOfOneofField) NewOneofElement() (Oneof, int, error) {
	of := field.NewElement().(OneofField)
	ofb, err := of.Oneof()
	if err != nil {
		return nil, -1, err
	}
	return ofb, of.IndexInParent(), nil
}

var _ ArrayOfOneofField = (*arrayOfOneofField)(nil)

func (field *arrayOfOneofField) NewContainerElement() (PropertySet, int, error) {
	return field.NewOneofElement()
}

func (field *arrayOfOneofField) AsArrayOfContainer() (ArrayOfContainerField, bool) {
	return field, true
}
