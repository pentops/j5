package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/internal/j5schema"
	"github.com/pentops/j5/internal/patherr"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type RangeValuesCallback func(Field) error
type RangePropertiesCallback func(Property) error

type Property interface {
	IsSet() bool
	Schema() *j5schema.ObjectProperty
	Field() (Field, error)
}

type property struct {
	schema    *j5schema.ObjectProperty
	protoPath []protoreflect.FieldDescriptor

	hasValue bool
	value    Field
}

var _ Property = &property{}

func (p *property) IsSet() bool {
	return p.hasValue
}

func (p *property) Schema() *j5schema.ObjectProperty {
	return p.schema
}

func (p *property) Field() (Field, error) {
	if !p.hasValue {
		return nil, fmt.Errorf("field %s is not set", p.schema.JSONName)
	}
	return p.value, nil
}

// PropertySet is implemented by Oneofs, Objects and Maps with String keys.
type PropertySet interface {
	SchemaName() string // Returns the full name of the entity wrapping the properties.

	RangeProperties(RangePropertiesCallback) error
	RangeValues(RangeValuesCallback) error

	// HasProperty returns true if there is a property with the given name in
	// the *schema* for the property set.
	HasProperty(name string) bool

	// GetProperty returns the property with the given name in the schema for
	// the property set. The property may not be set to a value.
	GetProperty(name string) (Property, error)
	GetValue(name string) (Field, bool, error)
	GetPropertyScalar(name string) (ScalarField, error)

	ListPropertyNames() []string

	implementsPropertySet()
}

var _ PropertySet = &propSet{}

type propSet struct {
	asMap    map[string]*property
	asSlice  []*property
	fullName string
	message  protoreflect.Message
}

func newPropSet(name string, message protoreflect.Message, props []*j5schema.ObjectProperty) (*propSet, error) {
	fs := &propSet{
		asMap:    map[string]*property{},
		fullName: name,
		message:  message,
	}

	rootDesc := message.Descriptor()
	if rootDesc == nil {
		return nil, fmt.Errorf("propSet root is not a message")
	}

	for _, propSchema := range props {
		prop := &property{
			schema:    propSchema,
			protoPath: make([]protoreflect.FieldDescriptor, len(propSchema.ProtoField)),
		}

		walk := rootDesc
		for idx, fieldNumber := range propSchema.ProtoField {
			fieldDesc := walk.Fields().ByNumber(fieldNumber)
			if fieldDesc.Kind() != protoreflect.MessageKind {
				return nil, fmt.Errorf("field %s is not a message but has nested types", fieldDesc.FullName())
			}
			prop.protoPath[idx] = fieldDesc
		}

		fs.asMap[propSchema.JSONName] = prop
		fs.asSlice = append(fs.asSlice, prop)
	}

	return fs, nil
}

func (*propSet) implementsPropertySet() {}

func (fs *propSet) SchemaName() string {
	return fs.fullName
}

func (fs *propSet) ListPropertyNames() []string {
	// in order, not using map.
	names := make([]string, 0, len(fs.asSlice))
	for _, prop := range fs.asSlice {
		names = append(names, prop.schema.JSONName)
	}
	return names
}

func (fs *propSet) HasProperty(name string) bool {
	_, ok := fs.asMap[name]
	return ok
}

func (fs *propSet) GetProperty(name string) (Property, error) {
	prop, ok := fs.asMap[name]
	if !ok {
		return nil, fmt.Errorf("%s has no property %s", fs.fullName, name)
	}
	return prop, nil
}

func (fs *propSet) GetValue(name string) (Field, bool, error) {
	return fs.getValue(name, false)
}

func (fs *propSet) GetPropertyScalar(name string) (ScalarField, error) {
	prop, err := fs.GetProperty(name)
	if err != nil {
		return nil, err
	}
	scalar, ok := prop.(ScalarField)
	if !ok {
		return nil, fmt.Errorf("%s is not a scalar field", name)
	}
	return scalar, nil
}

func (fs *propSet) RangeProperties(callback RangePropertiesCallback) error {
	var err error
	for _, prop := range fs.asSlice {
		err = callback(prop)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fs *propSet) RangeValues(callback RangeValuesCallback) error {
	var err error
	for _, prop := range fs.asSlice {
		if prop.hasValue {
			err = callback(prop.value)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (fs *propSet) getValue(name string, create bool) (Field, bool, error) {
	prop, ok := fs.asMap[name]
	if !ok {
		return nil, false, fmt.Errorf("%q has no property %q", fs.fullName, name)
	}

	if prop.value != nil {
		return prop.value, true, nil
	}

	fieldContext := &propertyContext{
		schema: prop.schema,
	}

	if len(prop.protoPath) == 0 {
		// 'create' is meaningless here as the field is 'virtual', in that it is
		// a subset of fields in an already existing message.

		wrapper, ok := prop.schema.Schema.(*j5schema.OneofField)
		if !ok {
			return nil, false, fmt.Errorf("Reflection Bug: no proto field and not a oneof")
		}

		preBuilt, err := newOneof(wrapper.Schema(), fs.message)
		if err != nil {
			return nil, false, patherr.Wrap(err, prop.schema.JSONName)
		}

		protoVal := newProtoMessage(fs.message)
		built := newOneofField(fieldContext, wrapper, protoVal)
		built._oneof = preBuilt
		return built, true, nil
	}

	var walkField protoreflect.FieldDescriptor
	walkMessage := fs.message
	walkPath := prop.protoPath[:]
	for len(walkPath) > 1 {
		walkField, walkPath = walkPath[0], walkPath[1:]
		fieldContext.walkedProtoPath = append(fieldContext.walkedProtoPath, walkField.JSONName())
		fieldValue := walkMessage.Get(walkField)
		if !fieldValue.IsValid() {
			if !create {
				return nil, false, fmt.Errorf("field %s is not set", walkField.FullName())
			}
			fieldValue = walkMessage.Mutable(walkField)
		}
		walkMessage = fieldValue.Message()
		if walkMessage == nil {
			return nil, false, fmt.Errorf("Reflection Bug: field %s is not a message", walkField.FullName())
		}
	}

	finalField := walkPath[0]
	fieldContext.walkedProtoPath = append(fieldContext.walkedProtoPath, finalField.JSONName())

	protoVal := newProtoPair(walkMessage, finalField)

	built, err := buildProperty(fieldContext, prop.schema, protoVal)
	if err != nil {
		return nil, false, patherr.Wrap(err, prop.schema.JSONName)
	}

	prop.hasValue = true
	prop.value = built

	return built, true, nil
}

func buildProperty(context fieldContext, schema *j5schema.ObjectProperty, value *protoPair) (Field, error) {

	ff, err := newFieldFactory(schema.Schema, value.fieldDescriptor())
	if err != nil {
		return nil, err
	}

	switch st := schema.Schema.(type) {

	case *j5schema.ArrayField:
		if !value.fieldInParent.IsList() {
			return nil, fmt.Errorf("Reflection Bug: ArrayField is not a list")
		}

		valVal, err := value.getMutableValue(true)
		if err != nil {
			return nil, err
		}
		listVal := valVal.List()

		field, err := newArrayField(context, st, listVal, ff)
		if err != nil {
			return nil, err
		}

		return field, nil

	case *j5schema.MapField:
		if !value.fieldInParent.IsMap() {
			return nil, fmt.Errorf("MapField is not a map")
		}

		valVal, err := value.getMutableValue(true)
		if err != nil {
			return nil, err
		}
		mapVal := valVal.Map()

		field, err := newMapField(context, st, mapVal, ff)
		if err != nil {
			return nil, err
		}
		return field, nil

	}

	field := ff.buildField(context, value)
	return field, nil

}

func copyReflect(a, b protoreflect.Message) {
	a.Range(func(fd protoreflect.FieldDescriptor, val protoreflect.Value) bool {
		b.Set(fd, val)
		return true
	})
}
