package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/internal/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type PropertyParent interface{}
type Root interface{}

type Object struct {
	schema  *j5schema.ObjectSchema
	message protoreflect.Message
	*fieldset
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

	descriptor := msg.Descriptor()

	schema, err := r.schemaSet.Schema(descriptor)
	if err != nil {
		return nil, nil
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

func (r *Reflector) NewObject(msg protoreflect.Message) (*Object, error) {

	descriptor := msg.Descriptor()
	schema, err := r.schemaSet.Schema(descriptor)
	if err != nil {
		return nil, err
	}

	obj, ok := schema.(*j5schema.ObjectSchema)
	if !ok {
		return nil, fmt.Errorf("expected object schema, got %T", schema)
	}

	return newObject(obj, msg)
}

func newObject(schema *j5schema.ObjectSchema, msg protoreflect.Message) (*Object, error) {

	props, err := collectProperties(schema.ClientProperties(), msg)
	if err != nil {
		return nil, err
	}

	fieldset, err := newFieldset(props)
	if err != nil {
		return nil, err
	}

	return &Object{
		schema:   schema,
		message:  msg,
		fieldset: fieldset,
	}, nil
}

type Oneof struct {
	schema *j5schema.OneofSchema
	msg    protoreflect.Message
	*fieldset
}

func newOneof(schema *j5schema.OneofSchema, msg protoreflect.Message) (*Oneof, error) {

	props, err := collectProperties(schema.Properties, msg)
	if err != nil {
		return nil, err
	}

	fieldset, err := newFieldset(props)
	if err != nil {
		return nil, err
	}

	return &Oneof{
		schema:   schema,
		msg:      msg,
		fieldset: fieldset,
	}, nil
}
