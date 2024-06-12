package codec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func Decode(jsonData []byte, msg protoreflect.Message) error {
	dec := json.NewDecoder(bytes.NewReader(jsonData))
	dec.UseNumber()
	d2 := &decoder{
		Decoder: dec,
	}
	return d2.decodeMessage(msg)
}

type decoder struct {
	*json.Decoder
	next json.Token
}

func (d *decoder) Peek() (json.Token, error) {
	if d.next != nil {
		return nil, fmt.Errorf("unexpected call to Peek after Peek")
	}

	tok, err := d.Token()
	if err != nil {
		return nil, err
	}

	d.next = tok
	return tok, nil
}

func (d *decoder) Token() (json.Token, error) {
	if d.next != nil {
		tok := d.next
		d.next = nil
		return tok, nil
	}

	return d.Decoder.Token()
}

type mappedMessage struct {
	protoFields map[string][]protoreflect.FieldDescriptor
}

// Returns a map of json field names to a list of field descriptors that
// represent the path from the root message to the field. For most fields this
// will be one field descriptor, but for some the alias may be a path to a
// nested message field
func mapMessageFields(msg protoreflect.MessageDescriptor) (*mappedMessage, error) {

	// TODO: Cache the result of this function
	// TODO: Include oneof maps here

	fields := msg.Fields()

	flatFields := make(map[string][]protoreflect.FieldDescriptor)

	for idx := 0; idx < fields.Len(); idx++ {
		field := fields.Get(idx)
		fieldOptions := proto.GetExtension(field.Options(), ext_j5pb.E_Field).(*ext_j5pb.FieldOptions)
		if fieldOptions != nil {
			switch option := fieldOptions.Type.(type) {
			case *ext_j5pb.FieldOptions_Message:
				if field.Kind() != protoreflect.MessageKind {
					return nil, fmt.Errorf("field %s is not a message but has a message annotation", field.FullName())
				}

				if option.Message.Flatten {
					subFields := field.Message().Fields()
					for idx := 0; idx < subFields.Len(); idx++ {
						subField := subFields.Get(idx)
						flatFields[subField.JSONName()] = []protoreflect.FieldDescriptor{field, subField}

						if subField.Kind() == protoreflect.MessageKind {
							subFieldFlatFields, err := mapMessageFields(subField.Message())
							if err != nil {
								return nil, fmt.Errorf("mapping subfield %s: %w", subField.FullName(), err)
							}
							for subFieldName, subFieldFlatField := range subFieldFlatFields.protoFields {
								flatFields[subFieldName] = append([]protoreflect.FieldDescriptor{field}, subFieldFlatField...)
							}
						}
					}
					continue
				}

			}
		}
		flatFields[field.JSONName()] = []protoreflect.FieldDescriptor{field}
	}

	return &mappedMessage{
		protoFields: flatFields,
	}, nil
}

type fieldError struct {
	pathToField []string
	err         error
}

func newFieldError(field, message string, args ...interface{}) error {
	return fieldError{
		pathToField: []string{field},
		err:         fmt.Errorf(message, args...),
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

func (dec *decoder) decodeEnum(enum protoreflect.EnumDescriptor, stringVal string) (protoreflect.EnumNumber, error) {

	vals := enum.Values()

	enumVal := vals.ByName(protoreflect.Name(stringVal))
	if enumVal != nil {
		return enumVal.Number(), nil
	}

	prefix := enumPrefix(enum)
	if prefix == "" {
		return 0, fmt.Errorf("unknown enum value %s", stringVal)
	}
	fullName := fmt.Sprintf("%s%s", prefix, stringVal)

	enumVal = vals.ByName(protoreflect.Name(fullName))
	if enumVal == nil {
		return 0, fmt.Errorf("unknown enum value %s (%s)", stringVal, fullName)
	}
	return enumVal.Number(), nil
}

func (dec *decoder) decodeMessage(msg protoreflect.Message) error {
	wktDecoder := wellKnownTypeUnmarshaler(msg.Descriptor().FullName())
	if wktDecoder != nil {
		return wktDecoder(dec, msg)
	}

	if err := dec.startObject(); err != nil {
		return err
	}

	descriptor := msg.Descriptor()
	oneofs := descriptor.Oneofs()

	isOneofWrapper := false

	msgOptions := msg.Descriptor().Options()
	ext := proto.GetExtension(msgOptions, ext_j5pb.E_Message).(*ext_j5pb.MessageOptions)
	if ext != nil {
		isOneofWrapper = ext.IsOneofWrapper
	}

	fieldAliases, err := mapMessageFields(msg.Descriptor())
	if err != nil {
		return fmt.Errorf("mapping fields: %w", err)
	}

	for {
		if !dec.More() {
			break
		}

		keyToken, err := dec.Token()
		if err != nil {
			return err
		}

		// Otherwise should be a key
		keyTokenStr, ok := keyToken.(string)
		if !ok {
			return unexpectedTokenError(keyToken, "string (object key)")
		}

		protoFieldPath, ok := fieldAliases.protoFields[keyTokenStr]
		if !ok {
			keyTokenStr = jsonNameToProto(keyTokenStr)
			oneof := oneofs.ByName(protoreflect.Name(keyTokenStr))
			if oneof == nil {
				return newFieldError(keyTokenStr, "no such field")
			}

			ext := proto.GetExtension(oneof.Options(), ext_j5pb.E_Oneof).(*ext_j5pb.OneofOptions)
			if ext == nil || !ext.Expose {
				return newFieldError(keyTokenStr, "no such field")
			}

			if err := dec.decodeOneofField(msg, oneof); err != nil {
				return passUpError(keyTokenStr, err)
			}
			continue
		}

		var protoField protoreflect.FieldDescriptor
		settingMessage := msg
		for {
			protoField = protoFieldPath[0]
			if len(protoFieldPath) == 1 {
				break
			}
			protoFieldPath = protoFieldPath[1:]

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

		// We may find a field which is supposed to be in an exposed oneof
		if !isOneofWrapper && protoField.ContainingOneof() != nil {
			containingOneof := protoField.ContainingOneof()
			if !containingOneof.IsSynthetic() {
				ext := proto.GetExtension(containingOneof.Options(), ext_j5pb.E_Oneof).(*ext_j5pb.OneofOptions)
				if ext != nil && ext.Expose {
					jsonName := protoNameToJSON(containingOneof.Name())
					return newFieldError(keyTokenStr, "no such field (but %s.%s is)", jsonName, keyTokenStr)
				}
			}
		}

		if protoField.IsMap() {
			if err := dec.decodeMapField(settingMessage, protoField); err != nil {
				return passUpError(keyTokenStr, err)
			}
		} else if protoField.IsList() {
			if err := dec.decodeListField(settingMessage, protoField); err != nil {
				return passUpError(keyTokenStr, err)
			}
		} else {
			if err := dec.decodeField(settingMessage, protoField); err != nil {
				return passUpError(keyTokenStr, err)
			}
		}
	}

	return dec.endObject()
}

func (dec *decoder) decodeOneofField(msg protoreflect.Message, oneof protoreflect.OneofDescriptor) error {

	if err := dec.startObject(); err != nil {
		return err
	}

	oneofKeyToken, err := dec.Token()
	if err != nil {
		return err
	}

	oneofKeyTokenStr, ok := oneofKeyToken.(string)
	if !ok {
		return unexpectedTokenError(oneofKeyToken, "string (oneof key)")
	}

	oneofField := oneof.Fields().ByJSONName(oneofKeyTokenStr)
	if oneofField == nil {
		return newFieldError(oneofKeyTokenStr, "no such key")
	}

	if err := dec.decodeField(msg, oneofField); err != nil {
		return passUpError(oneofKeyTokenStr, err)
	}

	if err := dec.endObject(); err != nil {
		return err
	}

	return nil
}

func (dec *decoder) startObject() error {
	return dec.expectDelim('{')
}
func (dec *decoder) endObject() error {
	return dec.expectDelim('}')
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

func (dec *decoder) decodeField(msg protoreflect.Message, field protoreflect.FieldDescriptor) error {
	switch field.Kind() {
	case protoreflect.MessageKind:
		return dec.decodeMessageField(msg, field)

	default:
		scalarVal, err := dec.decodeScalarField(field)
		if err != nil {
			return err
		}
		msg.Set(field, scalarVal)
	}
	return nil
}

func (dec *decoder) decodeMapField(msg protoreflect.Message, field protoreflect.FieldDescriptor) error {
	token, err := dec.Token()
	if err != nil {
		return err
	}

	if token != json.Delim('{') {
		return unexpectedTokenError(token, "{")
	}

	mapValue := field.MapValue()
	mapValueKind := mapValue.Kind()

	list := msg.Mutable(field).Map()

	for {
		if !dec.More() {
			_, err := dec.Token()
			if err != nil {
				return err
			}
			break
		}

		keyValue, err := dec.decodeScalarField(field.MapKey())
		if err != nil {
			return err
		}

		switch mapValueKind {
		case protoreflect.MessageKind:
			subMsg := list.NewValue()
			if err := dec.decodeMessage(subMsg.Message()); err != nil {
				return err
			}
			list.Set(keyValue.MapKey(), subMsg)

		default:
			value, err := dec.decodeScalarField(mapValue)
			if err != nil {
				return err
			}
			list.Set(keyValue.MapKey(), value)
		}

	}

	msg.Set(field, protoreflect.ValueOf(list))
	return nil

}

func (dec *decoder) decodeListField(msg protoreflect.Message, field protoreflect.FieldDescriptor) error {

	tok, err := dec.Token()
	if err != nil {
		return err
	}

	if tok != json.Delim('[') {
		return fmt.Errorf("expected '[' but got %v", tok)
	}

	kind := field.Kind()
	list := msg.Mutable(field).List()

	idx := 0
	for {
		if !dec.More() {
			_, err := dec.Token()
			if err != nil {
				return err
			}
			break
		}

		switch kind {
		case protoreflect.MessageKind:
			subMsg := list.NewElement()
			if err := dec.decodeMessage(subMsg.Message()); err != nil {
				return passUpError(fmt.Sprint(idx), err)
			}
			list.Append(subMsg)

		default:

			value, err := dec.decodeScalarField(field)
			if err != nil {
				return err
			}
			list.Append(value)
		}
		idx++

	}

	msg.Set(field, protoreflect.ValueOf(list))
	return nil
}

func (dec *decoder) decodeMessageField(msg protoreflect.Message, field protoreflect.FieldDescriptor) error {

	subMsg := msg.Mutable(field).Message()
	if err := dec.decodeMessage(subMsg); err != nil {
		return err
	}

	return nil
}
