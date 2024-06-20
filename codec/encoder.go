package codec

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/pentops/j5/schema/j5reflect"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (c *Codec) encode(msg protoreflect.Message) ([]byte, error) {
	enc := &encoder{
		b: &bytes.Buffer{},
	}

	descriptor := msg.Descriptor()

	schema, err := c.schemaSet.SchemaReflect(descriptor)
	if err != nil {
		return nil, fmt.Errorf("schema object: %w", err)
	}

	switch schema := schema.Type().(type) {
	case *j5reflect.ObjectSchema:
		if err := enc.encodeObject(schema, msg); err != nil {
			return nil, err
		}
	case *j5reflect.OneofSchema:
		if err := enc.encodeOneof(schema, msg); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported schema type %T", schema)
	}
	return enc.b.Bytes(), nil
}

type encoder struct {
	b *bytes.Buffer
}

func (enc *encoder) add(b []byte) {
	enc.b.Write(b)
}

func (enc *encoder) openObject() {
	enc.add([]byte("{"))
}

func (enc *encoder) closeObject() {
	enc.add([]byte("}"))
}

func (enc *encoder) openArray() {
	enc.add([]byte("["))
}

func (enc *encoder) closeArray() {
	enc.add([]byte("]"))
}

func (enc *encoder) fieldSep() {
	enc.add([]byte(","))
}

func (enc *encoder) fieldLabel(label string) error {
	if err := enc.addString(label); err != nil {
		return err
	}
	enc.add([]byte(":"))
	return nil
}

func (enc *encoder) addString(unclean string) error {
	buffer := make([]byte, 0, len(unclean)+2)
	var err error
	buffer, err = appendString(buffer, unclean)
	if err != nil {
		return err
	}
	enc.add(buffer)
	return nil
}

func (enc *encoder) addFloat(val float64, bitSize int) {
	str := strconv.FormatFloat(val, 'g', -1, bitSize)
	enc.add([]byte(str))
}

func (enc *encoder) addInt(val int64) {
	v := strconv.FormatInt(val, 10)
	enc.add([]byte(v))
}

func (enc *encoder) addUint(val uint64) {
	v := strconv.FormatUint(val, 10)
	enc.add([]byte(v))
}

func (enc *encoder) addBool(val bool) {
	if val {
		enc.add([]byte("true"))
	} else {
		enc.add([]byte("false"))
	}
}

/*
func (enc *encoder) addNull() {
	enc.add([]byte("null"))
}*/
