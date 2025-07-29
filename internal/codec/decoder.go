package codec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/pentops/j5/j5types/any_j5t"
	"github.com/pentops/j5/lib/j5reflect"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func (c *Codec) decode(jsonData []byte, msg protoreflect.Message) error {
	root, err := c.refl.NewRoot(msg)
	if err != nil {
		return err
	}

	return c.decodeRoot(jsonData, root)
}

func (c *Codec) decodeRoot(jsonData []byte, root j5reflect.Root) error {
	dec := json.NewDecoder(bytes.NewReader(jsonData))
	dec.UseNumber()
	d2 := &decoder{
		jd:    dec,
		codec: c,
	}

	switch schema := root.(type) {
	case j5reflect.Object:
		return d2.decodeObject(schema)
	case j5reflect.Oneof:
		return d2.decodeOneof(schema)
	default:
		return fmt.Errorf("unsupported root schema type %T", schema)
	}
}

// decoder is an instance for decoding a single message, not reusable.
type decoder struct {
	jd    *json.Decoder
	codec *Codec
}

func (d *decoder) Token() (json.Token, error) {
	return d.jd.Token()
}

func (dec *decoder) expectDelimOrNull(delim rune) (isNull bool, err error) {

	tok, err := dec.Token()
	if err != nil {
		return false, err
	}

	if tok == nil {
		return true, nil
	}

	if tok != json.Delim(delim) {
		return false, unexpectedTokenError(tok, string(delim))
	}
	return false, nil
}

func (dec *decoder) expectDelim(delim rune) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}

	if tok != json.Delim(delim) {
		return unexpectedTokenError(tok, string(delim))
	}
	return nil
}

func (dec *decoder) jsonObjectBody(callback func(key string) error) error {
	for dec.jd.More() {
		keyToken, err := dec.Token()
		if err != nil {
			return err
		}

		keyTokenStr, ok := keyToken.(string)
		if !ok {
			return unexpectedTokenError(keyToken, "string (object key)")
		}

		if err := callback(keyTokenStr); err != nil {
			return passUpError(keyTokenStr, err)
		}
	}
	return nil
}

func (dec *decoder) popValueAsBytes() (json.RawMessage, error) {
	raw := &json.RawMessage{}
	if err := dec.jd.Decode(raw); err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := json.Compact(buf, *raw); err != nil {
		return nil, newFieldError("value", err.Error())
	}
	return json.RawMessage(buf.Bytes()), nil
}

type fieldError struct {
	pathToField []string
	err         error
}

func newFieldError(field, message string) error {
	return fieldError{
		pathToField: []string{field},
		err:         errors.New(message),
	}
}

func unexpectedTokenError(got, expected any) error {
	return fieldError{
		pathToField: []string{fmt.Sprint(got)},
		err:         fmt.Errorf("unexpected token %v, expected %v", got, expected),
	}
}

func (e fieldError) Error() string {
	return fmt.Sprintf("field %s: %s", strings.Join(e.pathToField, "."), e.err.Error())
}

func (e fieldError) parent(field string) error {
	return fieldError{
		pathToField: append([]string{field}, e.pathToField...),
		err:         e.err,
	}
}

func passUpError(field string, err error) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(fieldError); ok {
		return e.parent(field)
	}
	return fieldError{
		err:         err,
		pathToField: []string{field},
	}
}

func (dec *decoder) decodeValue(prop j5reflect.Property) error {
	switch prop.PropertyType() {
	case j5reflect.MapProperty:
		return dec.decodeMapProperty(prop)

	case j5reflect.ArrayProperty:
		return dec.decodeArrayProperty(prop)

	case j5reflect.ObjectProperty:
		return dec.decodeObjectProperty(prop)

	case j5reflect.OneofProperty:
		return dec.decodeOneofProperty(prop)

	case j5reflect.EnumProperty:
		return dec.decodeEnum(prop)

	case j5reflect.ScalarProperty:
		return dec.decodeScalar(prop)

	case j5reflect.AnyProperty:
		return dec.decodeAny(prop)

	case j5reflect.PolymorphProperty:
		return dec.decodePolymorphProperty(prop)

	default:
		return fmt.Errorf("unknown field schema type %v", prop.PropertyType())
	}

}

func (dec *decoder) decodeObject(object j5reflect.PropertySet) error {
	err := dec.expectDelim('{')
	if err != nil {
		return err
	}
	err = dec.decodeObjectInner(object)
	if err != nil {
		return err
	}

	return dec.expectDelim('}')
}

func (dec *decoder) decodeObjectProperty(prop j5reflect.Property) error {
	wasNull, err := dec.expectDelimOrNull('{')
	if err != nil {
		return err
	}

	if wasNull {
		return nil
	}

	genField, err := prop.Field()
	if err != nil {
		return err
	}
	err = genField.SetDefaultValue()
	if err != nil {
		return err
	}

	object, ok := genField.AsObject()
	if !ok {
		return fmt.Errorf("object property produced non-object field")
	}

	if err := dec.decodeObjectInner(object); err != nil {
		return err
	}

	return dec.expectDelim('}')
}

func (dec *decoder) decodeObjectInner(object j5reflect.PropertySet) error {
	err := dec.jsonObjectBody(func(keyTokenStr string) error {
		prop, err := object.GetProperty(keyTokenStr)
		if err != nil {
			return err //errors.New("no such field")
		}
		if err := dec.decodeValue(prop); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return passUpError("object", err)
	}
	return nil
}

func (dec *decoder) decodeOneof(oneof j5reflect.Oneof) error {
	err := dec.expectDelim('{')
	if err != nil {
		return err
	}
	err = dec.decodeOneofInner(oneof)
	if err != nil {
		return err
	}
	return dec.expectDelim('}')
}

func (dec *decoder) decodeOneofProperty(prop j5reflect.Property) error {
	wasNull, err := dec.expectDelimOrNull('{')
	if err != nil {
		return err
	}

	if wasNull {
		return nil
	}

	genField, err := prop.Field()
	if err != nil {
		return err
	}
	err = genField.SetDefaultValue()
	if err != nil {
		return err
	}

	oneof, ok := genField.AsOneof()
	if !ok {
		return fmt.Errorf("oneof property produced non-oneof field")
	}

	err = dec.decodeOneofInner(oneof)
	if err != nil {
		return err
	}

	return dec.expectDelim('}')
}

func (dec *decoder) decodeOneofInner(oneof j5reflect.Oneof) error {

	foundKeys := []string{}
	var constrainType *string

	if err := dec.jsonObjectBody(func(keyTokenStr string) error {

		// !type is an optional parameter, when the consumer sets it we validate
		// it matches the type they actually sent.
		if keyTokenStr == "!type" {
			tok, err := dec.Token()
			if err != nil {
				return err
			}
			str, ok := tok.(string)
			if !ok {
				return unexpectedTokenError(tok, "string")
			}
			constrainType = &str
			return nil
		}

		matchedProperty, err := oneof.GetProperty(keyTokenStr)
		if err != nil {
			return newFieldError(keyTokenStr, "no such key")
		}
		foundKeys = append(foundKeys, keyTokenStr)

		if err := dec.decodeValue(matchedProperty); err != nil {
			return err
		}

		return nil

	}); err != nil {
		return err
	}

	if len(foundKeys) == 0 {
		if constrainType == nil {
			return nil // if it's required, validation picks that up later
		}
		keyTokenStr := *constrainType

		// Special case, allows the consumer to set a nil value on a oneof
		// just by using the type parameter
		prop, err := oneof.GetProperty(keyTokenStr)
		if err != nil {
			return newFieldError(keyTokenStr, "no such key")
		}
		field, err := prop.Field()
		if err != nil {
			return err
		}
		err = field.SetDefaultValue()
		if err != nil {
			return newFieldError(keyTokenStr, err.Error())
		}

	}

	if len(foundKeys) > 1 {
		return newFieldError(strings.Join(foundKeys, ", "), "multiple keys found in oneof")
	}

	if constrainType != nil && foundKeys[0] != *constrainType {
		return newFieldError(foundKeys[0], fmt.Sprintf("key %q does not match type %q", foundKeys[0], *constrainType))
	}

	return nil
}

func (dec *decoder) decodeScalar(prop j5reflect.Property) error {
	token, err := dec.Token()
	if err != nil {
		return err
	}

	genProp, err := prop.Field()
	if err != nil {
		return err
	}
	// token is nil when it's literally 'null' in the JSON
	if token == nil {
		return nil
	}

	field, ok := genProp.AsScalar()
	if !ok {
		return fmt.Errorf("scalar property produced non-scalar field")
	}

	if _, ok := token.(json.Delim); ok {
		return unexpectedTokenError(token, "scalar")
	}

	return field.SetGoValue(token)
}

func (dec *decoder) decodeEnum(prop j5reflect.Property) error {
	token, err := dec.Token()
	if err != nil {
		return err
	}

	if token == nil {
		return nil
	}

	genField, err := prop.Field()
	if err != nil {
		return err
	}

	field, ok := genField.AsEnum()
	if !ok {
		return fmt.Errorf("enum property produced non-enum field")
	}

	stringVal, ok := token.(string)
	if !ok {
		return unexpectedTokenError(token, "string")
	}
	if stringVal == "" {
		return nil
	}

	return field.SetFromString(stringVal)
}

func (dec *decoder) decodePolymorphProperty(prop j5reflect.Property) error {

	wasNull, err := dec.expectDelimOrNull('{')
	if err != nil {
		return err
	}

	if wasNull {
		return nil
	}

	genWrapper, err := prop.Field()
	if err != nil {
		return err
	}

	polyWrapper, ok := genWrapper.AsPolymorph()
	if !ok {
		return fmt.Errorf("any property produced non poly field")
	}

	field, err := polyWrapper.Unwrap()
	if err != nil {
		return err
	}

	err = dec.decodeAnyInner(field)
	if err != nil {
		return err
	}

	return dec.expectDelim('}')

}

func (dec *decoder) decodeAny(prop j5reflect.Property) error {

	wasNull, err := dec.expectDelimOrNull('{')
	if err != nil {
		return err
	}

	if wasNull {
		return nil
	}

	genField, err := prop.Field()
	if err != nil {
		return err
	}

	field, ok := genField.AsAny()
	if !ok {
		return fmt.Errorf("any property produced non-any field")
	}

	err = dec.decodeAnyInner(field)
	if err != nil {
		return err
	}

	return dec.expectDelim('}')
}

func (dec *decoder) decodeAnyInner(field j5reflect.AnyField) error {

	var valueBytes []byte
	var constrainType *string

	if err := dec.jsonObjectBody(func(keyTokenStr string) error {

		// !type is an optional parameter, when the consumer sets it we validate
		// it matches the type they actually sent.
		if keyTokenStr == "!type" {
			tok, err := dec.Token()
			if err != nil {
				return err
			}
			str, ok := tok.(string)
			if !ok {
				return unexpectedTokenError(tok, "string")
			}
			constrainType = &str
			return nil
		}

		if valueBytes != nil {
			return newFieldError(keyTokenStr, "multiple keys found in Any")
		}

		var err error
		valueBytes, err = dec.popValueAsBytes()
		if err != nil {
			return err
		}

		return nil

	}); err != nil {
		return err
	}

	if constrainType == nil {
		return newFieldError("value", "no type found in Any")
	}
	if valueBytes == nil {
		return newFieldError("value", "no value found in Any")
	}

	anyVal := &any_j5t.Any{
		J5Json:   valueBytes,
		TypeName: *constrainType,
	}

	if dec.codec.addProtoToAny && dec.codec.resolver != nil {
		// takes the PROTO name, which should match the encoder.
		innerDesc, err := dec.codec.resolver.FindMessageByName(protoreflect.FullName(*constrainType))
		if err != nil {
			if err == protoregistry.NotFound {
				return newFieldError(*constrainType, fmt.Sprintf("no type %q in registry", *constrainType))
			}
			return newFieldError(*constrainType, err.Error())
		}
		msg := innerDesc.New()

		if err := dec.codec.decode(valueBytes, msg); err != nil {
			return newFieldError(*constrainType, err.Error())
		}

		protoBytes, err := proto.Marshal(msg.Interface())
		if err != nil {
			return newFieldError(*constrainType, err.Error())
		}

		anyVal.Proto = protoBytes
	}

	err := field.SetJ5Any(anyVal)
	if err != nil {
		return newFieldError("value", err.Error())
	}

	return nil
}

func (dec *decoder) decodeMapProperty(prop j5reflect.Property) error {

	wasNull, err := dec.expectDelimOrNull('{')
	if err != nil {
		return err
	}

	if wasNull {
		return nil
	}

	field, err := prop.Field()
	if err != nil {
		return err
	}
	mapField, ok := field.AsMap()
	if !ok {
		return fmt.Errorf("map property produced non-map field")
	}

	err = dec.decodeMapField(mapField)
	if err != nil {
		return err
	}

	return dec.expectDelim('}')
}

func (dec *decoder) decodeMapField(field j5reflect.MapField) error {

	switch field := field.(type) {
	case j5reflect.MapOfScalarField:
		return dec.jsonObjectBody(func(keyTokenStr string) error {
			tok, err := dec.Token()
			if err != nil {
				return err
			}

			if _, ok := tok.(json.Delim); ok {
				return unexpectedTokenError(tok, "scalar")
			}

			return field.SetGoValue(keyTokenStr, tok)
		})

	case j5reflect.MapOfEnumField:
		return dec.jsonObjectBody(func(keyTokenStr string) error {
			tok, err := dec.Token()
			if err != nil {
				return err
			}

			str, ok := tok.(string)
			if !ok {
				return unexpectedTokenError(tok, "string")
			}

			return field.SetEnum(keyTokenStr, str)
		})

	case j5reflect.MapOfObjectField:
		return dec.jsonObjectBody(func(keyTokenStr string) error {
			subMsg, err := field.NewObjectElement(keyTokenStr)
			if err != nil {
				return err
			}
			return dec.decodeObject(subMsg)
		})

	case j5reflect.MapOfOneofField:
		return dec.jsonObjectBody(func(keyTokenStr string) error {
			subMsg, err := field.NewOneofElement(keyTokenStr)
			if err != nil {
				return err
			}
			return dec.decodeOneof(subMsg)
		})

	default:
		return fmt.Errorf("unknown map schema type %T", field)

	}
}

func (dec *decoder) decodeArrayProperty(prop j5reflect.Property) error {

	wasNull, err := dec.expectDelimOrNull('[')
	if err != nil {
		return err
	}
	if wasNull {
		return nil
	}

	genField, err := prop.Field()
	if err != nil {
		return err
	}

	field, ok := genField.AsArray()
	if !ok {
		return fmt.Errorf("array property produced non-array field %T", genField)

	}

	for dec.jd.More() {
		err = dec.decodeArrayFieldValue(field)
		if err != nil {
			return err
		}
	}

	return dec.expectDelim(']')

}

func (dec *decoder) decodeArrayFieldValue(field j5reflect.ArrayField) error {

	if field, ok := field.AsArrayOfScalar(); ok {
		tok, err := dec.Token()
		if err != nil {
			return err
		}

		if _, ok := tok.(json.Delim); ok {
			return unexpectedTokenError(tok, "scalar")
		}

		_, err = field.AppendGoValue(tok)
		return err

	}

	if field, ok := field.AsArrayOfObject(); ok {
		subMsg, _ := field.NewObjectElement() // ignoring index, not error
		return dec.decodeObject(subMsg)

	}

	if field, ok := field.AsArrayOfOneof(); ok {
		subMsg, _, err := field.NewOneofElement()
		if err != nil {
			return err
		}
		return dec.decodeOneof(subMsg)
	}

	return fmt.Errorf("unknown array schema type %T", field)

}
