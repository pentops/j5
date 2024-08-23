package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/internal/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Root interface{}

type Reflector struct {
	schemaSet *j5schema.SchemaCache
}

func New() *Reflector {
	return &Reflector{
		schemaSet: j5schema.NewSchemaCache(),
	}
}

func NewWithCache(cache *j5schema.SchemaCache) *Reflector {
	return &Reflector{
		schemaSet: cache,
	}
}

func (r *Reflector) NewRoot(protoMsg protoreflect.Message) (Root, error) {

	descriptor := protoMsg.Descriptor()

	schema, err := r.schemaSet.Schema(descriptor)
	if err != nil {
		return nil, nil
	}

	msg, err := newRootMessageValue(protoMsg)
	if err != nil {
		return nil, err
	}

	switch schema := schema.(type) {
	case *j5schema.ObjectSchema:
		return newObject(schema, msg)
	case *j5schema.OneofSchema:
		return newOneof(schema, msg)
	default:
		return nil, fmt.Errorf("unsupported root schema type %T", schema)
	}
}

func (r *Reflector) NewObject(msg protoreflect.Message) (*ObjectImpl, error) {
	descriptor := msg.Descriptor()
	schema, err := r.schemaSet.Schema(descriptor)
	if err != nil {
		return nil, err
	}

	obj, ok := schema.(*j5schema.ObjectSchema)
	if !ok {
		return nil, fmt.Errorf("expected object schema, got %T", schema)
	}

	mv, err := newRootMessageValue(msg)
	if err != nil {
		return nil, err
	}

	return newObject(obj, mv)
}
