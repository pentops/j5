package j5reflect

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/j5types/date_j5t"
	"github.com/pentops/j5/j5types/decimal_j5t"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AnyValue interface{}

type KeyValue string

func scalarReflectFromGo(schema *schema_j5pb.Field, value interface{}) (protoreflect.Value, error) {
	var pv protoreflect.Value
	switch st := schema.Type.(type) {
	case *schema_j5pb.Field_Any:
		pv = protoreflect.ValueOf(value)
		// and hope for the best?
		return pv, nil

	case *schema_j5pb.Field_Bool:
		switch val := value.(type) {
		case bool:
			return protoreflect.ValueOfBool(val), nil

		case *bool:
			if val == nil {
				return protoreflect.Value{}, nil
			}
			return protoreflect.ValueOfBool(*val), nil

		case nil:
			return protoreflect.Value{}, nil

		default:
			return pv, fmt.Errorf("expected bool, got %T", value)
		}

	case *schema_j5pb.Field_String_:
		switch val := value.(type) {
		case string:
			return protoreflect.ValueOfString(val), nil

		case *string:
			if val == nil {
				return protoreflect.Value{}, nil
			}
			return protoreflect.ValueOfString(*val), nil

		case nil:
			return protoreflect.Value{}, nil

		default:
			return pv, fmt.Errorf("expected string, got %T", value)
		}

	case *schema_j5pb.Field_Key:

		switch val := value.(type) {

		case string:
			return protoreflect.ValueOfString(string(val)), nil

		case *string:
			if val == nil {
				return protoreflect.Value{}, nil
			}

			return protoreflect.ValueOfString(string(*val)), nil

		case KeyValue:
			return protoreflect.ValueOfString(string(val)), nil

		case *KeyValue:
			if val == nil {
				return protoreflect.Value{}, nil
			}

			return protoreflect.ValueOfString(string(*val)), nil

		case nil:
			return protoreflect.Value{}, nil

		default:
			return pv, fmt.Errorf("expected KeyValue, got %T", value)
		}

	case *schema_j5pb.Field_Integer:
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return protoreflect.Value{}, nil
			}
			rv = rv.Elem()
		}

		switch st.Integer.Format {
		case schema_j5pb.IntegerField_FORMAT_INT32:
			if !rv.CanInt() {
				return pv, fmt.Errorf("expected int, got %T", value)
			}
			val := rv.Int()
			if val > math.MaxInt32 || val < math.MinInt32 {
				return pv, fmt.Errorf("int value %v is out of range for int32", val)
			}
			return protoreflect.ValueOfInt32(int32(val)), nil

		case schema_j5pb.IntegerField_FORMAT_INT64:
			if !rv.CanInt() {
				return pv, fmt.Errorf("expected int, got %T", value)
			}
			val := rv.Int()
			return protoreflect.ValueOfInt64(val), nil

		case schema_j5pb.IntegerField_FORMAT_UINT32:
			if !rv.CanUint() {
				return pv, fmt.Errorf("expected int, got %T", value)
			}
			val := rv.Uint()
			if val > math.MaxUint32 {
				return pv, fmt.Errorf("int value %v is out of range for uint32", val)
			}
			return protoreflect.ValueOfUint32(uint32(val)), nil

		case schema_j5pb.IntegerField_FORMAT_UINT64:
			if !rv.CanUint() {
				return pv, fmt.Errorf("expected int, got %T", value)
			}
			val := rv.Uint()
			return protoreflect.ValueOfUint64(val), nil

		default:
			return pv, fmt.Errorf("unsupported integer format %v", st.Integer.Format)
		}

	case *schema_j5pb.Field_Float:
		if number, ok := value.(json.Number); ok {
			var err error
			value, err = number.Float64()
			if err != nil {
				return pv, err
			}
		}
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return protoreflect.Value{}, nil
			}
			rv = rv.Elem()
		}

		if !rv.CanFloat() {
			return pv, fmt.Errorf("value can't float, got %T", value)
		}

		val := rv.Float()

		switch st.Float.Format {

		case schema_j5pb.FloatField_FORMAT_FLOAT32:
			if val > math.MaxFloat32 || val < -math.MaxFloat32 {
				return pv, fmt.Errorf("float64 value %v is out of range for float32", val)
			}

			return protoreflect.ValueOfFloat32(float32(val)), nil

		case schema_j5pb.FloatField_FORMAT_FLOAT64:
			return protoreflect.ValueOfFloat64(val), nil

		default:
			return pv, fmt.Errorf("unsupported float format %v", st.Float.Format)
		}

	case *schema_j5pb.Field_Bytes:
		switch val := value.(type) {
		case []byte:
			return protoreflect.ValueOfBytes(val), nil

		case string:
			return byteValueFromString(val)

		case *string:
			if val == nil {
				return protoreflect.Value{}, nil
			}
			return byteValueFromString(*val)

		default:
			return pv, fmt.Errorf("expected []byte, got %T", value)
		}

	case *schema_j5pb.Field_Timestamp:
		switch val := value.(type) {
		case string:
			return timestampFromString(val)

		case *string:
			if val == nil {
				return protoreflect.Value{}, nil
			}
			return timestampFromString(*val)

		case *timestamppb.Timestamp:
			return protoreflect.ValueOfMessage(val.ProtoReflect()), nil

		case time.Time:
			msg := timestamppb.New(val)
			return protoreflect.ValueOfMessage(msg.ProtoReflect()), nil

		default:
			return pv, fmt.Errorf("expected *timestamppb.Timestamp or time.Time, got %T", value)
		}

	case *schema_j5pb.Field_Decimal:
		switch val := value.(type) {
		case *decimal_j5t.Decimal:
			return protoreflect.ValueOfMessage(val.ProtoReflect()), nil

		case *decimal.Decimal:
			msg := decimal_j5t.FromShop(val)

			return protoreflect.ValueOfMessage(msg.ProtoReflect()), nil
		default:
			return pv, fmt.Errorf("expected *decimal_j5t.Decimal, got %T", value)
		}

	case *schema_j5pb.Field_Date:
		switch val := value.(type) {
		case *date_j5t.Date:
			return protoreflect.ValueOfMessage(val.ProtoReflect()), nil

		case string:
			dat, err := date_j5t.DateFromString(val)
			if err != nil {
				return pv, err
			}
			return protoreflect.ValueOfMessage(dat.ProtoReflect()), nil

		case *string:
			if val == nil {
				return protoreflect.Value{}, nil
			}
			dat, err := date_j5t.DateFromString(*val)
			if err != nil {
				return pv, err
			}
			return protoreflect.ValueOfMessage(dat.ProtoReflect()), nil

		default:
			return pv, fmt.Errorf("expected *date_j5t.Date, got %T", value)
		}

	default:
		return pv, fmt.Errorf("unsupported scalar type %T", schema.Type)
	}
}

func timestampFromString(val string) (protoreflect.Value, error) {
	t, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return protoreflect.Value{}, err
	}
	msg := timestamppb.New(t)
	return protoreflect.ValueOfMessage(msg.ProtoReflect()), nil
}

func byteValueFromString(val string) (protoreflect.Value, error) {

	// is base64, could be url or standard
	// Luck would have it, they are the same bar those two characters, so even if
	// it doesn't contain either it should work.
	if strings.ContainsAny(val, "+/") {
		// url
		b, err := base64.StdEncoding.DecodeString(val)
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfBytes(b), nil
	}

	b, err := base64.URLEncoding.DecodeString(val)
	if err != nil {
		return protoreflect.Value{}, err
	}
	return protoreflect.ValueOfBytes(b), nil
}

func scalarGoFromReflect(schema *schema_j5pb.Field, val protoreflect.Value) (interface{}, error) {
	switch st := schema.Type.(type) {
	case *schema_j5pb.Field_Any:
		return AnyValue(val.Interface()), nil

	case *schema_j5pb.Field_Bool:
		return val.Bool(), nil

	case *schema_j5pb.Field_String_:
		return val.String(), nil

	case *schema_j5pb.Field_Key:
		return KeyValue(val.String()), nil

	case *schema_j5pb.Field_Integer:
		switch st.Integer.Format {
		case schema_j5pb.IntegerField_FORMAT_INT32:
			return int32(val.Int()), nil

		case schema_j5pb.IntegerField_FORMAT_INT64:
			return val.Int(), nil

		case schema_j5pb.IntegerField_FORMAT_UINT32:
			return uint32(val.Uint()), nil

		case schema_j5pb.IntegerField_FORMAT_UINT64:
			return val.Uint(), nil

		default:
			return nil, fmt.Errorf("unsupported integer format %v", st.Integer.Format)
		}

	case *schema_j5pb.Field_Float:
		switch st.Float.Format {
		case schema_j5pb.FloatField_FORMAT_FLOAT32:
			return float32(val.Float()), nil

		case schema_j5pb.FloatField_FORMAT_FLOAT64:
			return val.Float(), nil

		default:
			return nil, fmt.Errorf("unsupported float format %v", st.Float.Format)
		}

	case *schema_j5pb.Field_Bytes:
		return val.Bytes(), nil

	case *schema_j5pb.Field_Date:
		msg := val.Message()
		val := &date_j5t.Date{}
		pv := val.ProtoReflect()
		copyReflect(msg, pv)
		return val, nil

	case *schema_j5pb.Field_Decimal:
		msg := val.Message()
		val := &decimal_j5t.Decimal{}
		pv := val.ProtoReflect()
		copyReflect(msg, pv)
		return val, nil

	case *schema_j5pb.Field_Timestamp:
		msg := val.Message()
		seconds := msg.Get(msg.Descriptor().Fields().ByName("seconds")).Int()
		nanos := msg.Get(msg.Descriptor().Fields().ByName("nanos")).Int()
		t := time.Unix(seconds, nanos).In(time.UTC)
		return t, nil

	default:
		return nil, fmt.Errorf("unsupported scalar type %T", st)
	}

}
