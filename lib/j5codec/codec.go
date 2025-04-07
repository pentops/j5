package j5codec

import (
	"github.com/pentops/j5/internal/codec"
)

var GlobalCodec = codec.GlobalCodec

type Codec = codec.Codec

type Option = codec.CodecOption

func NewCodec(opts ...Option) *Codec {
	cc := codec.NewCodec(opts...)
	return cc
}

type MessageTypeResolver = codec.MessageTypeResolver
type CodecOption = codec.CodecOption

// WithResolver sets a custom resolver for decoding any types.
func WithResolver(resolver MessageTypeResolver) CodecOption {
	return codec.WithResolver(resolver)
}

// WithProtoToAny adds a proto encoding to j5.any.v1 messages using the resolver
func WithProtoToAny() CodecOption {
	return codec.WithProtoToAny()
}
