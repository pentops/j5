package codec

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/pentops/sugar-go/v1/sugar_pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func Decode(opts Options, jsonData []byte, msg protoreflect.Message) error {
	dec := json.NewDecoder(bytes.NewReader(jsonData))
	dec.UseNumber()
	d2 := &decoder{
		Decoder: dec,
		Options: opts,
	}
	return d2.decodeMessage(msg)
}

type decoder struct {
	*json.Decoder
	next json.Token
	Options
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

func (dec *decoder) decodeMessage(msg protoreflect.Message) error {
	wktDecoder := wellKnownTypeUnmarshaler(msg.Descriptor().FullName())
	if wktDecoder != nil {
		return wktDecoder(dec, msg)
	}

	if err := dec.startObject(); err != nil {
		return err
	}

	descriptor := msg.Descriptor()
	fields := descriptor.Fields()
	oneofs := descriptor.Oneofs()

	isOneofWrapper := false

	msgOptions := msg.Descriptor().Options()
	ext := proto.GetExtension(msgOptions, sugar_pb.E_Message).(*sugar_pb.Message)
	if ext != nil {
		isOneofWrapper = ext.OneofWrapper
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
			return fmt.Errorf("expected string key but got %v", keyToken)
		}

		protoField := fields.ByJSONName(keyTokenStr)
		if protoField == nil {
			if !dec.Options.WrapOneof {
				return fmt.Errorf("no such field %s", keyTokenStr)
			}
			keyTokenStr = jsonNameToProto(keyTokenStr)
			oneof := oneofs.ByName(protoreflect.Name(keyTokenStr))
			if oneof == nil {
				return fmt.Errorf("no such field %s", keyTokenStr)
			}

			if err := dec.decodeOneofField(msg, oneof); err != nil {
				return fmt.Errorf("decoding '%s': %w", keyTokenStr, err)
			}
			continue
		}
		if !isOneofWrapper && dec.Options.WrapOneof && protoField.ContainingOneof() != nil {
			return fmt.Errorf("field %s is part of a oneof", keyTokenStr)
		}

		if protoField.IsMap() {
			if err := dec.decodeMapField(msg, protoField); err != nil {
				return fmt.Errorf("decoding '%s': %w", keyTokenStr, err)
			}
		} else if protoField.IsList() {
			if err := dec.decodeListField(msg, protoField); err != nil {
				return fmt.Errorf("decoding '%s[]': %w", keyTokenStr, err)
			}
		} else {
			if err := dec.decodeField(msg, protoField); err != nil {
				return fmt.Errorf("decoding '%s': %w", keyTokenStr, err)
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
		return fmt.Errorf("no such oneof type %s", oneofKeyTokenStr)
	}

	if err := dec.decodeField(msg, oneofField); err != nil {
		return fmt.Errorf("decoding oneof child '%s': %w", oneofKeyTokenStr, err)
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
				return err
			}
			list.Append(subMsg)

		default:

			value, err := dec.decodeScalarField(field)
			if err != nil {
				return err
			}
			list.Append(value)
		}

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
