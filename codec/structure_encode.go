package codec

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/pentops/j5/internal/j5reflect"
	"github.com/pentops/j5/j5types/date_j5t"
	"github.com/pentops/j5/j5types/decimal_j5t"
)

func (enc *encoder) encodeObjectBody(fieldSet j5reflect.FieldSet) error {

	first := true
	enc.openObject()
	defer enc.closeObject()

	return fieldSet.RangeSetProperties(func(prop j5reflect.Property) error {
		if !first {
			enc.fieldSep()
		}
		first = false
		if err := enc.fieldLabel(prop.JSONName()); err != nil {
			return err
		}

		val, err := prop.Field()
		if err != nil {
			return err
		}
		if err := enc.encodeValue(val); err != nil {
			return err
		}

		return nil
	})

}

func (enc *encoder) encodeOneofBody(fieldSet j5reflect.FieldSet) error {

	prop, err := fieldSet.GetOne()
	if err != nil {
		return err
	}
	if prop == nil {
		return nil
	}

	enc.openObject()

	err = enc.fieldLabel("!type")
	if err != nil {
		return err
	}

	err = enc.addString(prop.JSONName())
	if err != nil {
		return err
	}

	enc.fieldSep()

	err = enc.fieldLabel(prop.JSONName())
	if err != nil {
		return err
	}

	val, err := prop.Field()
	if err != nil {
		return err
	}

	if err := enc.encodeValue(val); err != nil {
		return err
	}

	enc.closeObject()
	return nil
}

func (enc *encoder) encodeObject(object *j5reflect.Object) error {
	return enc.encodeObjectBody(object)
}

func (enc *encoder) encodeOneof(oneof *j5reflect.Oneof) error {
	return enc.encodeOneofBody(oneof)
}

func (enc *encoder) encodeValue(field j5reflect.Value) error {

	switch ft := field.(type) {
	case j5reflect.ObjectField:
		val, err := ft.Object()
		if err != nil {
			return err
		}
		return enc.encodeObject(val)

	case j5reflect.OneofField:
		val, err := ft.Oneof()
		if err != nil {
			return err
		}
		return enc.encodeOneof(val)

	case j5reflect.EnumField:
		return enc.encodeEnum(ft)

	case j5reflect.ArrayField:
		return enc.encodeArray(ft)

	case j5reflect.MapField:
		return enc.encodeMap(ft)

	case j5reflect.ScalarField:
		return enc.encodeScalarField(ft)

	default:
		return fmt.Errorf("encode value of type %q, unsupported", field.Type())
	}
}

func (enc *encoder) encodeMap(field j5reflect.MapField) error {
	enc.openObject()
	first := true
	defer enc.closeObject()
	return field.Range(func(key string, val j5reflect.Value) error {
		if !first {
			enc.fieldSep()
		}
		first = false

		err := enc.fieldLabel(key)
		if err != nil {
			return err
		}
		return enc.encodeValue(val)
	})
}

func (enc *encoder) encodeArray(array j5reflect.ArrayField) error {
	enc.openArray()
	defer enc.closeArray()
	first := true
	return array.Range(func(prop j5reflect.Value) error {
		if !first {
			enc.fieldSep()
		}
		first = false
		return enc.encodeValue(prop)
	})
}

func (enc *encoder) encodeEnum(enum j5reflect.EnumField) error {
	val, err := enum.GetValue()
	if err != nil {
		return err
	}
	return enc.addString(val.Name)

}

func (enc *encoder) encodeScalarField(scalar j5reflect.ScalarField) error {
	val := scalar.GetInterface()
	switch vt := val.(type) {
	case string:
		return enc.addString(vt)
	case bool:
		enc.addBool(vt)
		return nil
	case int32:
		enc.addInt(int64(vt))
		return nil
	case int64:
		enc.addInt(vt)
		return nil
	case uint32:
		enc.addUint(uint64(vt))
		return nil
	case uint64:
		enc.addUint(vt)
		return nil
	case float32:
		enc.addFloat(float64(vt), 32)
		return nil
	case float64:
		enc.addFloat(vt, 64)
		return nil
	case []byte:
		vv := base64.StdEncoding.EncodeToString(vt)
		return enc.addString(vv)

	case *date_j5t.Date:
		return enc.addString(vt.DateString())

	case *decimal_j5t.Decimal:
		return enc.addString(vt.Value)

	case time.Time:
		return enc.addString(vt.In(time.UTC).Format(time.RFC3339))

	default:
		return fmt.Errorf("unsupported scalar type %T", vt)

	}

}
