package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/internal/j5schema"
	"github.com/pentops/j5/internal/patherr"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// PropertySet is implemented by Oneofs, Objects and Maps with String keys.
type PropertySet interface {
	Name() string // Returns the full name of the entity wrapping the properties.

	RangeProperties(RangeCallback) error
	RangeSetProperties(RangeCallback) error

	// HasProperty returns true if there is a property with the given name in
	// the *schema* for the property set.
	HasProperty(name string) bool

	// GetProperty returns the property with the given name in the schema for
	// the property set. The property may not be set to a value.
	GetProperty(name string) (Property, error)

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
	asMap    map[string]Property
	asSlice  []Property
	fullName string
}

func newPropSet(name string, props []Property) (*propSet, error) {
	fs := &propSet{
		asMap:    map[string]Property{},
		fullName: name,
	}

	for _, prop := range props {
		fs.asMap[prop.JSONName()] = prop
		fs.asSlice = append(fs.asSlice, prop)
	}

	return fs, nil
}

func (*propSet) implementsPropertySet() {}

func (fs *propSet) Name() string {
	return fs.fullName
}

func (fs *propSet) ListPropertyNames() []string {
	names := make([]string, 0, len(fs.asSlice))
	for _, prop := range fs.asSlice {
		names = append(names, prop.JSONName())
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

func collectProperties(properties []*j5schema.ObjectProperty, msg *protoMessageWrapper) ([]Property, error) {
	out := make([]Property, 0)
	var err error

	for _, schema := range properties {
		if len(schema.ProtoField) == 0 {

			// Then we have a 'fake' object, usually an exposed oneof.
			// It shows as a object in clients, but in proto the fields are
			// directly on the parent message.

			fieldBase := fieldBase{
				schema: schema,
			}

			switch wrapper := schema.Schema.(type) {
			case *j5schema.ObjectField:
				preBuilt, err := newObject(wrapper.Schema(), msg)
				if err != nil {
					return nil, patherr.Wrap(err, schema.JSONName)
				}
				built := newObjectField(wrapper, msg.virtualField(preBuilt.propSet))
				built._object = preBuilt
				out = append(out, &objectProperty{
					field:     built,
					fieldBase: fieldBase,
				})

			case *j5schema.OneofField:

				preBuilt, err := newOneof(wrapper.Schema(), msg)
				if err != nil {
					return nil, patherr.Wrap(err, schema.JSONName)
				}
				built := newOneofField(wrapper, msg.virtualField(preBuilt.propSet))
				built._oneof = preBuilt
				out = append(out, &oneofProperty{
					field:     built,
					fieldBase: fieldBase,
				})

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

		built, err := buildProperty(schema, fieldValue)
		if err != nil {
			return nil, patherr.Wrap(err, schema.JSONName)
		}

		out = append(out, built)
	}

	return out, nil
}

func buildProperty(schema *j5schema.ObjectProperty, value *realProtoMessageField) (Property, error) {
	fieldBase := fieldBase{
		schema: schema,
	}
	switch st := schema.Schema.(type) {

	case *j5schema.ArrayField:
		field, err := newArrayField(st, value)
		if err != nil {
			return nil, err
		}

		return &arrayProperty{
			field:     field,
			fieldBase: fieldBase,
		}, nil

	case *j5schema.MapField:
		field, err := newMapField(st, value)
		if err != nil {
			return nil, err
		}
		return &mapProperty{
			field:     field,
			fieldBase: fieldBase,
		}, nil

	}

	ff, err := newFieldFactory(schema.Schema, value.fieldInParent)
	if err != nil {
		return nil, err
	}

	field := ff.buildField(value).asProperty(fieldBase)

	return field, nil

}
