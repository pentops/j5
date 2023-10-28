package codec

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func Encode(opts Options, msg protoreflect.Message) ([]byte, error) {
	if opts.UnspecifiedEnumSuffix == "" {
		opts.UnspecifiedEnumSuffix = "_UNSPECIFIED"
	}
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

func (enc *encoder) encodeMessage(msg protoreflect.Message) error {

	wktEncoder := wellKnownTypeMarshaler(msg.Descriptor().FullName())
	if wktEncoder != nil {
		return wktEncoder(enc, msg)
	}

	enc.add([]byte("{"))

	first := true

	fields := msg.Descriptor().Fields()
	for idx := 0; idx < fields.Len(); idx++ {
		field := fields.Get(idx)
		if !msg.Has(field) {
			continue
		}

		if !first {
			enc.add([]byte(","))
		}
		first = false

		value := msg.Get(field)

		enc.add([]byte(fmt.Sprintf(`"%s":`, field.JSONName())))

		if err := enc.encodeField(field, value); err != nil {
			return err
		}

	}

	enc.add([]byte("}"))
	return nil
}

func (enc *encoder) encodeField(field protoreflect.FieldDescriptor, value protoreflect.Value) error {
	if field.IsMap() {
		return enc.encodeMapField(field, value)
	}
	if field.IsList() {
		return enc.encodeListField(field, value)
	}
	return enc.encodeValue(field, value)
}

func (enc *encoder) encodeMapField(field protoreflect.FieldDescriptor, value protoreflect.Value) error {
	enc.add([]byte("{"))
	first := true
	var outerError error
	keyDesc := field.MapKey()
	valDesc := field.MapValue()

	value.Map().Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
		fmt.Printf("enc key %v\n", key.Interface())
		if !first {
			enc.add([]byte(","))
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
	enc.add([]byte("}"))
	return nil
}

func (enc *encoder) encodeListField(field protoreflect.FieldDescriptor, value protoreflect.Value) error {
	enc.add([]byte("["))
	first := true
	var outerError error
	list := value.List()
	for i := 0; i < list.Len(); i++ {
		if !first {
			enc.add([]byte(","))
		}
		first = false
		if err := enc.encodeValue(field, value.List().Get(i)); err != nil {
			return err
		}
	}

	if outerError != nil {
		return outerError
	}
	enc.add([]byte("]"))
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
		return enc.addJSON(uint32(value.Int()))

	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return enc.addJSON(uint64(value.Int()))

	case protoreflect.FloatKind:
		return enc.addJSON(float32(value.Float()))

	case protoreflect.DoubleKind:
		return enc.addJSON(float64(value.Float()))

	case protoreflect.EnumKind:
		enumVals := field.Enum().Values()
		unspecifiedField := enumVals.ByNumber(0)
		specifiedField := enumVals.ByNumber(value.Enum())
		returnName := string(specifiedField.Name())

		if enc.ShortEnums {
			if unspecifiedField != nil {
				unspecifiedName := string(unspecifiedField.Name())
				if strings.HasSuffix(unspecifiedName, enc.UnspecifiedEnumSuffix) {
					unspecifiedPrefix := strings.TrimSuffix(unspecifiedName, enc.UnspecifiedEnumSuffix)
					returnName = strings.TrimPrefix(returnName, unspecifiedPrefix+"_")
				}
			}
		}

		return enc.addJSON(returnName)

	case protoreflect.BytesKind:
		byteVal := value.Bytes()
		encoded := base64.StdEncoding.EncodeToString(byteVal)
		return enc.addJSON(encoded)

	default:
		return fmt.Errorf("unsupported kind %v", field.Kind())

	}

}
