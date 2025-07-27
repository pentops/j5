package j5reflect

import (
	"fmt"
	"strings"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/lib/j5reflect/protoval"
	"github.com/pentops/j5/lib/j5schema"
	"github.com/pentops/j5/lib/patherr"
	"google.golang.org/protobuf/reflect/protoreflect"
)

/*** Interface ***/

type RangeValuesCallback func(Field) error
type RangePropertiesCallback func(Property) error
type RangePropertySchemasCallback func(name string, required bool, schema *schema_j5pb.Field) error

// PropertySet is implemented by Oneofs, Objects and Maps with String keys.
type PropertySet interface {
	SchemaName() string // Returns the full name of the entity wrapping the properties.

	ContainerSchema() j5schema.Container
	RootSchema() (j5schema.RootSchema, bool)

	RangeProperties(RangePropertiesCallback) error
	RangePropertySchemas(RangePropertySchemasCallback) error
	RangeValues(RangeValuesCallback) error

	// HasProperty returns true if there is a property with the given name in
	// the *schema* for the property set.
	HasProperty(name string) bool

	// HasAvailableProperty is like HasProperty, but returns false in a oneof
	// which is already set to another value
	HasAvailableProperty(name string) bool

	// GetProperty returns the property with the given name in the schema for
	// the property set. The property may not be set to a value.
	GetProperty(name string) (Property, error)

	GetField(name ...string) (Field, bool, error)
	GetOrCreateValue(path ...string) (Field, error)
	SetScalar(value any, path ...string) error
	//NewValue(name string) (Field, error)

	ListPropertyNames() []string

	// ProtoMessage returns the underlying protoreflect message. From there you are
	// on your own, the schema may not match.
	ProtoReflect() protoreflect.Message

	implementsPropertySet()
}

type ContainerField interface {
	PropertySet
	Field
}

/*** Implementation ***/

type clientProperty struct {
	fullPath []string // the proto-json path to the client property, to walk flattened messages
	*j5schema.ObjectProperty
}

func clientProperties(sourceSet j5schema.PropertySet) []clientProperty {
	properties := make([]clientProperty, 0, len(sourceSet))
	for _, prop := range sourceSet {
		switch propType := prop.Schema.(type) {
		case *j5schema.ObjectField:
			if propType.Flatten {
				children := clientProperties(propType.ObjectSchema().AllProperties())
				for _, child := range children {
					childPath := []string{prop.JSONName}
					childPath = append(childPath, child.fullPath...)
					properties = append(properties, clientProperty{
						ObjectProperty: child.ObjectProperty,
						fullPath:       childPath,
					})
				}
				continue
			}
		}
		properties = append(properties, clientProperty{
			ObjectProperty: prop,
			fullPath:       []string{prop.JSONName},
		})
	}
	return properties

}

type propSetFactory struct {
	properties []*propertySchema
	schema     propSetSchema
}

func newPropSetFactory(schema propSetSchema, rootDesc protoreflect.MessageDescriptor) (propSetFactory, error) {
	if rootDesc == nil {
		return propSetFactory{}, fmt.Errorf("propSet root is not a message")
	}

	if rootDesc.IsMapEntry() {
		return propSetFactory{}, fmt.Errorf("propSet root is a map entry (virtual message)")
	}

	props := clientProperties(schema.AllProperties())

	builtProps := make([]*propertySchema, 0, len(props))
	for _, propSchema := range props {
		prop := &propertySchema{
			schema:    propSchema.ObjectProperty,
			protoPath: make([]protoreflect.FieldDescriptor, 0),
		}

		rootDesc.Fields().ByJSONName(propSchema.JSONName)

		walk := rootDesc
		for idx, elem := range propSchema.fullPath {
			fieldDesc := walk.Fields().ByJSONName(elem)
			if fieldDesc == nil {
				return propSetFactory{}, fmt.Errorf("newPropSet: field %s not found in %s", elem, walk.FullName())
			}
			prop.protoPath = append(prop.protoPath, fieldDesc)
			if idx < len(propSchema.fullPath)-1 {
				if fieldDesc.Kind() != protoreflect.MessageKind {
					return propSetFactory{}, fmt.Errorf("field %s is not a message but has nested types", fieldDesc.FullName())
				}
				walk = fieldDesc.Message()
			}
		}

		builtProps = append(builtProps, prop)
	}

	return propSetFactory{
		schema:     schema,
		properties: builtProps,
	}, nil
}

func (factory propSetFactory) buildForMessage(msg protoval.MessageValue) *propSet {
	ps := &propSet{
		asMap:  map[string]*property{},
		schema: factory.schema,
		value:  msg,
	}

	for _, propSchema := range factory.properties {
		property := propSchema.newProperty(ps)
		ps.asMap[propSchema.schema.JSONName] = property
		ps.asSlice = append(ps.asSlice, property)
	}

	return ps
}

type propSetSchema interface {
	FullName() string
	ClientProperties() j5schema.PropertySet
	AllProperties() j5schema.PropertySet
}

type propSet struct {
	asMap   map[string]*property
	asSlice []*property
	schema  propSetSchema

	value protoval.MessageValue
}

// ProtoMessage returns the underlying protoreflect message. From there you are
// on your own, the schema may not match.
func (ps *propSet) ProtoReflect() protoreflect.Message {
	val, ok := ps.value.MaybeMessageValue()
	if !ok {
		return nil
	}
	return val
}

func (*propSet) implementsPropertySet() {}

func (fs *propSet) SchemaName() string {
	return fs.schema.FullName()
}

func (fs *propSet) ContainerSchema() j5schema.Container {
	return j5schema.PropertySet(fs.schema.ClientProperties())
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
		return nil, fmt.Errorf("%s has no property %s", fs.schema.FullName(), name)
	}
	return prop, nil
}

func (fs *propSet) GetOrCreateValue(nameParts ...string) (Field, error) {
	if len(nameParts) == 0 {
		return nil, fmt.Errorf("GetOrCreateValue requires at least one name part")
	}
	namePart := nameParts[0]
	isArray := false
	if strings.HasSuffix(namePart, "[]") {
		namePart = strings.TrimSuffix(namePart, "[]")
		isArray = true
	}
	prop, ok := fs.asMap[namePart]
	if !ok {
		return nil, fmt.Errorf("%s has no property %s", fs.schema.FullName(), nameParts[0])
	}
	field, err := prop.Field()
	if err != nil {
		return nil, patherr.Wrap(err, nameParts[0])
	}
	if !field.IsSet() {
		err = field.SetDefaultValue()
		if err != nil {
			return nil, patherr.Wrap(err, nameParts[0])
		}
	}
	if isArray {
		if len(nameParts) == 1 {
			return field, nil
		} else {
			array, ok := field.AsArrayOfContainer()
			if !ok {
				return nil, fmt.Errorf("property %s is not an array", nameParts[0])
			}
			container, _ := array.NewContainerElement() // create a new element in the array
			return container.GetOrCreateValue(nameParts[1:]...)
		}
	}

	if len(nameParts) == 1 {
		return field, nil
	}

	container, ok := field.AsContainer()
	if !ok {
		return nil, fmt.Errorf("property %s is not a container", nameParts[0])
	}
	return container.GetOrCreateValue(nameParts[1:]...)
}

func (fs *propSet) SetScalar(value any, nameParts ...string) error {

	reflectField, err := fs.GetOrCreateValue(nameParts...)
	if err != nil {
		return patherr.Wrap(err, nameParts...)
	}

	scalar, ok := reflectField.AsScalar()
	if ok {
		return scalar.SetGoValue(value)
	}
	asScalar, ok := reflectField.AsArrayOfScalar()
	if !ok {
		return fmt.Errorf("property %s is not a scalar, or array of scalar", reflectField.FullTypeName())
	}
	_, err = asScalar.AppendGoValue(value)
	return err
}

func (fs *propSet) GetField(nameParts ...string) (Field, bool, error) {

	next, err := fs.getValue(nameParts[0])
	if err != nil {
		return nil, false, err
	}
	if len(nameParts) == 1 {
		return next, true, nil
	}

	container, ok := next.AsContainer()
	if !ok {
		return nil, false, fmt.Errorf("property %s is not a container", nameParts[0])
	}
	return container.GetField(nameParts[1:]...)
}

func (fs *propSet) getValue(name string) (Field, error) {
	prop, ok := fs.asMap[name]
	if !ok {
		keys := make([]string, 0, len(fs.asMap))
		for k := range fs.asMap {
			keys = append(keys, k)
		}
		return nil, fmt.Errorf("%q has no property %q, has %q", fs.schema.FullName(), name, keys)
	}
	return prop.Field()
}

/*
func (fs *propSet) NewValue(name string) (Field, error) {
	prop, ok := fs.asMap[name]
	if !ok {
		return nil, fmt.Errorf("%q has no property %q", fs.schema.FullName(), name)
	}
	if prop.field != nil {
		return prop.field, fmt.Errorf("field %s is already set", name)
	}
	return fs.buildOrCreate(prop)
}*/

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

func (fs *propSet) RangePropertySchemas(callback RangePropertySchemasCallback) error {
	var err error
	for _, prop := range fs.asSlice {
		err = callback(prop.schema.JSONName, prop.schema.Required, prop.schema.ToJ5Proto().Schema)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fs *propSet) RangeValues(callback RangeValuesCallback) error {
	for _, prop := range fs.asSlice {
		val, has, err := fs.GetField(prop.schema.JSONName)
		if err != nil {
			return err
		}
		if !has {
			continue
		}
		if !val.IsSet() {
			continue
		}

		if err := callback(val); err != nil {
			return err
		}
	}
	return nil
}

func (fs *propSet) GetOne() (Field, bool, error) {
	var property Field
	var found bool

	for _, search := range fs.asSlice {
		prop, err := fs.GetProperty(search.schema.JSONName)
		if err != nil {
			return nil, false, err
		}

		field, err := prop.Field()
		if err != nil {
			return nil, false, err
		}

		if !field.IsSet() {
			continue
		}

		if found {
			return nil, true, fmt.Errorf("multiple values set for oneof")
		}
		property = field
		found = true
	}
	return property, found, nil
}
