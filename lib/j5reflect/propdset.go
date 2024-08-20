package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/internal/j5schema"
	"github.com/pentops/j5/internal/patherr"
	"google.golang.org/protobuf/reflect/protoreflect"
)

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

func (fs *propSet) Name() string {
	return fs.fullName
}

func (fs *propSet) MaybeGetProperty(name string) Property {
	prop := fs.asMap[name]
	return prop
}

func (fs *propSet) GetProperty(name string) (Property, error) {
	prop := fs.MaybeGetProperty(name)
	if prop == nil {
		fmt.Printf("no property %s in %s. Has:\n", name, fs.fullName)
		for _, p := range fs.asSlice {
			fmt.Printf("  %s\n", p.JSONName())
		}
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

func (fs *propSet) HasAnyValue() bool {
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

var cb func(name string, params ...interface{})

func collectProperties(properties []*j5schema.ObjectProperty, msg *protoMessageWrapper) ([]Property, error) {
	out := make([]Property, 0)
	var err error

	for _, schema := range properties {
		if cb != nil {
			cb("building property: %s, path %v", schema.JSONName, schema.ProtoField)
		}
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

		if cb != nil {
			cb("is normal %q, path %v", schema.JSONName, schema.ProtoField)
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

		if cb != nil {
			cb("final field build, schema %s", schema.ToJ5Proto())
			obj, ok := built.(*objectProperty)
			if ok {
				cb("schema.Ref.FullName: %v", obj.field.schema.Ref.FullName())
				cb("schema.Ref.To.Full : %v", obj.field.schema.Ref.To.FullName())
				cb("schema.Schema().Ful: %v", obj.field.schema.Schema().FullName())
			}
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
