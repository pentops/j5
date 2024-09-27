package codec

import (
	"github.com/pentops/j5/j5types/any_j5t"
	"github.com/pentops/j5/lib/j5reflect"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// MessageTypeResolver is a subset of protoregistry.MessageTypeResolver
type MessageTypeResolver interface {
	FindMessageByName(message protoreflect.FullName) (protoreflect.MessageType, error)
}

type Codec struct {
	refl     *j5reflect.Reflector
	resolver MessageTypeResolver
}

type CodecOption func(*Codec)

// WithResolver sets a custom resolver for decoding any types.
func WithResolver(resolver protoregistry.MessageTypeResolver) CodecOption {
	return func(c *Codec) {
		c.resolver = resolver
	}
}

func NewCodec(...CodecOption) *Codec {
	refl := j5reflect.New()
	return &Codec{
		refl:     refl,
		resolver: protoregistry.GlobalTypes,
	}
}

func (c *Codec) JSONToProto(jsonData []byte, msg protoreflect.Message) error {
	return c.decode(jsonData, msg)
}

func (c *Codec) ProtoToJSON(msg protoreflect.Message) ([]byte, error) {
	return c.encode(msg)
}

func (c *Codec) EncodeAsEmbed(msg protoreflect.Message) (*any_j5t.Any, error) {
	jsonData, err := c.ProtoToJSON(msg)
	if err != nil {
		return nil, err
	}

	protoData, err := proto.Marshal(msg.Interface())
	if err != nil {
		return nil, err
	}

	return &any_j5t.Any{
		TypeName: string(msg.Descriptor().FullName()),
		J5Json:   jsonData,
		Proto:    protoData,
	}, nil
}
