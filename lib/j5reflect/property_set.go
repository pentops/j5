package j5reflect

import (
	"fmt"
	"strings"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/lib/j5schema"
	"github.com/pentops/j5/lib/patherr"
	"google.golang.org/protobuf/reflect/protoreflect"
)

/*** Interface ***/

type RangeValuesCallback func(Field) error
type RangePropertiesCallback func(Property) error
type RangePropertySchemasCallback func(name string, required bool, schema *schema_j5pb.Field) error

type Property interface {
	IsSet() bool
	Schema() *j5schema.ObjectProperty
	CreateField() (Field, error)
	Field() (Field, error)

	IsArray() bool
	IsMap() bool

	PropertyType() PropertyType
}

type PropertyType int

const (
	MapProperty PropertyType = iota
	ArrayProperty
	ObjectProperty
	OneofProperty
	EnumProperty
	ScalarProperty
	AnyProperty
	PolymorphProperty
)

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
	NewValue(name string) (Field, error)

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

type propertyStub struct {
	schema    *j5schema.ObjectProperty
	protoPath []protoreflect.FieldDescriptor
}

func (p *propertyStub) newEmpty(ps *propSet) *property {
	return &property{
		schema:    p.schema,
		protoPath: p.protoPath,
		propSet:   ps,
	}
}

type property struct {
	schema    *j5schema.ObjectProperty
	protoPath []protoreflect.FieldDescriptor
	propSet   *propSet

	hasValue bool
	value    Field
}

var _ Property = &property{}

func (p *property) IsSet() bool {
	if !p.hasValue {
		return false
	}
	return p.value.IsSet()
}

func (p *property) IsArray() bool {
	_, ok := p.schema.Schema.(*j5schema.ArrayField)
	return ok
}

func (p *property) IsMap() bool {
	_, ok := p.schema.Schema.(*j5schema.MapField)
	return ok
}

func (p *property) PropertyType() PropertyType {
	switch p.schema.Schema.(type) {
	case *j5schema.ArrayField:
		return ArrayProperty
	case *j5schema.MapField:
		return MapProperty
	case *j5schema.ObjectField:
		return ObjectProperty
	case *j5schema.OneofField:
		return OneofProperty
	case *j5schema.EnumField:
		return EnumProperty
	case *j5schema.ScalarSchema:
		return ScalarProperty
	case *j5schema.AnyField:
		return AnyProperty
	case *j5schema.PolymorphField:
		return PolymorphProperty
	default:
		return -1
	}
}

func (p *property) Schema() *j5schema.ObjectProperty {
	return p.schema
}

// CreateField assigns a new empty value to the property in the parent property
// set
func (p *property) CreateField() (Field, error) {
	if p.hasValue {
		return nil, fmt.Errorf("field %s is already set", p.schema.JSONName)
	}
	vv, err := p.propSet.buildOrCreate(p)
	if err != nil {
		return nil, err
	}
	p.value = vv
	p.hasValue = true
	return vv, nil
}

func (p *property) Field() (Field, error) {
	if !p.hasValue {
		_, _, err := p.propSet.buildValue(p, false)
		if err != nil {
			return nil, fmt.Errorf("field %s is not set: %w", p.schema.JSONName, err)
		}
	}
	return p.value, nil
}

type propSetFactory struct {
	properties []*propertyStub
	schema     hasProps
}

func (factory propSetFactory) newMessage(msg protoreflect.Message) *propSet {
	ps := &propSet{
		asMap:  map[string]*property{},
		schema: factory.schema,
		value:  msg,
	}

	for _, prop := range factory.properties {
		cloned := prop.newEmpty(ps)

		ps.asMap[prop.schema.JSONName] = cloned
		ps.asSlice = append(ps.asSlice, cloned)
	}

	return ps
}

type hasProps interface {
	FullName() string
	ClientProperties() j5schema.PropertySet

	AllProperties() j5schema.PropertySet
}

//var _ PropertySet = &propSet{}

type propSet struct {
	asMap   map[string]*property
	asSlice []*property
	schema  hasProps

	value protoreflect.Message
}

// ProtoMessage returns the underlying protoreflect message. From there you are
// on your own, the schema may not match.
func (ps *propSet) ProtoReflect() protoreflect.Message {
	return ps.value
}

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

func newPropSet(schema hasProps, rootDesc protoreflect.MessageDescriptor) (propSetFactory, error) {

	if rootDesc == nil {
		return propSetFactory{}, fmt.Errorf("propSet root is not a message")
	}

	if rootDesc.IsMapEntry() {
		return propSetFactory{}, fmt.Errorf("propSet root is a map entry (virtual message)")
	}

	props := clientProperties(schema.AllProperties())

	builtProps := make([]*propertyStub, 0, len(props))
	for _, propSchema := range props {
		prop := &propertyStub{
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
	var value Field
	if prop.value != nil {
		value = prop.value
	} else {
		var err error
		value, err = fs.buildOrCreate(prop)
		if err != nil {
			return nil, patherr.Wrap(err, nameParts[0])
		}
	}
	if isArray {
		if len(nameParts) == 1 {
			return value, nil
		} else {
			array, ok := value.AsArrayOfContainer()
			if !ok {
				return nil, fmt.Errorf("property %s is not an array", nameParts[0])
			}
			container, _ := array.NewContainerElement() // create a new element in the array
			return container.GetOrCreateValue(nameParts[1:]...)
		}
	}

	if len(nameParts) == 1 {
		return value, nil
	}

	container, ok := value.AsContainer()
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

	next, ok, err := fs.getValue(nameParts[0])
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
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

func (fs *propSet) getValue(name string) (Field, bool, error) {
	prop, ok := fs.asMap[name]
	if !ok {
		keys := make([]string, 0, len(fs.asMap))
		for k := range fs.asMap {
			keys = append(keys, k)
		}
		return nil, false, fmt.Errorf("%q has no property %q, has %q", fs.schema.FullName(), name, keys)
	}
	if prop.value != nil {
		return prop.value, true, nil
	}
	return fs.buildValue(prop, false)

}

func (fs *propSet) NewValue(name string) (Field, error) {
	prop, ok := fs.asMap[name]
	if !ok {
		return nil, fmt.Errorf("%q has no property %q", fs.schema.FullName(), name)
	}
	if prop.value != nil {
		return prop.value, fmt.Errorf("field %s is already set", name)
	}
	return fs.buildOrCreate(prop)
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
		err = callback(val)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fs *propSet) buildOrCreate(prop *property) (Field, error) {
	val, _, err := fs.buildValue(prop, true)
	return val, err
}

func (fs *propSet) GetOne() (Field, bool, error) {
	var property Field
	var found bool

	for _, search := range fs.asSlice {
		field, has, err := fs.GetField(search.schema.JSONName)
		if err != nil {
			return nil, false, err
		}
		if !has {
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

func (fs *propSet) buildValue(prop *property, create bool) (Field, bool, error) {

	msg := fs.value
	if msg == nil {
		return nil, false, fmt.Errorf("reflection Bug: no fs.value in buildValue")
	}

	fieldContext := &propertyContext{
		schema: prop.schema,
	}

	if len(prop.protoPath) == 0 {
		wrapper, ok := prop.schema.Schema.(*j5schema.OneofField)
		if !ok {
			return nil, false, fmt.Errorf("reflection Bug: no proto field and not a oneof")
		}

		descriptor := msg.Descriptor()

		propSetFactory, err := newPropSet(wrapper.OneofSchema(), descriptor)
		if err != nil {
			return nil, false, patherr.Wrap(err, prop.schema.JSONName)
		}

		propSet := propSetFactory.newMessage(fs.value)
		oneof := &oneofImpl{
			schema:  wrapper.OneofSchema(),
			propSet: propSet,
		}

		built := newOneofField(fieldContext, wrapper, oneof)
		return built, built.IsSet(), nil
	}

	var walkField protoreflect.FieldDescriptor
	walkMessage := msg
	walkPath := prop.protoPath[:]
	for len(walkPath) > 1 {
		walkField, walkPath = walkPath[0], walkPath[1:]
		fieldContext.walkedProtoPath = append(fieldContext.walkedProtoPath, walkField.JSONName())
		if !create {
			if !walkMessage.Has(walkField) {
				return nil, false, nil
			}
		}
		fieldValue := walkMessage.Mutable(walkField)
		if !fieldValue.IsValid() {
			panic(fmt.Sprintf("Reflection Bug: field %s is not valid", walkField.FullName()))
		}
		walkMessage = fieldValue.Message()
		if walkMessage == nil {
			return nil, false, fmt.Errorf("reflection bug: field %s is not a message", walkField.FullName())
		}
	}

	finalField := walkPath[0]
	if !create {
		if !walkMessage.Has(finalField) {
			return nil, false, nil
		}
	}
	fieldContext.walkedProtoPath = append(fieldContext.walkedProtoPath, finalField.JSONName())

	protoVal := newProtoPair(walkMessage, finalField)

	built, err := buildProperty(fieldContext, prop.schema, protoVal)
	if err != nil {
		return nil, false, err
	}
	prop.value = built

	prop.hasValue = true

	return prop.value, true, nil
}

func buildProperty(context fieldContext, schema *j5schema.ObjectProperty, value *protoPair) (Field, error) {

	switch st := schema.Schema.(type) {

	case *j5schema.ArrayField:
		if !value.fieldInParent.IsList() {
			return nil, fmt.Errorf("reflection bug: ArrayField is not a list")
		}

		valVal, err := value.getMutableValue(true)
		if err != nil {
			return nil, err
		}
		listVal := valVal.List()

		if st.ItemSchema.Mutable() {

			ff, err := newMessageFieldFactory(st.ItemSchema, value.fieldInParent.Message())
			if err != nil {
				return nil, err
			}

			field, err := newMessageArrayField(context, st, listVal, ff)
			if err != nil {
				return nil, err
			}

			return field, nil
		}

		ff, err := newFieldFactory(st.ItemSchema, value.fieldInParent)
		if err != nil {
			return nil, err
		}

		field, err := newLeafArrayField(context, st, listVal, ff)
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

		if st.ItemSchema.Mutable() {
			ff, err := newMessageFieldFactory(st.ItemSchema, value.fieldInParent.MapValue().Message())
			if err != nil {
				return nil, err
			}
			field, err := newMessageMapField(context, st, mapVal, ff)
			if err != nil {
				return nil, err
			}
			return field, nil
		}

		ff, err := newFieldFactory(st.ItemSchema, value.fieldInParent.MapValue())
		if err != nil {
			return nil, err
		}

		field, err := newLeafMapField(context, st, mapVal, ff)
		if err != nil {
			return nil, err
		}
		return field, nil
	}

	if schema.Schema.Mutable() {
		messageVal, err := value.getMutableValue(true)
		if err != nil {
			return nil, err
		}
		message := messageVal.Message()
		ff, err := newMessageFieldFactory(schema.Schema, message.Descriptor())
		if err != nil {
			return nil, err
		}
		field := ff.buildField(context, message)
		return field, nil
	}

	ff, err := newFieldFactory(schema.Schema, value.fieldDescriptor())
	if err != nil {
		return nil, err
	}

	field := ff.buildField(context, value)
	return field, nil

}

func copyReflect(a, b protoreflect.Message) {
	bFields := b.Descriptor().Fields()
	a.Range(func(fd protoreflect.FieldDescriptor, val protoreflect.Value) bool {
		bField := bFields.ByNumber(fd.Number())
		if bField == nil || bField.Kind() != fd.Kind() || bField.Name() != fd.Name() {
			panic(fmt.Sprintf("CopyReflect: field %s not found in %s", fd.FullName(), b.Descriptor().FullName()))
		}
		b.Set(bField, val)
		return true
	})
}

type messageFieldFactory interface {
	buildField(schema fieldContext, value protoreflect.Message) Field
}

type fieldFactory interface {
	buildField(schema fieldContext, value protoContext) Field
}

func newMessageFieldFactory(schema j5schema.FieldSchema, desc protoreflect.MessageDescriptor) (messageFieldFactory, error) {

	switch st := schema.(type) {
	case *j5schema.ObjectField:
		propSetFactory, err := newPropSet(st.ObjectSchema(), desc)
		if err != nil {
			return nil, err
		}
		return &objectFieldFactory{
			schema:  st,
			propSet: propSetFactory,
		}, nil

	case *j5schema.OneofField:
		propSetFactory, err := newPropSet(st.OneofSchema(), desc)
		if err != nil {
			return nil, err
		}
		return &oneofFieldFactory{
			schema:  st,
			propSet: propSetFactory,
		}, nil

	case *j5schema.AnyField:
		return &anyFieldFactory{schema: st}, nil

	case *j5schema.PolymorphField:
		return &polymorphFieldFactory{schema: st}, nil

	default:
		panic(fmt.Sprintf("invalid schema for message field: %T", schema))
	}
}

func newFieldFactory(schema j5schema.FieldSchema, field protoreflect.FieldDescriptor) (fieldFactory, error) {
	switch st := schema.(type) {
	case *j5schema.EnumField:
		if field.Kind() != protoreflect.EnumKind {
			return nil, fmt.Errorf("EnumField is kind %s", field.Kind())
		}
		return &enumFieldFactory{schema: st}, nil

	case *j5schema.ScalarSchema:
		if st.WellKnownTypeName != "" {
			if field.Kind() != protoreflect.MessageKind {
				return nil, fmt.Errorf("ScalarField is proto kind %s, want message for %T", field.Kind(), st.Proto.Type)
			}
			if string(field.Message().FullName()) != string(st.WellKnownTypeName) {
				return nil, fmt.Errorf("ScalarField message is %s, want %s for %T", field.Message().FullName(), st.WellKnownTypeName, st.Proto.Type)
			}
		} else if field.Kind() != st.Kind {
			return nil, fmt.Errorf("ScalarField is proto kind %s, want schema %q for %T", field.Kind(), st.Kind, st.Proto.Type)
		}
		return &scalarFieldFactory{schema: st}, nil

	default:
		panic(fmt.Sprintf("invalid schema for leaf field: %T", schema))
	}
}
