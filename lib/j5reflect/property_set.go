package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/internal/j5schema"
	"github.com/pentops/j5/internal/patherr"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// PropertySet is implemented by Oneofs, Objects and Maps with String keys.
type PropertySet interface {
	SchemaName() string // Returns the full name of the entity wrapping the properties.

	RangeProperties(RangeCallback) error
	RangeSetProperties(RangeCallback) error

	// HasProperty returns true if there is a property with the given name in
	// the *schema* for the property set.
	HasProperty(name string) bool

	// GetProperty returns the property with the given name in the schema for
	// the property set. The property may not be set to a value.
	GetProperty(name string) (Field, error)

	ListPropertyNames() []string

	implementsPropertySet()
}

func copyReflect(a, b protoreflect.Message) {
	a.Range(func(fd protoreflect.FieldDescriptor, val protoreflect.Value) bool {
		b.Set(fd, val)
		return true
	})
}

type propSet struct {
	asMap    map[string]Field
	asSlice  []Field
	fullName string
}

func newPropSet(name string, props []Field) (*propSet, error) {
	fs := &propSet{
		asMap:    map[string]Field{},
		fullName: name,
	}

	for _, prop := range props {
		fs.asMap[prop.NameInParent()] = prop
		fs.asSlice = append(fs.asSlice, prop)
	}

	return fs, nil
}

func (*propSet) implementsPropertySet() {}

func (fs *propSet) SchemaName() string {
	return fs.fullName
}

func (fs *propSet) ListPropertyNames() []string {
	names := make([]string, 0, len(fs.asSlice))
	for _, prop := range fs.asSlice {
		names = append(names, prop.NameInParent())
	}
	return names

}

func (fs *propSet) HasProperty(name string) bool {
	_, ok := fs.asMap[name]
	return ok
}

func (fs *propSet) GetProperty(name string) (Field, error) {
	prop, ok := fs.asMap[name]
	if !ok {
		return nil, fmt.Errorf("%s has no property %s", fs.fullName, name)
	}
	return prop, nil
}

func (fs *propSet) RangeProperties(callback RangeCallback) error {
	var err error
	for _, prop := range fs.asSlice {
		err = callback(prop)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fs *propSet) RangeSetProperties(callback RangeCallback) error {
	var err error
	for _, prop := range fs.asSlice {
		if prop.IsSet() {
			err = callback(prop)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func collectProperties(properties []*j5schema.ObjectProperty, msg *protoMessageWrapper) ([]Field, error) {
	out := make([]Field, 0)
	var err error

	for _, schema := range properties {
		fieldContext := &propertyContext{
			schema: schema,
		}

		if len(schema.ProtoField) == 0 {

			// Then we have a 'fake' object, usually an exposed oneof.
			// It shows as a object in clients, but in proto the fields are
			// directly on the parent message.

			switch wrapper := schema.Schema.(type) {
			case *j5schema.ObjectField:
				preBuilt, err := newObject(wrapper.Schema(), msg)
				if err != nil {
					return nil, patherr.Wrap(err, schema.JSONName)
				}
				built := newObjectField(fieldContext, wrapper, msg.virtualField(preBuilt.propSet))
				built._object = preBuilt

				out = append(out, built)

			case *j5schema.OneofField:

				preBuilt, err := newOneof(wrapper.Schema(), msg)
				if err != nil {
					return nil, patherr.Wrap(err, schema.JSONName)
				}
				built := newOneofField(fieldContext, wrapper, msg.virtualField(preBuilt.propSet))
				built._oneof = preBuilt
				out = append(out, built)

			default:
				return nil, fmt.Errorf("unsupported schema type %T for nested json", wrapper)
			}

			continue
		}

		var walkFieldNumber protoreflect.FieldNumber
		walkPath := schema.ProtoField[:]
		//fmt.Printf("property: %v\n", schema.JSONName)
		walkMessage := msg
		for len(walkPath) > 1 {
			//fmt.Printf("walkPath: %v\n", walkPath)
			walkFieldNumber, walkPath = walkPath[0], walkPath[1:]
			fd := walkMessage.descriptor.Fields().ByNumber(walkFieldNumber)
			if fd.Kind() != protoreflect.MessageKind {
				return nil, fmt.Errorf("field %s is not a message but has nested types", fd.FullName())
			}
			fieldContext.walkedProtoPath = append(fieldContext.walkedProtoPath, fd.JSONName())
			walkMessage, err = walkMessage.fieldAsWrapper(fd)
			if err != nil {
				return nil, err
			}
		}
		walkFieldNumber = walkPath[0]
		fieldValue, err := walkMessage.fieldByNumber(walkFieldNumber)
		if err != nil {
			return nil, err
		}
		fieldContext.walkedProtoPath = append(fieldContext.walkedProtoPath, fieldValue.fieldInParent.JSONName())

		built, err := buildProperty(fieldContext, schema, fieldValue)
		if err != nil {
			return nil, patherr.Wrap(err, schema.JSONName)
		}

		out = append(out, built)
	}

	return out, nil
}

func buildProperty(context fieldContext, schema *j5schema.ObjectProperty, value *realProtoMessageField) (Field, error) {
	switch st := schema.Schema.(type) {

	case *j5schema.ArrayField:
		field, err := newArrayField(context, st, value)
		if err != nil {
			return nil, err
		}

		return field, nil

	case *j5schema.MapField:
		field, err := newMapField(context, st, value)
		if err != nil {
			return nil, err
		}
		return field, nil

	}

	ff, err := newFieldFactory(schema.Schema, value.fieldInParent)
	if err != nil {
		return nil, err
	}
	field := ff.buildField(context, value)
	return field, nil

}
