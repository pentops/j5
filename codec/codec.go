package codec

import (
	"github.com/pentops/j5/lib/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Codec struct {
	schemaSet *j5schema.SchemaCache
}

func NewCodec() *Codec {
	return &Codec{
		schemaSet: j5schema.NewSchemaCache(),
	}
}

func (c *Codec) JSONToProto(jsonData []byte, msg protoreflect.Message) error {
	return c.decode(jsonData, msg)
}

func (c *Codec) ProtoToJSON(msg protoreflect.Message) ([]byte, error) {
	return c.encode(msg)
}
