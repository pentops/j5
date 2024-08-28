package j5reflect

import (
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

type oneofField struct {
	fieldDefaults
	fieldContext
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
		propSet: f.propSet.newMessage(value),
	}
	return newOneofField(context, f.schema, oneof)
}

func newOneofField(context fieldContext, schema *j5schema.OneofField, value *oneofImpl) *oneofField {
	return &oneofField{
		fieldContext: context,
		oneofImpl:    *value,
		schema:       schema,
	}
}

func (field *oneofField) IsSet() bool {
	_, ok, err := field.oneofImpl.GetOne()
	return ok && err == nil
}

/*** Explicitly Implements ***/

func (field *oneofField) AsContainer() (ContainerField, bool) {
	return field, true
}

func (field *oneofField) AsOneof() (OneofField, bool) {
	return field, true
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

func (field *arrayOfOneofField) NewContainerElement() (ContainerField, int) {
	of := field.NewElement().(OneofField)
	return of, of.IndexInParent()
}

func (field *arrayOfOneofField) AsArrayOfContainer() (ArrayOfContainerField, bool) {
	return field, true
}

func (field *arrayOfOneofField) RangeContainers(cb func(int, ContainerField) error) error {
	return field.RangeValues(func(idx int, f Field) error {
		val, ok := f.(*oneofField)
		if !ok {
			return nil
		}
		return cb(idx, val)
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
