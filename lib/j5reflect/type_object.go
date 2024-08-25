package j5reflect

import (
	"github.com/pentops/j5/internal/j5schema"
)

type Object interface {
	PropertySet

	// HasAnyValue returns true if any of the properties have a valid value
	HasAnyValue() bool
}

type ObjectField interface {
	ContainerField

	// Object returns the existing object, or creates a new object up the chain,
	// i.e. it sets default values for all object, oneof and map<string,x> nodes on
	// the way from the refl root
	Object() (Object, error)
}

type ObjectImpl struct {
	schema *j5schema.ObjectSchema
	value  *protoMessageWrapper
	*propSet
}

var _ Object = &ObjectImpl{}

func newObject(schema *j5schema.ObjectSchema, value *protoMessageWrapper) (*ObjectImpl, error) {

	props, err := collectProperties(schema.ClientProperties(), value)
	if err != nil {
		return nil, err
	}
	fieldset, err := newPropSet(schema.FullName(), props)
	if err != nil {
		return nil, err
	}

	return &ObjectImpl{
		schema:  schema,
		value:   value,
		propSet: fieldset,
	}, nil
}

func (fs *ObjectImpl) asDetachedField() (Field, error) {
	return &existingObjectField{
		object: fs,
	}, nil
}

func (fs *ObjectImpl) HasAnyValue() bool {
	for _, prop := range fs.asSlice {
		if prop.IsSet() {
			return true
		}
	}
	return false
}

type existingObjectField struct {
	fieldDefaults
	object *ObjectImpl
}

func (obj *existingObjectField) Type() FieldType {
	return FieldTypeObject
}

func (obj *existingObjectField) IsSet() bool {
	return true
}

func (obj *existingObjectField) IsContainer() bool {
	return true
}

func (obj *existingObjectField) AsContainer() (ContainerField, bool) {
	return obj, true // returns self.
}

func (obj *existingObjectField) GetOrCreateContainer() (PropertySet, error) {
	return obj.object, nil
}

func (obj *existingObjectField) Object() (Object, error) {
	return obj.object, nil
}

func (obj *existingObjectField) SetDefault() error {
	// already exists, nothing to do.
	return nil
}

var _ ObjectField = (*existingObjectField)(nil)

type objectField struct {
	fieldDefaults
	value         protoValueContext
	_object       *ObjectImpl
	_objectSchema *j5schema.ObjectSchema
}

var _ ObjectField = (*objectField)(nil)

func newObjectField(context fieldContext, fieldSchema *j5schema.ObjectField, value protoValueContext) *objectField {
	of := &objectField{
		fieldDefaults: fieldDefaults{
			fieldType: FieldTypeObject,
			context:   context,
		},
		value:         value,
		_objectSchema: fieldSchema.Schema(),
	}
	return of
}

type objectFieldFactory struct {
	schema *j5schema.ObjectField
}

func (f *objectFieldFactory) buildField(context fieldContext, value protoValueContext) Field {
	return newObjectField(context, f.schema, value)
}

func (obj *objectField) Type() FieldType {
	return FieldTypeObject
}

func (obj *objectField) IsSet() bool {
	return obj.value.isSet()
}

func (obj *objectField) SetDefault() error {
	_ = obj.value.getOrCreateMutable()
	return nil
}

func (obj *objectField) AsContainer() (ContainerField, bool) {
	return obj, true
}

func (obj *objectField) GetOrCreateContainer() (PropertySet, error) {
	val, err := obj.Object()
	if err != nil {
		return nil, err
	}
	_ = obj.value.getOrCreateMutable()
	return val, nil
}

func (obj *objectField) Object() (Object, error) {
	if obj._object == nil {
		msgChild, err := obj.value.getOrCreateChildMessage()
		if err != nil {
			return nil, err
		}
		built, err := newObject(obj._objectSchema, msgChild)
		if err != nil {
			return nil, err
		}
		obj._object = built
	}
	return obj._object, nil
}

type arrayOfObjectField struct {
	mutableArrayField
}

var _ ArrayOfObjectField = (*arrayOfObjectField)(nil)

func (field *arrayOfObjectField) NewObjectElement() (Object, int, error) {
	of := field.NewElement().(ObjectField)
	ofb, err := of.Object()
	if err != nil {
		return nil, -1, err
	}
	return ofb, of.IndexInParent(), nil
}

func (field *arrayOfObjectField) NewContainerElement() (PropertySet, int, error) {
	return field.NewObjectElement()
}

func (field *arrayOfObjectField) AsArrayOfContainer() (ArrayOfContainerField, bool) {
	return field, true
}
