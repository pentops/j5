package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/j5types/decimal_j5t"
	"github.com/pentops/j5/lib/j5reflect/protoval"
	"github.com/pentops/j5/lib/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

/*** Interface ***/

type ScalarField interface {
	Field
	ToGoValue() (any, error)
	SetGoValue(value any) error
	SetASTValue(ASTValue) error
	IsSet() bool
}

type ArrayOfScalarField interface {
	ArrayField
	AppendGoValue(value any) (int, error)
	AppendASTValue(ASTValue) (int, error)
}

type MapOfScalarField interface {
	SetGoValue(key string, value any) error
	SetASTValue(key string, value ASTValue) error
}

/*** Implementation ***/

type scalarField struct {
	fieldDefaults
	fieldContext
	value  protoval.Value
	schema *j5schema.ScalarSchema
}

type scalarFieldFactory struct {
	schema *j5schema.ScalarSchema
}

func (f *scalarFieldFactory) buildField(field fieldContext, value protoval.Value) Field {
	return &scalarField{
		fieldContext: field,
		value:        value,
		schema:       f.schema,
	}
}

func (sf *scalarField) IsSet() bool {
	val, ok := sf.value.GetValue()
	if !ok {
		return false
	}

	// value isn't null, so include it
	if sf.PropertySchema().ExplicitlyOptional {
		return true
	}

	switch st := sf.schema.Proto.Type.(type) {
	case *schema_j5pb.Field_String_:
		return val.String() != ""
	case *schema_j5pb.Field_Bool:
		return val.Bool()
	case *schema_j5pb.Field_Bytes:
		return val.Bytes() != nil && len(val.Bytes()) > 0
	case *schema_j5pb.Field_Float:
		return val.Float() != 0.0
	case *schema_j5pb.Field_Integer:
		switch st.Integer.Format {
		case schema_j5pb.IntegerField_FORMAT_INT32, schema_j5pb.IntegerField_FORMAT_INT64:
			return val.Int() != 0
		case schema_j5pb.IntegerField_FORMAT_UINT32, schema_j5pb.IntegerField_FORMAT_UINT64:
			return val.Uint() != 0
		default:
			panic(fmt.Sprintf("unknown integer format %s", st.Integer.Format))
		}

	case *schema_j5pb.Field_Key:
		return val.String() != ""
	case *schema_j5pb.Field_Date:
		return val.Interface() != nil
	case *schema_j5pb.Field_Timestamp:
		return val.Interface() != nil
	}

	return ok

}

func (sf *scalarField) SetDefaultValue() error {
	switch st := sf.schema.Proto.Type.(type) {
	case *schema_j5pb.Field_String_:
		return sf.setValue(protoreflect.ValueOfString(""))
	case *schema_j5pb.Field_Bool:
		return sf.setValue(protoreflect.ValueOfBool(false))
	case *schema_j5pb.Field_Bytes:
		return sf.setValue(protoreflect.ValueOfBytes(nil))
	case *schema_j5pb.Field_Float:
		switch st.Float.Format {
		case schema_j5pb.FloatField_FORMAT_FLOAT32:
			return sf.setValue(protoreflect.ValueOfFloat32(0.0))
		case schema_j5pb.FloatField_FORMAT_FLOAT64:
			return sf.setValue(protoreflect.ValueOfFloat64(0.0))
		default:
			return fmt.Errorf("unknown float format %s", st.Float.Format)
		}
	case *schema_j5pb.Field_Integer:
		switch st.Integer.Format {
		case schema_j5pb.IntegerField_FORMAT_INT32:
			return sf.setValue(protoreflect.ValueOfInt32(0))
		case schema_j5pb.IntegerField_FORMAT_INT64:
			return sf.setValue(protoreflect.ValueOfInt64(0))
		case schema_j5pb.IntegerField_FORMAT_UINT32:
			return sf.setValue(protoreflect.ValueOfUint32(0))
		case schema_j5pb.IntegerField_FORMAT_UINT64:
			return sf.setValue(protoreflect.ValueOfUint64(0))
		default:
			return fmt.Errorf("unknown integer format %s", st.Integer.Format)

		}
	case *schema_j5pb.Field_Key:
		return sf.setValue(protoreflect.ValueOfString(""))
	case *schema_j5pb.Field_Date, *schema_j5pb.Field_Timestamp:
		return sf.setValue(protoreflect.ValueOf(nil))
	case *schema_j5pb.Field_Decimal:
		val := decimal_j5t.Zero()
		return sf.setValue(protoreflect.ValueOfMessage(val.ProtoReflect()))
	default:
		return fmt.Errorf("unsupported scalar type %T", st)
	}
}

func (sf *scalarField) AsScalar() (ScalarField, bool) {
	return sf, true
}

func (sf *scalarField) SetASTValue(value ASTValue) error {
	reflectValue, err := scalarReflectFromAST(sf.schema.Proto, value)
	if err != nil {
		return err
	}

	return sf.setValue(reflectValue)
}

func (sf *scalarField) setValue(reflectValue protoreflect.Value) error {
	return sf.value.SetValue(reflectValue)
}

func (sf *scalarField) SetGoValue(value any) error {
	reflectValue, err := scalarReflectFromGo(sf.schema.Proto, value)
	if err != nil {
		return fmt.Errorf("setting field %s: %w", sf.FullTypeName(), err)
	}
	return sf.setValue(reflectValue)
}

func (sf *scalarField) ToGoValue() (any, error) {
	val, ok := sf.value.GetValue()
	if !ok {
		propSchema := sf.PropertySchema()
		if propSchema.ExplicitlyOptional {
			return nil, nil
		}
		return defaultGoValue(sf.schema.Proto)
	}
	return scalarGoFromReflect(sf.schema.Proto, val)
}

type arrayOfScalarField struct {
	leafArrayField
	itemSchema *j5schema.ScalarSchema
}

var _ ArrayOfScalarField = (*arrayOfScalarField)(nil)

func (array *arrayOfScalarField) AsArray() (ArrayField, bool) {
	return array, true
}

func (array *arrayOfScalarField) AsArrayOfScalar() (ArrayOfScalarField, bool) {
	return array, true
}

func (array *arrayOfScalarField) AppendGoValue(value any) (int, error) {
	reflectValue, err := scalarReflectFromGo(array.itemSchema.Proto, value)
	if err != nil {
		return -1, err
	}
	return array.appendProtoValue(reflectValue), nil
}

func (array *arrayOfScalarField) AppendASTValue(value ASTValue) (int, error) {
	reflectValue, err := scalarReflectFromAST(array.itemSchema.Proto, value)
	if err != nil {
		return -1, err
	}
	return array.appendProtoValue(reflectValue), nil
}

type mapOfScalarField struct {
	leafMapField
	itemSchema *j5schema.ScalarSchema
}

var _ MapOfScalarField = (*mapOfScalarField)(nil)

func (mapField *mapOfScalarField) AsMap() (MapField, bool) {
	return mapField, true
}

func (mapField *mapOfScalarField) AsMapOfScalar() (MapOfScalarField, bool) {
	return mapField, true
}

func (mapField *mapOfScalarField) SetGoValue(key string, value any) error {
	reflVal, err := scalarReflectFromGo(mapField.itemSchema.Proto, value)
	if err != nil {
		return fmt.Errorf("converting value to proto: %w", err)
	}
	return mapField.value.ValueAt(key).SetValue(reflVal)
}

func (mapField *mapOfScalarField) SetASTValue(key string, value ASTValue) error {
	reflVal, err := scalarReflectFromAST(mapField.itemSchema.Proto, value)

	if err != nil {
		return fmt.Errorf("converting value to proto: %w", err)
	}
	return mapField.value.ValueAt(key).SetValue(reflVal)
}
