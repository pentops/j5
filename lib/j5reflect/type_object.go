package j5reflect

import (
	"github.com/pentops/j5/lib/j5reflect/protoval"
	"github.com/pentops/j5/lib/j5schema"
)

/*** Interface ***/

type Object interface {
	PropertySet

	// HasAnyValue returns true if any of the properties have a valid value
	HasAnyValue() bool
	ObjectSchema() *j5schema.ObjectSchema
	Interface() any
}

type ObjectField interface {
	Object
	Field
}

type MapOfObjectField interface {
	NewObjectElement(key string) (ObjectField, error)
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

func (obj *objectImpl) HasAvailableProperty(name string) bool {
	return obj.HasProperty(name)
}

func (fs *objectImpl) SetDefaultValue() error {
	// messageValue sets the value recursively from the last known value point
	_ = fs.value.MessageValue()
	return nil
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

func (fs *objectImpl) ObjectSchema() *j5schema.ObjectSchema {
	return fs.schema
}

func (fs *objectImpl) RootSchema() (j5schema.RootSchema, bool) {
	return fs.schema, true
}

func (fs *objectImpl) Interface() any {
	val := fs.value.MessageValue()
	return val.Interface()
}

type objectFieldFactory struct {
	schema  *j5schema.ObjectField
	propSet propSetFactory
}

var _ fieldFactory = (*objectFieldFactory)(nil)

func (f *objectFieldFactory) buildField(context fieldContext, value protoval.Value) Field {
	msgValue, ok := value.AsMessage()
	if !ok {
		panic("objectFieldFactory.buildField called with non-message value")
	}
	obj := &objectImpl{
		propSet: f.propSet.buildForMessage(msgValue),
		schema:  f.schema.ObjectSchema(),
	}
	return newObjectField(context, obj)
}

type objectField struct {
	fieldDefaults
	fieldContext
	*objectImpl
}

func newObjectField(context fieldContext, obj *objectImpl) ObjectField {
	return &objectField{
		objectImpl:   obj,
		fieldContext: context,
	}
}

func (obj *objectField) IsSet() bool {
	return obj.value.IsSet()
}

/*** Explicitly Implements ***/

func (obj *objectField) AsContainer() (ContainerField, bool) {
	return obj, true
}

func (obj *objectField) AsObject() (ObjectField, bool) {
	return obj, true
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

func (field *arrayOfObjectField) NewContainerElement() (ContainerField, int) {
	of := field.NewElement().(ObjectField)
	return of, of.IndexInParent()
}

func (field *arrayOfObjectField) AsArray() (ArrayField, bool) {
	return field, true
}

func (field *arrayOfObjectField) AsArrayOfContainer() (ArrayOfContainerField, bool) {
	return field, true
}

func (field *arrayOfObjectField) AsArrayOfObject() (ArrayOfObjectField, bool) {
	return field, true
}

func (field *arrayOfObjectField) RangeContainers(cb func(int, ContainerField) error) error {
	return field.RangeValues(func(idx int, f Field) error {
		val, ok := f.(*objectField)
		if !ok {
			return nil
		}
		return cb(idx, val)
	})
}

/*** Implement Map Of Object ***/

type mapOfObjectField struct {
	MutableMapField
}

func (field *mapOfObjectField) AsMap() (MapField, bool) {
	return field, true
}

func (field *mapOfObjectField) AsMapOfObject() (MapOfObjectField, bool) {
	return field, true
}

func (field *mapOfObjectField) NewObjectElement(key string) (ObjectField, error) {
	val, err := field.NewElement(key)
	if err != nil {
		return nil, err
	}
	of := val.(ObjectField)
	return of, nil
}

func (field *mapOfObjectField) AsMapOfContainer() (MapOfContainerField, bool) {
	return field, true
}

func (field *mapOfObjectField) NewContainerElement(key string) (ContainerField, error) {
	return field.NewObjectElement(key)
}
