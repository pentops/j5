package codec

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/pentops/j5/lib/j5reflect"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (c *Codec) encode(msg protoreflect.Message) ([]byte, error) {

	root, err := c.refl.NewRoot(msg)
	if err != nil {
		return nil, err
	}

	return c.encodeRoot(root)
}

func (c *Codec) encodeRoot(root j5reflect.Root) ([]byte, error) {
	enc := &encoder{
		codec: c,
		b:     &bytes.Buffer{},
	}

	switch schema := root.(type) {
	case j5reflect.Object:
		if err := enc.encodeObject(schema); err != nil {
			return nil, err
		}
	case j5reflect.Oneof:
		if err := enc.encodeOneofBody(schema); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported root schema type %T", schema)
	}
	return enc.b.Bytes(), nil
}

type encoder struct {
	b     *bytes.Buffer
	codec *Codec
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

func (enc *encoder) addQuoted(b []byte) {
	enc.add([]byte(`"`))
	enc.add(b)
	enc.add([]byte(`"`))
}

func (enc *encoder) addInt32(val int32) {
	v := strconv.FormatInt(int64(val), 10)
	enc.add([]byte(v))
}

func (enc *encoder) addUint32(val uint32) {
	v := strconv.FormatUint(uint64(val), 10)
	enc.add([]byte(v))
}

func (enc *encoder) addInt64(val int64) {
	v := strconv.FormatInt(val, 10)
	enc.addQuoted([]byte(v))
}

func (enc *encoder) addUint64(val uint64) {
	v := strconv.FormatUint(val, 10)
	enc.addQuoted([]byte(v))
}

func (enc *encoder) addBool(val bool) {
	if val {
		enc.add([]byte("true"))
	} else {
		enc.add([]byte("false"))
	}
}

func (enc *encoder) addFloat(val float64, bitSize int) {
	str := strconv.FormatFloat(val, 'g', -1, bitSize)
	enc.add([]byte(str))
}

func (enc *encoder) addNull() {
	enc.add([]byte("null"))
}
