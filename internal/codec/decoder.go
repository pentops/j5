package codec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/pentops/j5/j5types/any_j5t"
	"github.com/pentops/j5/lib/j5reflect"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func (c *Codec) decode(jsonData []byte, msg protoreflect.Message) error {
	dec := json.NewDecoder(bytes.NewReader(jsonData))
	dec.UseNumber()
	d2 := &decoder{
		jd:    dec,
		codec: c,
	}

	root, err := c.refl.NewRoot(msg)
	if err != nil {
		return err
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

// working backwards from the standard library... I'm scared there is a really
// good reason they didn't expose Peek() on the decoder.

func isSpace(c byte) bool {
	return c <= ' ' && (c == ' ' || c == '\t' || c == '\r' || c == '\n')
}

func (dec *decoder) nextIsNull() (bool, error) {
	rd := dec.jd.Buffered()

	expect := []byte{':', 'n', 'u', 'l', 'l'}
	offset := 0

	dd := make([]byte, 1)
	for {
		ll, err := rd.Read(dd)
		if err != nil {
			// TODO: Buffered only returns the bytes which have been read
			// by the decoder from the underlying reader... even though the underlying
			// reader is also a byte slice... anyway.
			// If the 'null' is split across two reads, it will return false
			// incorrectly.
			// This is better than the alternative which is to return the EOF
			// error even when the next value is not 'null' which comes up far
			// more often.
			// The fix involves a rethink of how objects are parsed and is
			// fairly urgent.
			if err == io.EOF {
				return false, nil
			}
			return false, err
		}
		if ll == 0 {
			return false, nil
		}
		if isSpace(dd[0]) {
			continue
		}
		if dd[0] != expect[offset] {
			return false, nil
		}
		offset++
		if offset == len(expect) {
			return true, nil
		}
	}

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

func (dec *decoder) jsonArray(callback func() error) error {
	if err := dec.expectDelim('['); err != nil {
		return err
	}

	for dec.jd.More() {
		if err := callback(); err != nil {
			return err
		}
	}

	return dec.expectDelim(']')
}

func (dec *decoder) jsonObject(callback func(key string) error) error {
	if err := dec.expectDelim('{'); err != nil {
		return err
	}

	for dec.jd.More() {
		keyToken, err := dec.Token()
		if err != nil {
			return err
		}

		keyTokenStr, ok := keyToken.(string)
		if !ok {
			return unexpectedTokenError(keyToken, "string (object key)")
		}

		/*
			isNul, err := dec.nextIsNull()
			if err != nil {
				return err
			}
			if isNul {
				_, err := dec.Token()
				if err != nil {
					return err
				}
				continue
			}
		*/
		if err := callback(keyTokenStr); err != nil {
			return passUpError(keyTokenStr, err)
		}
	}
	return dec.expectDelim('}')
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

func unexpectedTokenError(got, expected interface{}) error {
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

func (dec *decoder) decodeObject(object j5reflect.PropertySet) error {
	return dec.jsonObject(func(keyTokenStr string) error {
		prop, err := object.NewValue(keyTokenStr)
		if err != nil {
			return newFieldError(keyTokenStr, "no such field")
		}
		if err := dec.decodeValue(prop); err != nil {
			return err
		}
		return nil
	})
}

func (dec *decoder) decodeValue(field j5reflect.Field) error {
	switch ft := field.(type) {
	case j5reflect.MapField:
		return dec.decodeMapField(ft)

	case j5reflect.ArrayField:
		return dec.decodeArrayField(ft)

	case j5reflect.ObjectField:
		return dec.decodeObject(ft)

	case j5reflect.OneofField:
		return dec.decodeOneof(ft)

	case j5reflect.EnumField:
		return dec.decodeEnum(ft)

	case j5reflect.ScalarField:
		return dec.decodeScalar(ft)

	case j5reflect.AnyField:
		return dec.decodeAny(ft)

	default:
		return fmt.Errorf("unknown field schema type %T", ft)
	}
}

func (dec *decoder) decodeScalar(field j5reflect.ScalarField) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}

	if _, ok := tok.(json.Delim); ok {
		return unexpectedTokenError(tok, "scalar")
	}

	return field.SetGoValue(tok)
}

func (dec *decoder) decodeEnum(field j5reflect.EnumField) error {
	token, err := dec.Token()
	if err != nil {
		return err

	}
	stringVal, ok := token.(string)
	if !ok {
		return unexpectedTokenError(token, "string")
	}

	return field.SetFromString(stringVal)
}

func (dec *decoder) decodeAny(field j5reflect.AnyField) error {
	var valueBytes []byte
	var constrainType *string

	if err := dec.jsonObject(func(keyTokenStr string) error {

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

	return field.SetJ5Any(anyVal)
}

func (dec *decoder) decodeOneof(oneof j5reflect.Oneof) error {

	foundKeys := []string{}
	var constrainType *string

	if err := dec.jsonObject(func(keyTokenStr string) error {

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

		matchedProperty, err := oneof.NewValue(keyTokenStr)
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
		_, err := oneof.NewValue(keyTokenStr)
		if err != nil {
			return newFieldError(keyTokenStr, "no such key")
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

func (dec *decoder) decodeMapField(field j5reflect.MapField) error {

	switch field := field.(type) {
	case j5reflect.MapOfScalarField:
		return dec.jsonObject(func(keyTokenStr string) error {
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
		return dec.jsonObject(func(keyTokenStr string) error {
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
		return dec.jsonObject(func(keyTokenStr string) error {
			subMsg, err := field.NewObjectElement(keyTokenStr)
			if err != nil {
				return err
			}
			return dec.decodeObject(subMsg)
		})

	case j5reflect.MapOfOneofField:
		return dec.jsonObject(func(keyTokenStr string) error {
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

func (dec *decoder) decodeArrayField(field j5reflect.ArrayField) error {

	return dec.jsonArray(func() error {
		switch field := field.(type) {
		case j5reflect.ArrayOfScalarField:
			tok, err := dec.Token()
			if err != nil {
				return err
			}

			if _, ok := tok.(json.Delim); ok {
				return unexpectedTokenError(tok, "scalar")
			}

			_, err = field.AppendGoValue(tok)
			return err

		case j5reflect.ArrayOfObjectField:
			subMsg, _ := field.NewObjectElement() // ignoring index, not error
			return dec.decodeObject(subMsg)

		case j5reflect.ArrayOfOneofField:
			subMsg, _, err := field.NewOneofElement()
			if err != nil {
				return err
			}
			return dec.decodeOneof(subMsg)

		default:
			return fmt.Errorf("unknown array schema type %T", field)
		}
	})

}
