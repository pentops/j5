package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type ASTValue interface {
	AsBool() (bool, error)
	AsString() (string, error)
	AsInt(bits int) (int64, error)
	AsUint(bits int) (uint64, error)
	AsFloat(bits int) (float64, error)
}

func scalarReflectFromAST(schema *schema_j5pb.Field, value ASTValue) (protoreflect.Value, error) {
	var pv protoreflect.Value
	switch st := schema.Type.(type) {
	case *schema_j5pb.Field_Any:
		pv = protoreflect.ValueOf(value)
		// and hope for the best?
		return pv, nil

	case *schema_j5pb.Field_Bool:
		val, err := value.AsBool()
		if err != nil {
			return pv, err
		}

		return protoreflect.ValueOfBool(val), nil

	case *schema_j5pb.Field_String_:
		val, err := value.AsString()
		if err != nil {
			return pv, err
		}

		return protoreflect.ValueOfString(val), nil

	case *schema_j5pb.Field_Key:
		val, err := value.AsString()
		if err != nil {
			return pv, err
		}

		return protoreflect.ValueOfString(val), nil

	case *schema_j5pb.Field_Integer:

		switch st.Integer.Format {
		case schema_j5pb.IntegerField_FORMAT_INT32:
			val, err := value.AsInt(32)
			if err != nil {
				return pv, err
			}

			return protoreflect.ValueOfInt32(int32(val)), nil

		case schema_j5pb.IntegerField_FORMAT_INT64:
			val, err := value.AsInt(32)
			if err != nil {
				return pv, err
			}

			return protoreflect.ValueOfInt64(int64(val)), nil

		case schema_j5pb.IntegerField_FORMAT_UINT32:
			val, err := value.AsUint(32)
			if err != nil {
				return pv, err
			}

			return protoreflect.ValueOfUint32(uint32(val)), nil

		case schema_j5pb.IntegerField_FORMAT_UINT64:
			val, err := value.AsUint(64)
			if err != nil {
				return pv, err
			}

			return protoreflect.ValueOfUint64(val), nil

		default:
			return pv, fmt.Errorf("unsupported integer format %v", st.Integer.Format)
		}

	case *schema_j5pb.Field_Float:
		switch st.Float.Format {

		case schema_j5pb.FloatField_FORMAT_FLOAT32:
			val, err := value.AsFloat(32)
			if err != nil {
				return pv, err
			}

			return protoreflect.ValueOfFloat32(float32(val)), nil

		case schema_j5pb.FloatField_FORMAT_FLOAT64:

			val, err := value.AsFloat(64)
			if err != nil {
				return pv, err
			}

			return protoreflect.ValueOfFloat64(val), nil

		default:
			return pv, fmt.Errorf("unsupported float format %v", st.Float.Format)
		}

	case *schema_j5pb.Field_Bytes:
		val, err := value.AsString()
		if err != nil {
			return pv, err
		}
		return byteValueFromString(val)

	case *schema_j5pb.Field_Timestamp:
		val, err := value.AsString()
		if err != nil {
			return pv, err
		}
		return timestampFromString(val)

	case *schema_j5pb.Field_Decimal:
		val, err := value.AsString()
		if err != nil {
			return pv, err
		}
		return decimalFromString(val)

	case *schema_j5pb.Field_Date:
		val, err := value.AsString()
		if err != nil {
			return pv, err
		}
		return dateFromString(val)

	default:
		return pv, fmt.Errorf("unsupported scalar type %T (AST)", schema.Type)
	}
}
