package codec

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func decodeScalarField(dec Decoder, field protoreflect.FieldDescriptor) (protoreflect.Value, error) {
	tok, err := dec.Token()
	if err != nil {
		return protoreflect.Value{}, err
	}

	switch field.Kind() {
	case protoreflect.StringKind:
		str, ok := tok.(string)
		if !ok {
			return protoreflect.Value{}, fmt.Errorf("expected string but got %v", tok)
		}
		return protoreflect.ValueOfString(str), nil

	case protoreflect.BoolKind:
		b, ok := tok.(bool)
		if !ok {
			return protoreflect.Value{}, fmt.Errorf("expected bool but got %v", tok)
		}
		return protoreflect.ValueOfBool(b), nil

	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		i, ok := tok.(json.Number)
		if !ok {
			return protoreflect.Value{}, fmt.Errorf("expected int32 but got %v", tok)
		}
		intVal, err := i.Int64()
		if err != nil {
			return protoreflect.Value{}, err
		}

		return protoreflect.ValueOfInt32(int32(intVal)), nil

	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		i, ok := tok.(json.Number)
		if !ok {
			return protoreflect.Value{}, fmt.Errorf("expected int64 but got %v", tok)
		}
		intVal, err := i.Int64()
		if err != nil {
			return protoreflect.Value{}, err
		}

		return protoreflect.ValueOfInt64(int64(intVal)), nil

	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		i, ok := tok.(json.Number)
		if !ok {
			return protoreflect.Value{}, fmt.Errorf("expected uint32 but got %v", tok)
		}
		intVal, err := i.Int64()
		if err != nil {
			return protoreflect.Value{}, err
		}

		return protoreflect.ValueOfUint32(uint32(intVal)), nil

	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		i, ok := tok.(json.Number)
		if !ok {
			return protoreflect.Value{}, fmt.Errorf("expected uint64 but got %v", tok)
		}
		intVal, err := i.Int64()
		if err != nil {
			return protoreflect.Value{}, err
		}

		return protoreflect.ValueOfUint64(uint64(intVal)), nil

	case protoreflect.FloatKind:
		f, ok := tok.(json.Number)
		if !ok {
			return protoreflect.Value{}, fmt.Errorf("expected float but got %v", tok)
		}
		floatVal, err := f.Float64()
		if err != nil {
			return protoreflect.Value{}, err
		}

		return protoreflect.ValueOfFloat32(float32(floatVal)), nil

	case protoreflect.DoubleKind:
		f, ok := tok.(json.Number)
		if !ok {
			return protoreflect.Value{}, fmt.Errorf("expected double but got %v", tok)
		}
		floatVal, err := f.Float64()
		if err != nil {
			return protoreflect.Value{}, err
		}

		return protoreflect.ValueOfFloat64(floatVal), nil

	case protoreflect.EnumKind:
		stringVal, ok := tok.(string)
		if !ok {
			return protoreflect.Value{}, unexpectedTokenError(tok, "string")
		}
		enumValue, err := dec.Options().ShortEnums.Decode(field.Enum(), stringVal)
		if err != nil {
			return protoreflect.Value{}, err
		}

		return protoreflect.ValueOfEnum(enumValue), nil

	case protoreflect.BytesKind:
		stringVal, ok := tok.(string)
		if !ok {
			return protoreflect.Value{}, unexpectedTokenError(tok, "base64 string")
		}

		// Copied from protojson
		enc := base64.StdEncoding
		if strings.ContainsAny(stringVal, "-_") {
			enc = base64.URLEncoding
		}
		if len(stringVal)%4 != 0 {
			enc = enc.WithPadding(base64.NoPadding)
		}
		bytesVal, err := enc.DecodeString(stringVal)
		if err != nil {
			return protoreflect.Value{}, err
		}

		return protoreflect.ValueOfBytes(bytesVal), nil
		// End copy

	default:
		return protoreflect.Value{}, fmt.Errorf("unsupported scalar kind %v", field.Kind())
	}
}
