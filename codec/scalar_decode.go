package codec

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/pentops/j5/j5types/date_j5t"
	"github.com/pentops/j5/j5types/decimal_j5t"
	"github.com/pentops/j5/schema/j5reflect"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (dec *decoder) decodeScalarField(schema *j5reflect.ScalarSchema) (protoreflect.Value, error) {

	if schema.WellKnownTypeName != "" {
		switch schema.WellKnownTypeName {

		//case WKTAny:
		//	return unmarshalAny

		case WKTTimestamp:
			return dec.unmarshalTimestamp()

		case JTDate:
			return dec.unmarshalDate()

			//case WKTDuration:
			//	return unmarshalDuration

		case WKTBool:
			return dec.decodeWrapper(&wrappers.BoolValue{})

		case WKTInt32:
			return dec.decodeWrapper(&wrappers.Int32Value{})

		case WKTInt64:
			return dec.decodeWrapper(&wrappers.Int64Value{})

		case WKTUInt32:
			return dec.decodeWrapper(&wrappers.UInt32Value{})

		case WKTUInt64:
			return dec.decodeWrapper(&wrappers.UInt64Value{})

		case WKTFloat:
			return dec.decodeWrapper(&wrappers.FloatValue{})

		case WKTDouble:
			return dec.decodeWrapper(&wrappers.DoubleValue{})

		case WKTString:
			return dec.decodeWrapper(&wrappers.StringValue{})

		case WKTBytes:
			return dec.decodeWrapper(&wrappers.BytesValue{})

		case WKTEmpty:
			return protoreflect.Value{}, dec.unmarshalEmptyObject()

		case JTDecimal:
			return dec.decodeWrapper(&decimal_j5t.Decimal{})

		default:
			return protoreflect.Value{}, fmt.Errorf("unsupported well known type %q", schema.WellKnownTypeName)
		}

	}
	return dec.simpleScalarValue(schema.Kind)
}

func (dec *decoder) unmarshalTimestamp() (protoreflect.Value, error) {
	stringVal, err := dec.stringToken()
	if err != nil {
		return protoreflect.Value{}, err
	}

	t, err := time.Parse(time.RFC3339Nano, stringVal)
	if err != nil {
		return protoreflect.Value{}, err
	}

	msg := timestamppb.New(t)
	return protoreflect.ValueOfMessage(msg.ProtoReflect()), nil
}

func (dec *decoder) unmarshalDate() (protoreflect.Value, error) {

	stringVal, err := dec.stringToken()
	if err != nil {
		return protoreflect.Value{}, err
	}

	stringParts := strings.Split(stringVal, "-")
	if len(stringParts) != 3 {
		return protoreflect.Value{}, fmt.Errorf("expected date as yyyy-mm-dd, got %q", stringVal)
	}

	parts := make([]int32, 3)

	for idx, val := range stringParts {
		intVal, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return protoreflect.Value{}, fmt.Errorf("expected date as yyyy-mm-dd, got %q", stringVal)
		}
		parts[idx] = int32(intVal)

	}
	msg := &date_j5t.Date{
		Year:  parts[0],
		Month: parts[1],
		Day:   parts[2],
	}
	return protoreflect.ValueOfMessage(msg.ProtoReflect()), nil
}

func (dec *decoder) decodeWrapper(msg proto.Message) (protoreflect.Value, error) {
	m := msg.ProtoReflect()
	fd := m.Descriptor().Fields().ByName("value")
	val, err := dec.simpleScalarValue(fd.Kind())
	if err != nil {
		return protoreflect.Value{}, err
	}
	m.Set(fd, val)
	return protoreflect.ValueOfMessage(m), nil
}

func (dec *decoder) simpleScalarValue(kind protoreflect.Kind) (protoreflect.Value, error) {

	tok, err := dec.Token()
	if err != nil {
		return protoreflect.Value{}, err
	}

	switch kind {
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
		return protoreflect.Value{}, fmt.Errorf("unsupported scalar kind %v", kind)
	}
}
