package j5reflect

import (
	"github.com/pentops/j5/internal/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

/*** Interface ***/

type Object interface {
	PropertySet

	// HasAnyValue returns true if any of the properties have a valid value
	HasAnyValue() bool
}

type ObjectField interface {
	Object
	Field
}

type MapOfObjectField interface {
	NewObjectValue(key string) (ObjectField, error)
}

type ArrayOfObjectField interface {
	ArrayOfContainerField
	NewObjectElement() (ObjectField, int)
}

/*** Implementation ***/

type objectImpl struct {
	schema *j5schema.ObjectSchema
	*propSet
}

var _ Object = &objectImpl{}

func (fs *objectImpl) HasAnyValue() bool {
	for _, prop := range fs.asSlice {
		if prop.IsSet() {
			return true
		}
	}
	return false
}

type objectFieldFactory struct {
	schema  *j5schema.ObjectField
	propSet propSetFactory
}

var _ messageFieldFactory = (*objectFieldFactory)(nil)

func (f *objectFieldFactory) buildField(context fieldContext, value protoreflect.Message) Field {
	obj := &objectImpl{
		propSet: f.propSet.linkMessage(value),
		schema:  f.schema.Schema(),
	}
	return newObjectField(context, f.schema, obj)
}

type existingObjectField struct {
	fieldDefaults
	*objectImpl
}

func newObjectField(context fieldContext, schema *j5schema.ObjectField, obj *objectImpl) ObjectField {
	return &existingObjectField{
		objectImpl: obj,
		fieldDefaults: fieldDefaults{
			fieldType: FieldTypeObject,
			context:   context,
		},
	}
}

var _ ObjectField = (*existingObjectField)(nil)

func (obj *existingObjectField) Type() FieldType {
	return FieldTypeObject
}

func (obj *existingObjectField) IsSet() bool {
	return true
}

func (obj *existingObjectField) AsContainer() (PropertySet, bool) {
	return obj, true // returns self.
}

/*** Implement Array Of Object ***/

type arrayOfObjectField struct {
	mutableArrayField
}

var _ ArrayOfObjectField = (*arrayOfObjectField)(nil)

func (field *arrayOfObjectField) NewObjectElement() (ObjectField, int) {
	of := field.NewElement().(ObjectField)
	return of, of.IndexInParent()
}

func (field *arrayOfObjectField) NewContainerElement() (PropertySet, int) {
	of := field.NewElement().(ObjectField)
	return of, of.IndexInParent()
}

func (field *arrayOfObjectField) AsArrayOfContainer() (ArrayOfContainerField, bool) {
	return field, true
}

func (field *arrayOfObjectField) RangeContainers(cb func(PropertySet) error) error {
	return field.RangeValues(func(idx int, f Field) error {
		val, ok := f.(*existingObjectField)
		if !ok {
			return nil
		}
		return cb(val)
	})
}

/*** Implement Map Of Object ***/

type mapOfObjectField struct {
	MutableMapField
}

func (field *mapOfObjectField) NewObjectValue(key string) (ObjectField, error) {
	val, err := field.NewElement(key)
	if err != nil {
		return nil, err
	}
	of := val.(ObjectField)
	return of, nil
}
