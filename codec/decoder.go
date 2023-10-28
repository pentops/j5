package codec

import (
	"bytes"
	"encoding/json"
	"fmt"

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

	tok, err := dec.Token()
	if err != nil {
		return err
	}

	if tok != json.Delim('{') {
		return fmt.Errorf("expected '{' but got %v", tok)
	}

	for {
		keyToken, err := dec.Token()
		if err != nil {
			return err
		}

		// Ends the object
		if keyToken == json.Delim('}') {
			return nil
		}

		// Otherwise should be a key
		keyTokenStr, ok := keyToken.(string)
		if !ok {
			return fmt.Errorf("expected string key but got %v", keyToken)
		}

		protoField := msg.Descriptor().Fields().ByJSONName(keyTokenStr)
		if protoField == nil {
			return fmt.Errorf("no such field %s", keyTokenStr)
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
