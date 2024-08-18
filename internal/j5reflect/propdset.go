package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/internal/j5schema"
	"github.com/pentops/j5/internal/patherr"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type RangeCallback func(Property) error

type PropertySet interface {
	RangeProperties(RangeCallback) error
	RangeSetProperties(RangeCallback) error
	GetOne() (Property, error)

	// AnySet returns true if any of the properties have a value
	AnySet() bool
}

func copyReflect(a, b protoreflect.Message) {
	a.Range(func(fd protoreflect.FieldDescriptor, val protoreflect.Value) bool {
		b.Set(fd, val)
		return true
	})
}

type propSet struct {
	asMap   map[string]Property
	asSlice []Property
}

func newPropSet(props []Property) (*propSet, error) {
	fs := &propSet{
		asMap: map[string]Property{},
	}

	for _, prop := range props {
		fs.asMap[prop.JSONName()] = prop
		fs.asSlice = append(fs.asSlice, prop)
	}

	return fs, nil
}

func (fs *propSet) GetProperty(name string) Property {
	prop := fs.asMap[name]
	return prop
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

func (fs *propSet) AnySet() bool {
	for _, prop := range fs.asSlice {
		if prop.IsSet() {
			return true
		}
	}
	return false
}

func (fs *propSet) GetOne() (Property, error) {
	var property Property
	for _, prop := range fs.asSlice {
		if prop.IsSet() {
			if property != nil {
				return nil, fmt.Errorf("multiple values set for oneof")
			}
			property = prop
		}
	}
	return property, nil
}

func collectProperties(properties []*j5schema.ObjectProperty, msg *protoMessage) ([]Property, error) {
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
					objectField: built,
					fieldBase:   fieldBase,
				})

			case *j5schema.OneofField:
				preBuilt, err := newOneof(wrapper.Schema(), msg)
				if err != nil {
					return nil, patherr.Wrap(err, schema.JSONName)
				}
				built := newOneofField(wrapper, msg.virtualField(preBuilt.propSet))
				built._oneof = preBuilt
				out = append(out, &oneofProperty{
					oneofField: built,
					fieldBase:  fieldBase,
				})

			default:
				return nil, fmt.Errorf("unsupported schema type %T for nested json", wrapper)
			}

			continue
		}

		var walkFieldNumber protoreflect.FieldNumber
		walkPath := schema.ProtoField[:]
		walkMessage := msg
		for len(walkPath) > 1 {
			walkFieldNumber, walkPath = walkPath[0], walkPath[1:]
			walkMessage, err = walkMessage.childByNumber(walkFieldNumber)
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

		switch ft := field.(type) {
		case *mutableArrayField:
			return &mutableArrayProperty{
				mutableArrayField: ft,
				fieldBase:         fieldBase,
			}, nil
		case *leafArrayField:
			return &leafArrayProperty{
				leafArrayField: ft,
				fieldBase:      fieldBase,
			}, nil
		default:
			return nil, fmt.Errorf("unsupported array field type %T", field)
		}

	case *j5schema.MapField:
		field, err := newMapField(st, value)
		if err != nil {
			return nil, err
		}
		switch ft := field.(type) {
		case *mutableMapField:
			return &mutableMapProperty{
				mutableMapField: ft,
				fieldBase:       fieldBase,
			}, nil
		case *leafMapField:
			return &leafMapProperty{
				leafMapField: ft,
				fieldBase:    fieldBase,
			}, nil
		default:
			return nil, fmt.Errorf("unsupported map field type %T", field)
		}

	}

	ff, err := newFieldFactory(schema.Schema, value.field)
	if err != nil {
		return nil, err
	}

	field := ff.buildField(value).asProperty(fieldBase)
	return field, nil

}
