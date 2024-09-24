package codec

import (
	"github.com/pentops/j5/lib/j5reflect"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type Codec struct {
	refl     *j5reflect.Reflector
	resolver protoregistry.MessageTypeResolver
}

func NewCodec(resolver protoregistry.MessageTypeResolver) *Codec {
	refl := j5reflect.New()
	return &Codec{
		refl:     refl,
		resolver: resolver,
	}
}

func (c *Codec) JSONToProto(jsonData []byte, msg protoreflect.Message) error {
	return c.decode(jsonData, msg)
}

func (c *Codec) ProtoToJSON(msg protoreflect.Message) ([]byte, error) {
	return c.encode(msg)
}
