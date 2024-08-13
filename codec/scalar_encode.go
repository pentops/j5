package codec

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/pentops/j5/internal/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	wktAny       = "google.protobuf.Any"
	wktTimestamp = "google.protobuf.Timestamp"
	wktDuration  = "google.protobuf.Duration"

	wktBool   = "google.protobuf.BoolValue"
	wktInt32  = "google.protobuf.Int32Value"
	wktInt64  = "google.protobuf.Int64Value"
	wktUInt32 = "google.protobuf.UInt32Value"
	wktUInt64 = "google.protobuf.UInt64Value"
	wktFloat  = "google.protobuf.FloatValue"
	wktDouble = "google.protobuf.DoubleValue"
	wktString = "google.protobuf.StringValue"
	wktBytes  = "google.protobuf.BytesValue"

	wktEmpty = "google.protobuf.Empty"

	jtDate    = protoreflect.FullName("j5.types.date.v1.Date")
	jtDecimal = protoreflect.FullName("j5.types.decimal.v1.Decimal")
)

func (e *encoder) encodeScalarField(field *j5schema.ScalarSchema, val protoreflect.Value) error {

	if field.WellKnownTypeName != "" {
		switch field.WellKnownTypeName {

		case wktTimestamp:
			return e.marshalTimestamp(val.Message())

		//case WKTAny:
		//	return marshalAny

		//case WKTDuration:
		//	return marshalDuration

		case
			wktBool,
			wktInt32,
			wktInt64,
			wktUInt32,
			wktUInt64,
			wktFloat,
			wktDouble,
			wktString,
			wktBytes:
			return e.marshalWrapperType(field, val.Message())
		case wktEmpty:
			e.marshalEmpty()
			return nil

		case jtDate:
			return e.marshalDate(val.Message())
		case jtDecimal:
			return e.marshalWrapperType(field, val.Message())
		}
		return nil
	}

	return e.simpleScalarValue(field.Kind, val)
}

func (enc *encoder) simpleScalarValue(kind protoreflect.Kind, value protoreflect.Value) error {
	switch kind {

	case protoreflect.StringKind:
		return enc.addString(value.String())

	case protoreflect.BoolKind:
		enc.addBool(value.Bool())
		return nil

	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		enc.addInt(value.Int())
		return nil

	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		enc.addInt(value.Int())
		return nil

	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		enc.addUint(value.Uint())
		return nil

	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		enc.addUint(value.Uint())
		return nil

	case protoreflect.FloatKind:
		enc.addFloat(value.Float(), 32)
		return nil

	case protoreflect.DoubleKind:
		enc.addFloat(value.Float(), 64)
		return nil

	case protoreflect.BytesKind:
		byteVal := value.Bytes()
		encoded := base64.StdEncoding.EncodeToString(byteVal)
		return enc.addString(encoded)

	default:
		return fmt.Errorf("unsupported scalar kind %v", kind)
	}
}

func (e *encoder) marshalTimestamp(msg protoreflect.Message) error {
	seconds := msg.Get(msg.Descriptor().Fields().ByName("seconds")).Int()
	nanos := msg.Get(msg.Descriptor().Fields().ByName("nanos")).Int()
	t := time.Unix(seconds, nanos).In(time.UTC)
	return e.addString(t.Format(time.RFC3339Nano))
}

func (e *encoder) marshalEmpty() {
	e.add([]byte("{}"))
}

func (e *encoder) marshalWrapperType(field *j5schema.ScalarSchema, msg protoreflect.Message) error {
	fd := msg.Descriptor().Fields().ByName("value")
	val := msg.Get(fd)
	return e.simpleScalarValue(field.Kind, val)
}

func (e *encoder) marshalDate(msg protoreflect.Message) error {
	intParts := make([]int32, 3)
	for idx, key := range []protoreflect.Name{"year", "month", "day"} {
		field := msg.Descriptor().Fields().ByName(key)
		if field == nil {
			return fmt.Errorf("field %s not found", key)
		}

		val := msg.Get(field).Int()
		intParts[idx] = int32(val)
	}

	stringVal := fmt.Sprintf("%04d-%02d-%02d", intParts[0], intParts[1], intParts[2])
	return e.addString(stringVal)
}
