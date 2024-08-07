package codec

import (
	"fmt"

	"github.com/pentops/j5/internal/j5reflect"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type fieldSpec struct {
	property *j5reflect.ObjectProperty

	value    protoreflect.Value
	children []fieldSpec
}

func collectProperties(properties []*j5reflect.ObjectProperty, msg protoreflect.Message) ([]fieldSpec, error) {
	var writeFields []fieldSpec

properties:
	for _, property := range properties {
		if len(property.ProtoField) == 0 {
			var childProperties []*j5reflect.ObjectProperty

			switch wrapper := property.Schema.(type) {
			case *j5reflect.ObjectField:
				childProperties = wrapper.Schema().Properties
			case *j5reflect.OneofField:
				childProperties = wrapper.Schema().Properties
			default:
				return nil, fmt.Errorf("unsupported schema type %T for nested json", wrapper)
			}
			children, err := collectProperties(childProperties, msg)
			if err != nil {
				return nil, err
			}
			if len(children) > 0 {
				writeFields = append(writeFields, fieldSpec{
					property: property,
					children: children,
				})
			}
			continue
		}

		var walkFieldNumber protoreflect.FieldNumber
		var walkField protoreflect.FieldDescriptor
		walkPath := property.ProtoField[:]
		walkMessage := msg
		for len(walkPath) > 1 {
			walkFieldNumber, walkPath = walkPath[0], walkPath[1:]
			walkField = walkMessage.Descriptor().Fields().ByNumber(walkFieldNumber)

			if !walkMessage.Has(walkField) {
				continue properties
			}
			if walkField.Kind() != protoreflect.MessageKind {
				return nil, fmt.Errorf("field %s is not a message but has nested types", walkField.FullName())
			}
			walkMessage = walkMessage.Get(walkField).Message()
		}
		walkFieldNumber = walkPath[0]
		walkField = walkMessage.Descriptor().Fields().ByNumber(walkFieldNumber)
		if !walkMessage.Has(walkField) {
			continue properties
		}
		value := walkMessage.Get(walkField)

		writeFields = append(writeFields, fieldSpec{
			property: property,
			value:    value,
		})
	}

	return writeFields, nil
}

func (enc *encoder) encodeObjectBody(fields []fieldSpec) error {

	enc.openObject()
	for idx, spec := range fields {
		if idx > 0 {
			enc.fieldSep()
		}
		if err := enc.fieldLabel(spec.property.JSONName); err != nil {
			return err
		}
		if len(spec.children) > 0 {
			switch subSchema := spec.property.Schema.(type) {
			case *j5reflect.ObjectField:
				if err := enc.encodeObjectBody(spec.children); err != nil {
					return err
				}
			case *j5reflect.OneofField:
				if err := enc.encodeOneofBody(spec.children); err != nil {
					return err
				}
			default:
				return fmt.Errorf("invalid schema type for children: %T", subSchema)
			}
		} else {
			if err := enc.encodeValue(spec.property.Schema, spec.value); err != nil {
				return err
			}
		}

	}
	enc.closeObject()
	return nil
}

func (enc *encoder) encodeOneofBody(properties []fieldSpec) error {

	if len(properties) == 0 {
		return nil
	}
	if len(properties) > 1 {
		return fmt.Errorf("multiple values set for oneof")
	}
	spec := properties[0]

	enc.openObject()

	var err error

	err = enc.fieldLabel("!type")
	if err != nil {
		return err
	}

	err = enc.addString(spec.property.JSONName)
	if err != nil {
		return err
	}

	enc.fieldSep()

	err = enc.fieldLabel(spec.property.JSONName)
	if err != nil {
		return err
	}

	if len(spec.children) > 0 {
		if err := enc.encodeObjectBody(spec.children); err != nil {
			return err
		}
	} else {
		if err := enc.encodeValue(spec.property.Schema, spec.value); err != nil {
			return err
		}
	}

	enc.closeObject()
	return nil
}

func (enc *encoder) encodeObject(schema *j5reflect.ObjectSchema, msg protoreflect.Message) error {

	fields, err := collectProperties(schema.ClientProperties(), msg)
	if err != nil {
		return err
	}

	if err := enc.encodeObjectBody(fields); err != nil {
		return err
	}

	return nil
}

func (enc *encoder) encodeOneof(schema *j5reflect.OneofSchema, msg protoreflect.Message) error {

	properties, err := collectProperties(schema.Properties, msg)
	if err != nil {
		return err
	}

	if err := enc.encodeOneofBody(properties); err != nil {
		return err
	}

	return nil
}

func (enc *encoder) encodeValue(schema j5reflect.FieldSchema, value protoreflect.Value) error {

	switch schema := schema.(type) {
	case *j5reflect.ObjectField:
		return enc.encodeObject(schema.Schema(), value.Message())

	case *j5reflect.OneofField:
		return enc.encodeOneof(schema.Schema(), value.Message())

	case *j5reflect.EnumField:
		return enc.encodeEnum(schema.Schema(), value.Enum())

	case *j5reflect.ArrayField:
		return enc.encodeArray(schema, value.List())

	case *j5reflect.MapField:
		return enc.encodeMap(schema, value.Map())

	case *j5reflect.ScalarSchema:
		return enc.encodeScalarField(schema, value)

	default:
		return fmt.Errorf("unsupported schema %T", schema)
	}
}

func (enc *encoder) encodeMap(schema *j5reflect.MapField, value protoreflect.Map) error {
	enc.openObject()
	first := true
	var outerError error

	value.Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
		if !first {
			enc.fieldSep()
		}
		first = false

		keyString := key.Value().String()
		outerError = enc.fieldLabel(keyString)
		if outerError != nil {
			return false
		}
		outerError = enc.encodeValue(schema.Schema, val)
		return outerError == nil
	})
	if outerError != nil {
		return outerError
	}
	enc.closeObject()
	return nil
}

func (enc *encoder) encodeArray(schema *j5reflect.ArrayField, list protoreflect.List) error {
	enc.openArray()
	first := true
	for i := 0; i < list.Len(); i++ {
		if !first {
			enc.fieldSep()
		}
		first = false
		if err := enc.encodeValue(schema.Schema, list.Get(i)); err != nil {
			return err
		}
	}

	enc.closeArray()
	return nil
}

func (enc *encoder) encodeEnum(schema *j5reflect.EnumSchema, enumVal protoreflect.EnumNumber) error {
	value := int32(enumVal)

	for _, val := range schema.Options {
		if val.Number == value {
			return enc.addString(val.Name)
		}
	}

	return fmt.Errorf("enum value %d not found", value)
}
