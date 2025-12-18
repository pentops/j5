package j5reflect

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/j5types/date_j5t"
	"github.com/pentops/j5/j5types/decimal_j5t"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AnyValue any

func scalarReflectFromGo(schema *schema_j5pb.Field, value any) (protoreflect.Value, error) {
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

		case nil:
			return protoreflect.Value{}, nil

		default:
			return pv, fmt.Errorf("expected KeyValue, got %T", value)
		}

	case *schema_j5pb.Field_Integer:
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Pointer {
			if rv.IsNil() {
				return protoreflect.Value{}, nil
			}
			rv = rv.Elem()
			value = rv.Interface()
		}

		if numVal, ok := value.(json.Number); ok {
			i64, err := numVal.Int64()
			if err != nil {
				return pv, err
			}
			value = i64
		}

		switch st.Integer.Format {
		case schema_j5pb.IntegerField_FORMAT_INT32:
			switch val := value.(type) {
			case uint:
				if val > math.MaxInt32 {
					return pv, fmt.Errorf("int value %v is out of range for int32", val)
				}
				return protoreflect.ValueOfInt32(int32(val)), nil
			case uint16:
				return protoreflect.ValueOfInt32(int32(val)), nil
			case uint32:
				if val > math.MaxInt32 {
					return pv, fmt.Errorf("int value %v is out of range for int32", val)
				}
				return protoreflect.ValueOfInt32(int32(val)), nil
			case uint64:
				if val > math.MaxInt32 {
					return pv, fmt.Errorf("int value %v is out of range for int32", val)
				}
				return protoreflect.ValueOfInt32(int32(val)), nil
			case int:
				if val > math.MaxInt32 || val < math.MinInt32 {
					return pv, fmt.Errorf("int value %v is out of range for int32", val)
				}
				return protoreflect.ValueOfInt32(int32(val)), nil
			case int16:
				return protoreflect.ValueOfInt32(int32(val)), nil
			case int32:
				return protoreflect.ValueOfInt32(val), nil
			case int64:
				if val > math.MaxInt32 || val < math.MinInt32 {
					return pv, fmt.Errorf("int value %v is out of range for int32", val)
				}
				return protoreflect.ValueOfInt32(int32(val)), nil
			case string:
				valAsInt, err := strconv.ParseInt(val, 10, 32)
				if err != nil {
					return pv, nil
				}

				return protoreflect.ValueOfInt32(int32(valAsInt)), nil
			default:
				return pv, fmt.Errorf("expected int, got %T", value)
			}

		case schema_j5pb.IntegerField_FORMAT_INT64:
			switch val := value.(type) {
			case uint:
				return protoreflect.ValueOfInt64(int64(val)), nil
			case uint16:
				return protoreflect.ValueOfInt64(int64(val)), nil
			case uint32:
				return protoreflect.ValueOfInt64(int64(val)), nil
			case uint64:
				return protoreflect.ValueOfInt64(int64(val)), nil
			case int:
				return protoreflect.ValueOfInt64(int64(val)), nil
			case int16:
				return protoreflect.ValueOfInt64(int64(val)), nil
			case int32:
				return protoreflect.ValueOfInt64(int64(val)), nil
			case int64:
				return protoreflect.ValueOfInt64(val), nil
			case string:
				valAsInt, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					return pv, nil
				}

				return protoreflect.ValueOfInt64(valAsInt), nil
			default:
				return pv, fmt.Errorf("expected int, got %T", value)
			}

		case schema_j5pb.IntegerField_FORMAT_UINT32:
			switch val := value.(type) {
			case uint:
				if val > math.MaxUint32 {
					return pv, fmt.Errorf("int value %v is out of range for uint32", val)
				}
				return protoreflect.ValueOfUint32(uint32(val)), nil
			case uint16:
				return protoreflect.ValueOfUint32(uint32(val)), nil
			case uint32:
				return protoreflect.ValueOfUint32(val), nil
			case uint64:
				if val > math.MaxUint32 {
					return pv, fmt.Errorf("int value %v is out of range for uint32", val)
				}
				return protoreflect.ValueOfUint32(uint32(val)), nil
			case int:
				if val < 0 {
					return pv, fmt.Errorf("int value %v is out of range for uint32", val)
				}
				if val > math.MaxUint32 {
					return pv, fmt.Errorf("int value %v is out of range for uint32", val)
				}
				return protoreflect.ValueOfUint32(uint32(val)), nil
			case int16:
				if val < 0 {
					return pv, fmt.Errorf("int value %v is out of range for uint32", val)
				}
				return protoreflect.ValueOfUint32(uint32(val)), nil
			case int32:
				if val < 0 {
					return pv, fmt.Errorf("int value %v is out of range for uint32", val)
				}
				return protoreflect.ValueOfUint32(uint32(val)), nil
			case int64:
				if val < 0 {
					return pv, fmt.Errorf("int value %v is out of range for uint32", val)
				}
				if val > math.MaxUint32 {
					return pv, fmt.Errorf("int value %v is out of range for uint32", val)
				}
				return protoreflect.ValueOfUint32(uint32(val)), nil
			case string:
				valAsInt, err := strconv.ParseUint(val, 10, 32)
				if err != nil {
					return pv, nil
				}

				return protoreflect.ValueOfUint32(uint32(valAsInt)), nil

			default:
				return pv, fmt.Errorf("expected uint32, got %T", value)
			}

		case schema_j5pb.IntegerField_FORMAT_UINT64:
			switch val := value.(type) {
			case uint:
				return protoreflect.ValueOfUint64(uint64(val)), nil
			case uint16:
				return protoreflect.ValueOfUint64(uint64(val)), nil
			case uint32:
				return protoreflect.ValueOfUint64(uint64(val)), nil
			case uint64:
				return protoreflect.ValueOfUint64(val), nil
			case int:
				if val < 0 {
					return pv, fmt.Errorf("int value %v is out of range for uint64", val)
				}
				return protoreflect.ValueOfUint64(uint64(val)), nil
			case int16:
				if val < 0 {
					return pv, fmt.Errorf("int value %v is out of range for uint64", val)
				}
				return protoreflect.ValueOfUint64(uint64(val)), nil
			case int32:
				if val < 0 {
					return pv, fmt.Errorf("int value %v is out of range for uint64", val)
				}
				return protoreflect.ValueOfUint64(uint64(val)), nil
			case int64:
				if val < 0 {
					return pv, fmt.Errorf("int value %v is out of range for uint64", val)
				}
				return protoreflect.ValueOfUint64(uint64(val)), nil
			case string:
				valAsInt, err := strconv.ParseUint(val, 10, 64)
				if err != nil {
					return pv, nil
				}

				return protoreflect.ValueOfUint64(valAsInt), nil

			default:
				return pv, fmt.Errorf("expected uint64, got %T", value)
			}

		default:
			return pv, fmt.Errorf("unsupported integer format %v", st.Integer.Format)
		}

	case *schema_j5pb.Field_Float:
		switch val := value.(type) {
		case json.Number:
			var err error
			value, err = val.Float64()
			if err != nil {
				return pv, err
			}
		case string:
			var err error
			value, err = strconv.ParseFloat(val, 64)
			if err != nil {
				return pv, err
			}
		}

		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Pointer {
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
		case string:
			return decimalFromString(val)

		case *string:
			if val == nil {
				return protoreflect.Value{}, nil
			}
			return decimalFromString(*val)

		case json.Number:
			return decimalFromString(val.String())

		case *decimal_j5t.Decimal:
			return protoreflect.ValueOfMessage(val.ProtoReflect()), nil

		case *decimal.Decimal:
			msg := decimal_j5t.FromShop(*val)
			return protoreflect.ValueOfMessage(msg.ProtoReflect()), nil

		case decimal.Decimal:
			msg := decimal_j5t.FromShop(val)

			return protoreflect.ValueOfMessage(msg.ProtoReflect()), nil
		default:
			return pv, fmt.Errorf("expected decimal got %T", value)
		}

	case *schema_j5pb.Field_Date:
		switch val := value.(type) {
		case *date_j5t.Date:
			return protoreflect.ValueOfMessage(val.ProtoReflect()), nil

		case string:
			return dateFromString(val)

		case *string:
			if val == nil {
				return protoreflect.Value{}, nil
			}
			return dateFromString(*val)

		default:
			return pv, fmt.Errorf("expected *date_j5t.Date, got %T", value)
		}

	default:
		return pv, fmt.Errorf("unsupported scalar type %T", schema.Type)
	}
}

func decimalFromString(val string) (protoreflect.Value, error) {
	if val == "" {
		return protoreflect.ValueOf(nil), nil

	}
	d, err := decimal.NewFromString(val)
	if err != nil {
		return protoreflect.Value{}, err
	}
	msg := decimal_j5t.FromShop(d)
	return protoreflect.ValueOfMessage(msg.ProtoReflect()), nil
}

func timestampFromString(val string) (protoreflect.Value, error) {
	t, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return protoreflect.Value{}, err
	}
	msg := timestamppb.New(t)
	return protoreflect.ValueOfMessage(msg.ProtoReflect()), nil
}

func dateFromString(val string) (protoreflect.Value, error) {
	d, err := date_j5t.DateFromString(val)
	if err != nil {
		return protoreflect.Value{}, err
	}
	return protoreflect.ValueOfMessage(d.ProtoReflect()), nil
}

func byteValueFromString(val string) (protoreflect.Value, error) {
	// is base64, could be url or standard
	// Luck would have it, they are the same bar those two characters, so even if
	// it doesn't contain either it should work.
	val = strings.ReplaceAll(val, "-", "+")
	val = strings.ReplaceAll(val, "_", "/")

	if len(val)%4 != 0 {
		val += strings.Repeat("=", 4-len(val)%4)
	}

	b, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return protoreflect.Value{}, err
	}
	return protoreflect.ValueOfBytes(b), nil
}

func defaultGoValue(schema *schema_j5pb.Field) (any, error) {
	switch st := schema.Type.(type) {
	case *schema_j5pb.Field_Any:
		return AnyValue(nil), nil

	case *schema_j5pb.Field_Bool:
		return false, nil

	case *schema_j5pb.Field_String_:
		return "", nil

	case *schema_j5pb.Field_Key:
		return "", nil

	case *schema_j5pb.Field_Integer:
		switch st.Integer.Format {
		case schema_j5pb.IntegerField_FORMAT_INT32:
			return int32(0), nil
		case schema_j5pb.IntegerField_FORMAT_INT64:
			return int64(0), nil
		case schema_j5pb.IntegerField_FORMAT_UINT32:
			return uint32(0), nil
		case schema_j5pb.IntegerField_FORMAT_UINT64:
			return uint64(0), nil
		default:
			return nil, fmt.Errorf("unsupported integer format %v", st.Integer.Format)
		}

	case *schema_j5pb.Field_Float:
		switch st.Float.Format {
		case schema_j5pb.FloatField_FORMAT_FLOAT32:
			return float32(0), nil
		case schema_j5pb.FloatField_FORMAT_FLOAT64:
			return float64(0), nil
		default:
			return nil, fmt.Errorf("unsupported float format %v", st.Float.Format)
		}

	case *schema_j5pb.Field_Bytes:
		return []byte{}, nil

	case *schema_j5pb.Field_Date:
		return &date_j5t.Date{}, nil

	case *schema_j5pb.Field_Decimal:
		return &decimal_j5t.Decimal{}, nil

	case *schema_j5pb.Field_Timestamp:
		return time.Time{}, nil

	default:
		return nil, fmt.Errorf("unsupported scalar type %T", st)
	}
}

func scalarGoFromReflect(schema *schema_j5pb.Field, val protoreflect.Value) (any, error) {
	switch st := schema.Type.(type) {
	case *schema_j5pb.Field_Any:
		return AnyValue(val.Interface()), nil

	case *schema_j5pb.Field_Bool:
		return val.Bool(), nil

	case *schema_j5pb.Field_String_:
		return val.String(), nil

	case *schema_j5pb.Field_Key:
		return val.String(), nil

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
		if !val.IsValid() {
			return nil, nil
		}
		msg := val.Message()
		val := &decimal_j5t.Decimal{}
		pv := val.ProtoReflect()
		copyReflect(msg, pv)
		if val.Value == "" {
			return nil, nil // empty decimal
		}
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

func copyReflect(a, b protoreflect.Message) {
	bFields := b.Descriptor().Fields()
	a.Range(func(fd protoreflect.FieldDescriptor, val protoreflect.Value) bool {
		bField := bFields.ByNumber(fd.Number())
		if bField == nil || bField.Kind() != fd.Kind() || bField.Name() != fd.Name() {
			panic(fmt.Sprintf("CopyReflect: field %s not found in %s", fd.FullName(), b.Descriptor().FullName()))
		}
		b.Set(bField, val)
		return true
	})
}
