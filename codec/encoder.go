package codec

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func Encode(opts Options, msg protoreflect.Message) ([]byte, error) {
	enc := &encoder{
		b:       &bytes.Buffer{},
		Options: opts,
	}
	if err := enc.encodeMessage(msg); err != nil {
		return nil, err

	}
	return enc.b.Bytes(), nil
}

type encoder struct {
	b *bytes.Buffer
	Options
}

func (enc *encoder) add(b []byte) {
	enc.b.Write(b)
}

// addJSON is a shortcut for actually writing the marshal code for scalars
func (enc *encoder) addJSON(v interface{}) error {
	jv, err := json.Marshal(v)
	if err != nil {
		return err
	}
	enc.add(jv)
	return nil
}

func (enc *encoder) openObject() {
	enc.add([]byte("{"))
}

func (enc *encoder) closeObject() {
	enc.add([]byte("}"))
}

func (enc *encoder) openArray() {
	enc.add([]byte("["))
}

func (enc *encoder) closeArray() {
	enc.add([]byte("]"))
}

func (enc *encoder) fieldSep() {
	enc.add([]byte(","))
}

func (enc *encoder) fieldLabel(label string) {
	enc.add([]byte(fmt.Sprintf(`"%s":`, label)))
}

func (enc *encoder) encodeMessage(msg protoreflect.Message) error {

	wktEncoder := wellKnownTypeMarshaler(msg.Descriptor().FullName())
	if wktEncoder != nil {
		return wktEncoder(enc, msg)
	}

	isOneofWrapper := false
	msgOptions := msg.Descriptor().Options()
	ext := proto.GetExtension(msgOptions, ext_j5pb.E_Message).(*ext_j5pb.MessageOptions)
	if ext != nil {
		isOneofWrapper = ext.IsOneofWrapper
	}

	enc.openObject()

	first := true

	fields := msg.Descriptor().Fields()
	for idx := 0; idx < fields.Len(); idx++ {
		field := fields.Get(idx)
		if !msg.Has(field) {
			continue
		}

		if !first {
			enc.fieldSep()
		}
		first = false

		value := msg.Get(field)

		if !isOneofWrapper && enc.WrapOneof {
			if oneof := field.ContainingOneof(); oneof != nil && !oneof.IsSynthetic() {
				enc.fieldLabel(protoNameToJSON(string(oneof.Name())))
				enc.openObject()
				enc.fieldLabel(field.JSONName())
				if err := enc.encodeValue(field, value); err != nil {
					return err
				}
				enc.closeObject()
				continue
			}
		}

		if err := enc.encodeField(field, value); err != nil {
			return err
		}

	}

	enc.closeObject()
	return nil
}

func (enc *encoder) encodeField(field protoreflect.FieldDescriptor, value protoreflect.Value) error {
	fieldOptions := proto.GetExtension(field.Options(), ext_j5pb.E_Field).(*ext_j5pb.FieldOptions)
	if fieldOptions != nil {
		switch option := fieldOptions.Type.(type) {
		case *ext_j5pb.FieldOptions_Message:
			if field.Kind() != protoreflect.MessageKind {
				return fmt.Errorf("field %s is not a message but has a message annotation", field.FullName())
			}

			msgVal := value.Message()

			if option.Message.Flatten {

				subFields := field.Message().Fields()

				for idx := 0; idx < subFields.Len(); idx++ {
					subField := subFields.Get(idx)

					if !msgVal.Has(subField) {
						continue
					}
					if err := enc.encodeField(subField, msgVal.Get(subField)); err != nil {
						return fmt.Errorf("field %s: %w", subField.Name(), err)
					}
				}
				return nil
			}
		}
	}

	enc.fieldLabel(field.JSONName())

	if field.IsMap() {
		return enc.encodeMapField(field, value)
	}
	if field.IsList() {
		return enc.encodeListField(field, value)
	}

	return enc.encodeValue(field, value)
}

func (enc *encoder) encodeMapField(field protoreflect.FieldDescriptor, value protoreflect.Value) error {
	enc.openObject()
	first := true
	var outerError error
	keyDesc := field.MapKey()
	valDesc := field.MapValue()

	value.Map().Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
		if !first {
			enc.fieldSep()
		}
		first = false
		if err := enc.encodeValue(keyDesc, key.Value()); err != nil {
			outerError = err
			return false
		}
		enc.add([]byte(":"))
		if err := enc.encodeValue(valDesc, val); err != nil {
			outerError = err
			return false
		}
		return true
	})
	if outerError != nil {
		return outerError
	}
	enc.closeObject()
	return nil
}

func (enc *encoder) encodeListField(field protoreflect.FieldDescriptor, value protoreflect.Value) error {
	enc.openArray()
	first := true
	var outerError error
	list := value.List()
	for i := 0; i < list.Len(); i++ {
		if !first {
			enc.fieldSep()
		}
		first = false
		if err := enc.encodeValue(field, value.List().Get(i)); err != nil {
			return err
		}
	}

	if outerError != nil {
		return outerError
	}
	enc.closeArray()
	return nil
}

func (enc *encoder) encodeValue(field protoreflect.FieldDescriptor, value protoreflect.Value) error {

	switch field.Kind() {
	case protoreflect.MessageKind:
		return enc.encodeMessage(value.Message())

	case protoreflect.StringKind:
		return enc.addJSON(value.String())

	case protoreflect.BoolKind:
		return enc.addJSON(value.Bool())

	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return enc.addJSON(int32(value.Int()))

	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return enc.addJSON(int64(value.Int()))

	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return enc.addJSON(uint32(value.Uint()))

	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return enc.addJSON(uint64(value.Uint()))

	case protoreflect.FloatKind:
		return enc.addJSON(float32(value.Float()))

	case protoreflect.DoubleKind:
		return enc.addJSON(float64(value.Float()))

	case protoreflect.EnumKind:
		stringVal, err := enc.ShortEnums.Encode(field.Enum(), value.Enum())
		if err != nil {
			return err
		}
		return enc.addJSON(stringVal)

	case protoreflect.BytesKind:
		byteVal := value.Bytes()
		encoded := base64.StdEncoding.EncodeToString(byteVal)
		return enc.addJSON(encoded)

	default:
		return fmt.Errorf("unsupported kind %v", field.Kind())

	}

}
