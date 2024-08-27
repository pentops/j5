package j5reflect

import (
	"fmt"

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
	ContainerField

	// Object returns the existing object, or creates a new object up the chain,
	// i.e. it sets default values for all object, oneof and map<string,x> nodes on
	// the way from the refl root
	Object() (Object, error)
}

type MapOfObjectField interface {
	NewObjectValue(key string) (Oneof, error)
}

type ArrayOfObjectField interface {
	ArrayOfContainerField
	NewObjectElement() (Object, int, error)
}

/*** Implementation ***/

type objectImpl struct {
	schema *j5schema.ObjectSchema
	value  protoreflect.Message
	*propSet
}

var _ Object = &objectImpl{}

func newObject(schema *j5schema.ObjectSchema, value protoreflect.Message) (*objectImpl, error) {
	fieldset, err := newPropSet(schema.FullName(), value, schema.ClientProperties())
	if err != nil {
		return nil, err
	}

	return &objectImpl{
		schema:  schema,
		value:   value,
		propSet: fieldset,
	}, nil
}

func (fs *objectImpl) NewValue(propName string) (Field, error) {
	prop, ok := fs.propSet.asMap[propName]
	if !ok {
		return nil, fmt.Errorf("unknown property %s", prop)
	}
	if prop.hasValue {
		return nil, fmt.Errorf("property %s already set", propName)
	}
	val, _, err := fs.propSet.getValue(propName, true)
	return val, err
}

func (fs *objectImpl) HasAnyValue() bool {
	for _, prop := range fs.asSlice {
		if prop.IsSet() {
			return true
		}
	}
	return false
}

type existingObjectField struct {
	fieldDefaults
	object *objectImpl
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

func (obj *existingObjectField) GetExistingContainer() (PropertySet, bool, error) {
	return obj.object, true, nil
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
	value         protoContext
	_object       *objectImpl
	_objectSchema *j5schema.ObjectSchema
}

var _ ObjectField = (*objectField)(nil)

func newObjectField(context fieldContext, fieldSchema *j5schema.ObjectField, value protoContext) *objectField {
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

func (f *objectFieldFactory) buildField(context fieldContext, value protoContext) Field {
	return newObjectField(context, f.schema, value)
}

func (obj *objectField) Type() FieldType {
	return FieldTypeObject
}

func (obj *objectField) IsSet() bool {
	return obj.value.isSet()
}

func (obj *objectField) SetDefault() error {
	_, err := obj.value.getMutableValue(true)
	return err
}

func (obj *objectField) AsContainer() (ContainerField, bool) {
	return obj, true
}

func (obj *objectField) GetOrCreateContainer() (PropertySet, error) {
	val, err := obj.Object()
	if err != nil {
		return nil, err
	}
	_, err = obj.value.getMutableValue(true)
	return val, err
}

func (obj *objectField) GetExistingContainer() (PropertySet, bool, error) {
	if !obj.value.isSet() {
		return nil, false, nil
	}
	val, err := obj.Object()
	if err != nil {
		return nil, false, err
	}
	return val, true, nil
}

func (obj *objectField) Object() (Object, error) {
	if obj._object == nil {
		val, err := obj.value.getMutableValue(true)
		if err != nil {
			return nil, err
		}
		msg := val.Message()

		built, err := newObject(obj._objectSchema, msg)
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

func (field *arrayOfObjectField) NewContainerElement() (ContainerField, int, error) {
	of := field.NewElement().(ObjectField)
	return of, of.IndexInParent(), nil
}

func (field *arrayOfObjectField) AsArrayOfContainer() (ArrayOfContainerField, bool) {
	return field, true
}

func (field *arrayOfObjectField) RangeContainers(cb func(ContainerField, PropertySet) error) error {
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
