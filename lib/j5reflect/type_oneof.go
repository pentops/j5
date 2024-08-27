package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/internal/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

/*** Interface ***/

type Oneof interface {
	PropertySet

	// GetOne returns the value existing in the oneof, or false if nothing is set.
	GetOne() (Field, bool, error)
}

type OneofField interface {
	Oneof
	Field
}

type MapOfOneofField interface {
	NewOneofValue(key string) (OneofField, error)
}

type ArrayOfOneofField interface {
	ArrayOfContainerField
	NewOneofElement() (Oneof, int, error)
}

/*** Implementation ***/

type oneofImpl struct {
	schema *j5schema.OneofSchema
	*propSet
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

type oneofField struct {
	fieldDefaults
	schema *j5schema.OneofField
	oneofImpl
}

type oneofFieldFactory struct {
	schema  *j5schema.OneofField
	propSet propSetFactory
}

var _ messageFieldFactory = (*oneofFieldFactory)(nil)

func (f *oneofFieldFactory) buildField(context fieldContext, value protoreflect.Message) Field {
	oneof := &oneofImpl{
		schema:  f.schema.Schema(),
		propSet: f.propSet.linkMessage(value),
	}
	return newOneofField(context, f.schema, oneof)
}

var _ OneofField = (*oneofField)(nil)

func newOneofField(context fieldContext, schema *j5schema.OneofField, value *oneofImpl) *oneofField {
	return &oneofField{
		fieldDefaults: fieldDefaults{
			fieldType: FieldTypeOneof,
			context:   context,
		},
		oneofImpl: *value,
		schema:    schema,
	}
}

func (field *oneofField) AsContainer() (PropertySet, bool) {
	return field, true
}

func (field *oneofField) IsSet() bool {
	return true
}

func (field *oneofField) Type() FieldType {
	return FieldTypeOneof
}

/*** Implement Array Of Oneof ***/

type arrayOfOneofField struct {
	mutableArrayField
}

var _ ArrayOfOneofField = (*arrayOfOneofField)(nil)

func (field *arrayOfOneofField) NewOneofElement() (Oneof, int, error) {
	of := field.NewElement().(OneofField)
	return of, of.IndexInParent(), nil
}

func (field *arrayOfOneofField) NewContainerElement() (PropertySet, int) {
	of := field.NewElement().(OneofField)
	return of, of.IndexInParent()
}

func (field *arrayOfOneofField) AsArrayOfContainer() (ArrayOfContainerField, bool) {
	return field, true
}

func (field *arrayOfOneofField) RangeContainers(cb func(PropertySet) error) error {
	return field.RangeValues(func(idx int, f Field) error {
		val, ok := f.(*oneofField)
		if !ok {
			return nil
		}
		return cb(val)
	})
}

/*** Implement Map Of Oneof ***/

type mapOfOneofField struct {
	MutableMapField
}

func (field *mapOfOneofField) NewOneofValue(key string) (OneofField, error) {
	val, err := field.NewElement(key)
	if err != nil {
		return nil, err
	}
	of := val.(OneofField)
	return of, nil
}
