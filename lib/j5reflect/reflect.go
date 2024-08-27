package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/internal/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Root interface {
	PropertySet
}

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

func (r *Reflector) NewRoot(msg protoreflect.Message) (Root, error) {
	if !msg.IsValid() {
		return nil, fmt.Errorf("invalid / nil message")
	}

	descriptor := msg.Descriptor()

	schema, err := r.schemaSet.Schema(descriptor)
	if err != nil {
		return nil, nil
	}

	switch schema := schema.(type) {
	case *j5schema.ObjectSchema:
		return buildObject(schema, msg)
	case *j5schema.OneofSchema:
		return buildOneof(schema, msg)
	default:
		return nil, fmt.Errorf("unsupported root schema type %T", schema)
	}
}

func (r *Reflector) NewObject(msg protoreflect.Message) (*objectImpl, error) {
	if !msg.IsValid() {
		return nil, fmt.Errorf("invalid / nil message")
	}

	descriptor := msg.Descriptor()
	schema, err := r.schemaSet.Schema(descriptor)
	if err != nil {
		return nil, err
	}

	obj, ok := schema.(*j5schema.ObjectSchema)
	if !ok {
		return nil, fmt.Errorf("expected object schema, got %T", schema)
	}

	return buildObject(obj, msg)
}

func buildObject(schema *j5schema.ObjectSchema, msg protoreflect.Message) (*objectImpl, error) {
	ps, err := newPropSet(schema, msg.Descriptor())
	if err != nil {
		return nil, err
	}
	linked := ps.linkMessage(msg)
	return &objectImpl{
		propSet: linked,
		schema:  schema,
	}, nil

}

func buildOneof(schema *j5schema.OneofSchema, msg protoreflect.Message) (*oneofImpl, error) {
	ps, err := newPropSet(schema, msg.Descriptor())
	if err != nil {
		return nil, err
	}
	linked := ps.linkMessage(msg)
	return &oneofImpl{
		propSet: linked,
		schema:  schema,
	}, nil
}
