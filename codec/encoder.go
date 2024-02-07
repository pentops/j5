package codec

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/pentops/jsonapi/gen/j5/ext/v1/ext_j5pb"
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

type Encoder interface {
	Null() error
	String(string) error
	Bool(bool) error
	Number(json.Number) error
	JSONMarshal(interface{}) error

	Scalar(protoreflect.Kind, protoreflect.Value) error

	BeginObject() error
	Key(string) error
	BeginArray() error
	End() error
}

type nestElement int

var (
	nestObject nestElement = 1
	nestArray  nestElement = 2
	nestKey    nestElement = 3
)

type safeEncoder struct {
	encoder *encoder
	nesting []nestElement
	current nestElement
}

func newSafeEncoder(enc *encoder) *safeEncoder {
	return &safeEncoder{encoder: enc, current: nestKey}
}

func (enc *safeEncoder) receiveValue() error {
	if enc.current != nestKey && enc.current != nestArray {
		return fmt.Errorf("not expecting a value")
	}

	enc.popNest()
	return nil
}

func (enc *safeEncoder) popNest() {
	if len(enc.nesting) == 0 {
		enc.current = 0
		return

	}
	enc.current = enc.nesting[0]
	enc.nesting = enc.nesting[1:]
}

func (enc *safeEncoder) nest(nest nestElement) {
	enc.nesting = append(enc.nesting, enc.current)
	enc.current = nest
}

func (enc *safeEncoder) Null() error {
	if err := enc.receiveValue(); err != nil {
		return err
	}
	enc.encoder.add([]byte("null"))
	return nil
}

func (enc *safeEncoder) JSONMarshal(object interface{}) error {
	if err := enc.receiveValue(); err != nil {
		return err
	}
	return enc.encoder.addJSON(object)
}

func (enc *safeEncoder) Scalar(kind protoreflect.Kind, v protoreflect.Value) error {
	if err := enc.receiveValue(); err != nil {
		return err
	}
	val, err := EncodeScalar(kind, v)
	if err != nil {
		return err
	}
	enc.encoder.add(val)
	return nil
}

func (enc *safeEncoder) scalar(v interface{}) error {
	if err := enc.receiveValue(); err != nil {
		return err
	}
	if err := enc.encoder.addJSON(v); err != nil {
		return err
	}
	enc.popNest()
	return nil
}

func (enc *safeEncoder) String(s string) error {
	return enc.scalar(s)
}

func (enc *safeEncoder) Bool(b bool) error {
	return enc.scalar(b)
}

func (enc *safeEncoder) Number(n json.Number) error {
	return enc.scalar(n)
}

func (enc *safeEncoder) Key(k string) error {
	if enc.current != nestObject {
		return fmt.Errorf("cannot write key outside of object")
	}
	enc.nest(nestKey)
	enc.encoder.fieldLabel(k)
	return nil
}

func (enc *safeEncoder) End() error {
	if enc.current == nestObject {
		enc.encoder.closeObject()
	} else if enc.current == nestArray {
		enc.encoder.closeArray()
	} else {
		return fmt.Errorf("cannot end outside of object or array")
	}
	enc.popNest()
	return nil
}

func (enc *safeEncoder) BeginObject() error {
	if err := enc.receiveValue(); err != nil {
		return err
	}
	enc.encoder.openObject()
	enc.nest(nestObject)
	return nil
}

func (enc *safeEncoder) BeginArray() error {
	if err := enc.receiveValue(); err != nil {
		return err
	}
	enc.encoder.openArray()
	enc.nest(nestArray)
	return nil
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

	customDecoder, ok := enc.Options.CustomEntities[msg.Descriptor().FullName()]
	if ok {
		return customDecoder.Marshal(newSafeEncoder(enc), msg)
	}

	wktEncoder := wellKnownType(msg.Descriptor().FullName())
	if wktEncoder != nil {
		return wktEncoder.Marshal(newSafeEncoder(enc), msg)
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

		enc.fieldLabel(field.JSONName())

		if err := enc.encodeField(field, value); err != nil {
			return err
		}

	}

	enc.closeObject()
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
	enc.openObject()
	first := true
	var outerError error
	keyDesc := field.MapKey()
	valDesc := field.MapValue()

	value.Map().Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
		fmt.Printf("enc key %v\n", key.Interface())
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

	kind := field.Kind()
	switch kind {
	case protoreflect.MessageKind:
		return enc.encodeMessage(value.Message())

	case protoreflect.EnumKind:
		stringVal, err := enc.ShortEnums.Encode(field.Enum(), value.Enum())
		if err != nil {
			return err
		}
		return enc.addJSON(stringVal)

	default:
		encoded, err := EncodeScalar(kind, value)
		if err != nil {
			return err
		}
		enc.add(encoded)
	}
	return nil
}

func EncodeScalar(kind protoreflect.Kind, value protoreflect.Value) ([]byte, error) {
	switch kind {
	case protoreflect.StringKind:
		return json.Marshal(value.String())

	case protoreflect.BoolKind:
		return json.Marshal(value.Bool())

	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return json.Marshal(int32(value.Int()))

	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return json.Marshal(int64(value.Int()))

	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return json.Marshal(uint32(value.Uint()))

	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return json.Marshal(uint64(value.Uint()))

	case protoreflect.FloatKind:
		return json.Marshal(float32(value.Float()))

	case protoreflect.DoubleKind:
		return json.Marshal(float64(value.Float()))

	case protoreflect.BytesKind:
		byteVal := value.Bytes()
		encoded := base64.StdEncoding.EncodeToString(byteVal)
		return json.Marshal(encoded)

	default:
		return nil, fmt.Errorf("unsupported scalar kind %v", kind)

	}
}
