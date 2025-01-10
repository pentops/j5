package j5codec

import (
	"github.com/pentops/j5/internal/codec"
)

type Codec = codec.Codec

type Option = codec.CodecOption

func NewCodec(opts ...Option) *Codec {
	cc := codec.NewCodec(opts...)
	return cc
}
