package codec

import (
	"github.com/pentops/jsonapi/gen/j5/config/v1/config_j5pb"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Codec struct {
	Options
}

func NewCodec(optsSrc *config_j5pb.CodecOptions) *Codec {
	opts := Options{}

	if optsSrc.WrapOneof {
		opts.WrapOneof = true
	}

	if optsSrc.ShortEnums != nil {
		opts.ShortEnums = &ShortEnumsOption{
			UnspecifiedSuffix: optsSrc.ShortEnums.UnspecifiedSuffix,
			StrictUnmarshal:   optsSrc.ShortEnums.StrictUnmarshal,
		}
	}

	return &Codec{
		Options: opts,
	}
}

func (c *Codec) ToProto(jsonData []byte, msg protoreflect.Message) error {
	return Decode(c.Options, jsonData, msg)
}

func (c *Codec) FromProto(msg protoreflect.Message) ([]byte, error) {
	return Encode(c.Options, msg)
}
