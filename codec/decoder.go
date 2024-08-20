package codec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/pentops/j5/lib/j5reflect"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (c *Codec) decode(jsonData []byte, msg protoreflect.Message) error {
	dec := json.NewDecoder(bytes.NewReader(jsonData))
	dec.UseNumber()
	d2 := &decoder{
		jd: dec,
	}

	root, err := j5reflect.NewWithCache(c.schemaSet).NewRoot(msg)
	if err != nil {
		return err
	}

	switch schema := root.(type) {
	case *j5reflect.ObjectImpl:
		return d2.decodeObject(schema)
	case *j5reflect.OneofImpl:
		return d2.decodeOneof(schema)
	default:
		return fmt.Errorf("unsupported root schema type %T", schema)
	}
}

type decoder struct {
	jd   *json.Decoder
	next json.Token
}

func (d *decoder) Token() (json.Token, error) {
	if d.next != nil {
		tok := d.next
		d.next = nil
		return tok, nil
	}

	return d.jd.Token()
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

		if err := callback(keyTokenStr); err != nil {
			return passUpError(keyTokenStr, err)
		}
	}
	return dec.expectDelim('}')
}
func (d *decoder) unmarshalEmptyObject() error {
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

func (d *decoder) stringToken() (string, error) {
	tok, err := d.Token()
	if err != nil {
		return "", err
	}

	stringVal, ok := tok.(string)
	if !ok {
		return "", fmt.Errorf("expected string but got %v", tok)
	}
	return stringVal, nil
}

type fieldError struct {
	pathToField []string
	err         error
}

func newFieldError(field, message string) error {
	return fieldError{
		pathToField: []string{field},
		err:         fmt.Errorf(message),
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

func (dec *decoder) decodeObject(object j5reflect.Object) error {

	return dec.jsonObject(func(keyTokenStr string) error {
		prop := object.MaybeGetProperty(keyTokenStr)
		if prop == nil {
			return newFieldError(keyTokenStr, "no such field")
		}

		/*
			protoFieldPath := field.ProtoField[:]
			if len(protoFieldPath) == 0 {
				// oneof annotations in proto messages are not fields, so there is
				// no 'path' from the message to the oneof.
				// Wrapped oneofs are messages, so they ARE fields, and don't get
				// handled here.
				oneofWrapper, ok := field.Schema.(*j5schema.OneofField)
				if !ok {
					return newFieldError(keyTokenStr, "field has no proto field and is not a oneof")
				}

				if err := dec.decodeOneof(oneofWrapper.Schema(), msg); err != nil {
					return err
				}
				return nil
			}

			var protoField protoreflect.FieldDescriptor
			var protoFieldNumber protoreflect.FieldNumber
			settingMessage := msg
			for {
				protoFieldNumber, protoFieldPath = protoFieldPath[0], protoFieldPath[1:]
				protoField = settingMessage.Descriptor().Fields().ByNumber(protoFieldNumber)
				if protoField == nil {
					return fmt.Errorf("no such field %d in msg %s", protoFieldNumber, settingMessage.Descriptor().FullName())
				}
				if len(protoFieldPath) == 0 {
					break
				}

				// The field should me a message, there are remaining fields in the
				// path
				if protoField.Kind() != protoreflect.MessageKind {
					return newFieldError(protoField.JSONName(), "field is not a message but has a message annotation")
				}

				// if the field is nil, create a new message
				if !settingMessage.Has(protoField) {
					subMsg := settingMessage.Mutable(protoField).Message()
					settingMessage = subMsg
				} else {
					settingMessage = settingMessage.Get(protoField).Message()
				}
			}*/

		if err := dec.decodeValue(prop.Field()); err != nil {
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
		object, err := ft.Object()
		if err != nil {
			return err
		}
		return dec.decodeObject(object)

	case j5reflect.OneofField:
		field, err := ft.Oneof()
		if err != nil {
			return err
		}
		return dec.decodeOneof(field)

	case j5reflect.EnumField:
		return dec.decodeEnum(ft)

	case j5reflect.ScalarField:
		tok, err := dec.Token()
		if err != nil {
			return err
		}

		if _, ok := tok.(json.Delim); ok {
			return unexpectedTokenError(tok, "scalar")
		}

		return ft.SetGoValue(tok)

	default:
		return fmt.Errorf("unknown field schema type %T", ft)
	}
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

		matchedProperty := oneof.MaybeGetProperty(keyTokenStr)
		if matchedProperty == nil {
			return errors.New("no such key")
		}
		foundKeys = append(foundKeys, keyTokenStr)

		if err := dec.decodeValue(matchedProperty.Field()); err != nil {
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
		matchedProperty := oneof.MaybeGetProperty(keyTokenStr)
		if matchedProperty == nil {
			return newFieldError(keyTokenStr, "no such key")
		}
		return matchedProperty.Field().SetDefault()
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

			return field.SetGoScalar(keyTokenStr, tok)
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
			subMsg, err := field.NewObjectValue(keyTokenStr)
			if err != nil {
				return err
			}
			return dec.decodeObject(subMsg)
		})

	case j5reflect.MapOfOneofField:
		return dec.jsonObject(func(keyTokenStr string) error {
			subMsg, err := field.NewOneofValue(keyTokenStr)
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

			return field.AppendGoScalar(tok)

		case j5reflect.ArrayOfEnumField:
			tok, err := dec.Token()
			if err != nil {
				return err
			}

			str, ok := tok.(string)
			if !ok {
				return unexpectedTokenError(tok, "string")
			}

			return field.AppendEnumFromString(str)

		case j5reflect.ArrayOfObjectField:
			subMsg, err := field.NewObjectElement()
			if err != nil {
				return err
			}
			return dec.decodeObject(subMsg)

		case j5reflect.ArrayOfOneofField:
			subMsg, err := field.NewOneofElement()
			if err != nil {
				return err
			}
			return dec.decodeOneof(subMsg)

		default:
			return fmt.Errorf("unknown array schema type %T", field)
		}
	})

}
