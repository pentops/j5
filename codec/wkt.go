package codec

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func unexpectedTokenError(got, expected interface{}) error {
	return fmt.Errorf("unexpected token %v, expected %v", got, expected)
}

var wktCustomEntities = map[protoreflect.FullName]CustomEntity{
	protoreflect.FullName("j5.types.date.v1.Date"):       &dateEntity{},
	protoreflect.FullName("j5.types.decimal.v1.Decimal"): &wrapperEntity{},
	protoreflect.FullName("google.protobuf.Timestamp"):   &timestampEntity{},
	protoreflect.FullName("google.protobuf.BoolValue"):   &wrapperEntity{},
	protoreflect.FullName("google.protobuf.Int32Value"):  &wrapperEntity{},
	protoreflect.FullName("google.protobuf.Int64Value"):  &wrapperEntity{},
	protoreflect.FullName("google.protobuf.UInt32Value"): &wrapperEntity{},
	protoreflect.FullName("google.protobuf.UInt64Value"): &wrapperEntity{},
	protoreflect.FullName("google.protobuf.FloatValue"):  &wrapperEntity{},
	protoreflect.FullName("google.protobuf.DoubleValue"): &wrapperEntity{},
	protoreflect.FullName("google.protobuf.StringValue"): &wrapperEntity{},
	protoreflect.FullName("google.protobuf.BytesValue"):  &wrapperEntity{},
	protoreflect.FullName("google.protobuf.Empty"):       &emptyEntity{},
}

type timestampEntity struct{}

func (timestampEntity) Unmarshal(dec Decoder, msg protoreflect.Message) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}

	stringVal, ok := tok.(string)
	if !ok {
		return fmt.Errorf("expected string but got %v", tok)
	}

	t, err := time.Parse(time.RFC3339Nano, stringVal)
	if err != nil {
		return err
	}

	msg.Set(msg.Descriptor().Fields().ByName("seconds"), protoreflect.ValueOf(t.Unix()))
	msg.Set(msg.Descriptor().Fields().ByName("nanos"), protoreflect.ValueOf(int32(t.Nanosecond())))
	return nil
}

func (timestampEntity) Marshal(enc Encoder, msg protoreflect.Message) error {
	seconds := msg.Get(msg.Descriptor().Fields().ByName("seconds")).Int()
	nanos := msg.Get(msg.Descriptor().Fields().ByName("nanos")).Int()
	t := time.Unix(seconds, nanos).In(time.UTC)

	return enc.String(t.Format(time.RFC3339Nano))
}

type wrapperEntity struct{}

func (wrapperEntity) Unmarshal(dec Decoder, m protoreflect.Message) error {
	fd := m.Descriptor().Fields().ByName("value")
	val, err := decodeScalarField(dec, fd)
	if err != nil {
		return err
	}
	m.Set(fd, val)
	return nil
}

func (wrapperEntity) Marshal(e Encoder, msg protoreflect.Message) error {
	fd := msg.Descriptor().Fields().ByName("value")
	val := msg.Get(fd)
	return e.Scalar(fd.Kind(), val)
}

type emptyEntity struct{}

func (emptyEntity) Unmarshal(d Decoder, msg protoreflect.Message) error {
	tok, err := d.Token()
	if err != nil {
		return err
	}
	if tok != json.Delim('{') {
		return unexpectedTokenError(tok, "{")
	}
	tok, err = d.Token()
	if err != nil {
		return err
	}
	if tok != json.Delim('}') {
		return unexpectedTokenError(tok, "}")
	}
	return nil
}

func (emptyEntity) Marshal(e Encoder, msg protoreflect.Message) error {
	if err := e.BeginObject(); err != nil {
		return err
	}
	if err := e.End(); err != nil {
		return err
	}
	return nil
}

type dateEntity struct{}

func (dateEntity) Unmarshal(d Decoder, msg protoreflect.Message) error {
	tok, err := d.Token()
	if err != nil {
		return err
	}

	stringVal, ok := tok.(string)
	if !ok {
		return fmt.Errorf("expected date as a string but got %v", tok)
	}

	stringParts := strings.Split(stringVal, "-")
	if len(stringParts) != 3 {
		return fmt.Errorf("expected date as a string but got %v", tok)
	}

	for idx, key := range []protoreflect.Name{"year", "month", "day"} {
		field := msg.Descriptor().Fields().ByName(key)
		if field == nil {
			return fmt.Errorf("field %s not found", key)
		}

		intVal, err := strconv.ParseInt(stringParts[idx], 10, 32)
		if err != nil {
			return fmt.Errorf("expected date as yyyy-mm-dd, got %s", stringVal)
		}

		msg.Set(field, protoreflect.ValueOf(int32(intVal)))
	}

	return nil
}

func (dateEntity) Marshal(e Encoder, msg protoreflect.Message) error {
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
	return e.String(stringVal)
}
