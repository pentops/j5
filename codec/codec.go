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

	addProtoToAny bool
}

type CodecOption func(*Codec)

// WithResolver sets a custom resolver for decoding any types.
func WithResolver(resolver MessageTypeResolver) CodecOption {
	return func(c *Codec) {
		c.resolver = resolver
	}
}

// WithProtoToAny adds a proto encoding to j5.any.v1 messages using the resolver
func WithProtoToAny() CodecOption {
	return func(c *Codec) {
		c.addProtoToAny = true
	}
}

func NewCodec(opts ...CodecOption) *Codec {
	refl := j5reflect.New()
	cc := &Codec{
		refl:     refl,
		resolver: protoregistry.GlobalTypes,
	}

	for _, opt := range opts {
		opt(cc)
	}

	return cc
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
