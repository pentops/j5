package codec

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/pentops/j5/j5types/date_j5t"
	"github.com/pentops/j5/j5types/decimal_j5t"
	"github.com/pentops/j5/lib/j5reflect"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (enc *encoder) encodeObjectBody(fieldSet j5reflect.PropertySet) error {
	first := true
	enc.openObject()
	defer enc.closeObject()

	return fieldSet.RangeValues(func(prop j5reflect.Field) error {
		if !first {
			enc.fieldSep()
		}
		first = false
		if err := enc.fieldLabel(prop.NameInParent()); err != nil {
			return err
		}

		if err := enc.encodeValue(prop); err != nil {
			return err
		}

		return nil
	})
}

func (enc *encoder) encodeOneofBody(fieldSet j5reflect.Oneof) error {

	prop, isSet, err := fieldSet.GetOne()
	if err != nil {
		return err
	}

	enc.openObject()
	defer enc.closeObject()
	if !isSet {
		return nil
	}

	err = enc.fieldLabel("!type")
	if err != nil {
		return err
	}

	err = enc.addString(prop.NameInParent())
	if err != nil {
		return err
	}

	enc.fieldSep()

	err = enc.fieldLabel(prop.NameInParent())
	if err != nil {
		return err
	}

	if err := enc.encodeValue(prop); err != nil {
		return err
	}

	return nil
}

func (enc *encoder) encodeObject(object j5reflect.Object) error {
	return enc.encodeObjectBody(object)
}

func (enc *encoder) encodeAny(anyField j5reflect.AnyField) error {
	val, err := anyField.GetJ5Any()
	if err != nil {
		return err
	}

	var jsonData []byte
	if val.J5Json != nil {
		jsonData = val.J5Json
	} else if val.Proto != nil {

		mt, err := enc.codec.resolver.FindMessageByName(protoreflect.FullName(val.TypeName))
		if err != nil {
			return fmt.Errorf("decoding any type %q: %w", val.TypeName, err)
		}

		dst := mt.New()
		if err := proto.Unmarshal(val.Proto, dst.Interface()); err != nil {
			return fmt.Errorf("decoding any type %q: %w", val.TypeName, err)
		}

		innerBytes, err := enc.codec.encode(dst)
		if err != nil {
			return err
		}
		jsonData = innerBytes
	}

	enc.openObject()
	defer enc.closeObject()

	err = enc.fieldLabel("!type")
	if err != nil {
		return err
	}

	err = enc.addString(val.TypeName)
	if err != nil {
		return err
	}

	enc.fieldSep()

	err = enc.fieldLabel("value")
	if err != nil {
		return err
	}

	enc.add(jsonData)

	return nil

}

func (enc *encoder) encodeValue(field j5reflect.Field) error {

	switch ft := field.(type) {
	case j5reflect.ObjectField:
		return enc.encodeObject(ft)

	case j5reflect.OneofField:
		return enc.encodeOneofBody(ft)

	case j5reflect.EnumField:
		return enc.encodeEnum(ft)

	case j5reflect.ArrayField:
		return enc.encodeArray(ft)

	case j5reflect.MapField:
		return enc.encodeMap(ft)

	case j5reflect.ScalarField:
		return enc.encodeScalarField(ft)

	case j5reflect.AnyField:
		return enc.encodeAny(ft)

	default:
		return fmt.Errorf("encode value of type %q, unsupported", field.FullTypeName())
	}
}

func (enc *encoder) encodeMap(field j5reflect.MapField) error {
	enc.openObject()
	first := true
	defer enc.closeObject()
	return field.Range(func(key string, val j5reflect.Field) error {
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
	return array.RangeValues(func(idx int, prop j5reflect.Field) error {
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
	return enc.addString(val.Name())

}

func (enc *encoder) encodeScalarField(scalar j5reflect.ScalarField) error {
	val, err := scalar.ToGoValue()
	if err != nil {
		return err
	}
	switch vt := val.(type) {

	case string:
		return enc.addString(vt)
	case bool:
		enc.addBool(vt)
		return nil
	case int32:
		enc.addInt32(vt)
		return nil
	case int64:
		enc.addInt64(vt)
		return nil
	case uint32:
		enc.addUint32(vt)
		return nil
	case uint64:
		enc.addUint64(vt)
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
