package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/internal/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

/*** Interface ***/

type Oneof interface {
	PropertySet
	GetOne() (Field, bool, error)
}

type MapOfOneofField interface {
	NewOneofValue(key string) (*oneofImpl, error)
}

type ArrayOfOneofField interface {
	ArrayOfContainerField
	NewOneofElement() (Oneof, int, error)
}

/*** Implementation ***/

type oneofImpl struct {
	schema  *j5schema.OneofSchema
	message protoreflect.Message
	*propSet
}

func newOneof(schema *j5schema.OneofSchema, message protoreflect.Message) (*oneofImpl, error) {
	fieldset, err := newPropSet(schema.FullName(), message, schema.Properties)
	if err != nil {
		return nil, err
	}
	return &oneofImpl{
		schema:  schema,
		message: message,
		propSet: fieldset,
	}, nil
}

func (fs *oneofImpl) GetOne() (Field, bool, error) {
	var property Field
	var found bool

	for _, search := range fs.asSlice {
		if search.hasValue {
			if found {
				return nil, true, fmt.Errorf("multiple values set for oneof")
			}
			property = search.value
			found = true
		}
	}
	return property, found, nil
}

type OneofField interface {
	ContainerField

	Oneof() (Oneof, error)
}

type oneofField struct {
	fieldDefaults
	value  protoContext
	schema *j5schema.OneofField
	_oneof *oneofImpl
}

type oneofFieldFactory struct {
	schema *j5schema.OneofField
}

var _ fieldFactory = (*oneofFieldFactory)(nil)

func (f *oneofFieldFactory) buildField(context fieldContext, value protoContext) Field {
	return newOneofField(context, f.schema, value)
}

var _ OneofField = (*oneofField)(nil)

func newOneofField(context fieldContext, schema *j5schema.OneofField, value protoContext) *oneofField {
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

func (field *oneofField) GetExistingContainer() (PropertySet, bool, error) {
	if !field.IsSet() {
		return nil, false, nil
	}
	oneof, err := field.Oneof()
	if err != nil {
		return nil, false, err
	}
	return oneof, true, nil
}

func (field *oneofField) IsSet() bool {
	return field.value.isSet()
}

func (field *oneofField) Type() FieldType {
	return FieldTypeOneof
}

func (field *oneofField) Oneof() (Oneof, error) {
	if field._oneof == nil {
		val, err := field.value.getMutableValue(true)
		if err != nil {
			return nil, err
		}
		msg := val.Message()

		built, err := newOneof(field.schema.Schema(), msg)
		if err != nil {
			return nil, err
		}
		field._oneof = built
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

func (field *arrayOfOneofField) NewContainerElement() (ContainerField, int, error) {
	of := field.NewElement().(OneofField)
	return of, of.IndexInParent(), nil
}

func (field *arrayOfOneofField) AsArrayOfContainer() (ArrayOfContainerField, bool) {
	return field, true
}

func (field *arrayOfOneofField) RangeContainers(cb func(ContainerField, PropertySet) error) error {
	return field.RangeValues(func(idx int, f Field) error {
		val, ok := f.(ContainerField)
		if !ok {
			return nil
		}
		valContainer, ok, err := val.GetExistingContainer()
		if err != nil {
			return err
		}

		if !ok {
			return fmt.Errorf("Reflect Internal Error: expected container field to be set")
		}
		return cb(val, valContainer)
	})
}
