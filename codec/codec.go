package codec

import (
	"github.com/pentops/j5/internal/structure"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Codec struct {
	schemaSet *structure.SchemaSet
}

func NewCodec() *Codec {
	return &Codec{
		schemaSet: structure.NewSchemaSet(),
	}
}

func (c *Codec) JSONToProto(jsonData []byte, msg protoreflect.Message) error {
	return c.decode(jsonData, msg)
}

func (c *Codec) ProtoToJSON(msg protoreflect.Message) ([]byte, error) {
	return c.encode(msg)
}
