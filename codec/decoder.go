package codec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/pentops/j5/internal/j5reflect"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (c *Codec) decode(jsonData []byte, msg protoreflect.Message) error {
	dec := json.NewDecoder(bytes.NewReader(jsonData))
	dec.UseNumber()
	d2 := &decoder{
		jd: dec,
	}

	descriptor := msg.Descriptor()

	schema, err := c.schemaSet.Schema(descriptor)
	if err != nil {
		return fmt.Errorf("schema object: %w", err)
	}

	switch schema := schema.(type) {
	case *j5reflect.ObjectSchema:
		return d2.decodeObject(schema, msg)
	case *j5reflect.OneofSchema:
		return d2.decodeOneof(schema, msg)
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

func (dec *decoder) decodeObject(object *j5reflect.ObjectSchema, msg protoreflect.Message) error {

	fieldMap := map[string]*j5reflect.ObjectProperty{}
	for _, prop := range object.ClientProperties() {
		fieldMap[prop.JSONName] = prop
	}

	return dec.jsonObject(func(keyTokenStr string) error {
		field, ok := fieldMap[keyTokenStr]
		if !ok {
			return newFieldError(keyTokenStr, "no such field")
		}

		protoFieldPath := field.ProtoField[:]
		if len(protoFieldPath) == 0 {
			// oneof annotations in proto messages are not fields, so there is
			// no 'path' from the message to the oneof.
			// Wrapped oneofs are messages, so they ARE fields, and don't get
			// handled here.
			oneofWrapper, ok := field.Schema.(*j5reflect.OneofField)
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
		}

		if err := dec.decodeField(field.Schema, settingMessage, protoField); err != nil {
			return err
		}

		return nil
	})
}

func (dec *decoder) decodeField(schema j5reflect.FieldSchema, msg protoreflect.Message, protoField protoreflect.FieldDescriptor) error {
	switch subSchema := schema.(type) {

	case *j5reflect.MapField:
		if !protoField.IsMap() {
			return errors.New("expected map")
		}

		list := msg.Mutable(protoField).Map()
		if err := dec.decodeMapField(subSchema.Schema, list); err != nil {
			return err
		}
		msg.Set(protoField, protoreflect.ValueOf(list))
		return nil

	case *j5reflect.ArrayField:
		if !protoField.IsList() {
			return errors.New("expected list")
		}

		list := msg.Mutable(protoField).List()
		if err := dec.decodeListField(subSchema.Schema, list); err != nil {
			return err
		}
		msg.Set(protoField, protoreflect.ValueOf(list))
		return nil

	case *j5reflect.ObjectField:
		if protoField.Kind() != protoreflect.MessageKind {
			return errors.New("expected message")
		}

		subMsg := msg.Mutable(protoField).Message()
		return dec.decodeObject(subSchema.Schema(), subMsg)

	case *j5reflect.OneofField:
		if protoField.Kind() != protoreflect.MessageKind {
			return errors.New("expected message for oneof")
		}
		subMsg := msg.Mutable(protoField).Message()
		return dec.decodeOneof(subSchema.Schema(), subMsg)

	case *j5reflect.EnumField:
		if protoField.Kind() != protoreflect.EnumKind {
			return errors.New("expected enum")
		}

		val, err := dec.decodeEnum(subSchema.Schema())
		if err != nil {
			return err
		}

		msg.Set(protoField, val)
		return nil

	case *j5reflect.ScalarSchema:
		if protoField.IsList() || protoField.IsMap() {
			return errors.New("expected scalar")
		}

		scalarVal, err := dec.decodeScalarField(subSchema)
		if err != nil {
			return err
		}
		msg.Set(protoField, scalarVal)
		return nil

	default:
		return fmt.Errorf("unsupported field schema type %T", subSchema)
	}
}

func (dec *decoder) decodeEnum(schema *j5reflect.EnumSchema) (protoreflect.Value, error) {
	token, err := dec.Token()
	if err != nil {
		return protoreflect.Value{}, err
	}
	stringVal, ok := token.(string)
	if !ok {
		return protoreflect.Value{}, unexpectedTokenError(token, "string")
	}

	stringVal = strings.TrimPrefix(stringVal, schema.NamePrefix)
	for _, val := range schema.Options {
		if val.Name == stringVal {
			return protoreflect.ValueOfEnum(protoreflect.EnumNumber(val.Number)), nil
		}
	}
	return protoreflect.Value{}, fmt.Errorf("enum value %s not found", stringVal)
}

func (dec *decoder) decodeOneof(oneof *j5reflect.OneofSchema, msg protoreflect.Message) error {

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

		matchedProperty := oneof.Properties.ByJSONName(keyTokenStr)
		if matchedProperty == nil {
			return errors.New("no such key")
		}
		foundKeys = append(foundKeys, keyTokenStr)

		if len(matchedProperty.ProtoField) != 1 {
			return fmt.Errorf("oneof property has proto path of %#v", matchedProperty.ProtoField)
		}

		protoFieldNumber := matchedProperty.ProtoField[0]
		protoField := msg.Descriptor().Fields().ByNumber(protoFieldNumber)
		if protoField == nil {
			return fmt.Errorf("no such field %d", protoFieldNumber)
		}
		if err := dec.decodeField(matchedProperty.Schema, msg, protoField); err != nil {
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
		matchedProperty := oneof.Properties.ByJSONName(keyTokenStr)
		if matchedProperty == nil {
			return newFieldError(keyTokenStr, "no such key")
		}
		if len(matchedProperty.ProtoField) != 1 {
			return newFieldError(keyTokenStr, "oneof has nested path")
		}
		protoFieldNumber := matchedProperty.ProtoField[0]
		protoField := msg.Descriptor().Fields().ByNumber(protoFieldNumber)
		if protoField == nil {
			return fmt.Errorf("no such field %d", protoFieldNumber)
		}
		if protoField.Kind() == protoreflect.MessageKind {
			msg.Mutable(protoField)
		} else {
			msg.Set(protoField, protoField.Default())
		}
		return nil
	}

	if len(foundKeys) > 1 {
		return newFieldError(strings.Join(foundKeys, ", "), "multiple keys found in oneof")
	}

	if constrainType != nil && foundKeys[0] != *constrainType {
		return newFieldError(foundKeys[0], fmt.Sprintf("key %q does not match type %q", foundKeys[0], *constrainType))
	}

	return nil
}

func (dec *decoder) decodeMapField(schema j5reflect.FieldSchema, list protoreflect.Map) error {
	switch subSchema := schema.(type) {

	case *j5reflect.ObjectField:
		return dec.jsonObject(func(keyTokenStr string) error {
			subMsg := list.NewValue()
			if err := dec.decodeObject(subSchema.Schema(), subMsg.Message()); err != nil {
				return err
			}

			list.Set(protoreflect.ValueOfString(keyTokenStr).MapKey(), subMsg)
			return nil
		})

	case *j5reflect.OneofField:
		return dec.jsonObject(func(keyTokenStr string) error {
			subMsg := list.NewValue()
			if err := dec.decodeOneof(subSchema.Schema(), subMsg.Message()); err != nil {
				return err
			}

			list.Set(protoreflect.ValueOfString(keyTokenStr).MapKey(), subMsg)
			return nil
		})

	case *j5reflect.EnumField:
		return dec.jsonObject(func(keyTokenStr string) error {
			value, err := dec.decodeEnum(subSchema.Schema())
			if err != nil {
				return err
			}

			list.Set(protoreflect.ValueOfString(keyTokenStr).MapKey(), value)
			return nil
		})

	case *j5reflect.ScalarSchema:
		return dec.jsonObject(func(keyTokenStr string) error {
			value, err := dec.decodeScalarField(subSchema)
			if err != nil {
				return err
			}
			list.Set(protoreflect.ValueOfString(keyTokenStr).MapKey(), value)
			return nil
		})
	default:
		return fmt.Errorf("unsupported map schema type %T", subSchema)

	}
}

func (dec *decoder) decodeListField(schema j5reflect.FieldSchema, list protoreflect.List) error {

	switch subSchema := schema.(type) {

	case *j5reflect.ObjectField:
		return dec.jsonArray(func() error {
			subMsg := list.NewElement()
			if err := dec.decodeObject(subSchema.Schema(), subMsg.Message()); err != nil {
				return err
			}

			list.Append(subMsg)
			return nil
		})

	case *j5reflect.OneofField:
		return dec.jsonArray(func() error {
			subMsg := list.NewElement()
			if err := dec.decodeOneof(subSchema.Schema(), subMsg.Message()); err != nil {
				return err
			}

			list.Append(subMsg)
			return nil
		})

	case *j5reflect.EnumField:

		return dec.jsonArray(func() error {
			value, err := dec.decodeEnum(subSchema.Schema())
			if err != nil {
				return err
			}

			list.Append(value)
			return nil
		})

	case *j5reflect.ScalarSchema:
		return dec.jsonArray(func() error {
			value, err := dec.decodeScalarField(subSchema)
			if err != nil {
				return err
			}
			list.Append(value)
			return nil
		})
	default:
		return fmt.Errorf("unsupported array schema type %T", subSchema)

	}
}
