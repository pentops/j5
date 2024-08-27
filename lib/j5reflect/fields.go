package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/internal/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type FieldType string

const (
	FieldTypeUnknown = FieldType("?")
	FieldTypeObject  = FieldType("object")
	FieldTypeOneof   = FieldType("oneof")
	FieldTypeEnum    = FieldType("enum")
	FieldTypeArray   = FieldType("array")
	FieldTypeMap     = FieldType("map")
	FieldTypeScalar  = FieldType("scalar")
)

type fieldContext interface {
	nameInParent() string
	indexInParent() int

	fieldSchema() schema_j5pb.IsField_Type
	propertySchema() *schema_j5pb.ObjectProperty
	typeName() string
	fullTypeName() string
	protoPath() []string
}

type propertyContext struct {
	walkedProtoPath []string
	schema          *j5schema.ObjectProperty
}

func (c propertyContext) nameInParent() string {
	return c.schema.JSONName
}

func (c propertyContext) indexInParent() int {
	return -1
}

func (c propertyContext) propertySchema() *schema_j5pb.ObjectProperty {
	return c.schema.ToJ5Proto()
}

func (c propertyContext) fieldSchema() schema_j5pb.IsField_Type {
	return c.schema.Schema.ToJ5Field().Type
}

func (c propertyContext) typeName() string {
	return c.schema.Schema.TypeName()
}

func (c propertyContext) fullTypeName() string {
	return c.schema.FullName()
}

func (c propertyContext) protoPath() []string {
	return c.walkedProtoPath
}

type fieldDefaults struct {
	fieldType FieldType
	context   fieldContext
}

func (fd fieldDefaults) Type() FieldType {
	return fd.fieldType
}

func (fieldDefaults) AsContainer() (ContainerField, bool) {
	return nil, false
}

func (fieldDefaults) AsScalar() (ScalarField, bool) {
	return nil, false
}

func (fieldDefaults) AsArrayOfContainer() (ArrayOfContainerField, bool) {
	return nil, false
}

func (fieldDefaults) AsArrayOfScalar() (ArrayOfScalarField, bool) {
	return nil, false
}

func (f fieldDefaults) NameInParent() string {
	return f.context.nameInParent()
}

func (f fieldDefaults) IndexInParent() int {
	return f.context.indexInParent()
}

func (f fieldDefaults) PropertySchema() *schema_j5pb.ObjectProperty {
	return f.context.propertySchema()
}

func (f fieldDefaults) Schema() schema_j5pb.IsField_Type {
	return f.context.fieldSchema()
}

func (f fieldDefaults) ProtoPath() []string {
	return f.context.protoPath()
}

func (fd fieldDefaults) TypeName() string {
	return fd.context.typeName()
}

func (fd fieldDefaults) FullTypeName() string {
	return fd.context.fullTypeName()
}

type Field interface {
	Type() FieldType
	TypeName() string
	IsSet() bool

	// NameInParent is the name this field has in the context it exists.
	// Object and Oneof: The property name
	// Map<string>x, the key
	// Arrays, the index as a string
	NameInParent() string

	// IndexInParent returns -1 for non array fields
	IndexInParent() int
	ProtoPath() []string
	FullTypeName() string

	Schema() schema_j5pb.IsField_Type
	PropertySchema() *schema_j5pb.ObjectProperty

	// Fighting with go typing here, the implementations of these return
	// themselves and true.
	AsScalar() (ScalarField, bool)
	AsContainer() (ContainerField, bool)
	AsArrayOfContainer() (ArrayOfContainerField, bool)
	AsArrayOfScalar() (ArrayOfScalarField, bool)
}

type ContainerField interface {
	Field
	GetOrCreateContainer() (PropertySet, error)
	GetExistingContainer() (PropertySet, bool, error)
}

type ArrayOfContainerField interface {
	MutableArrayField
	NewContainerElement() (ContainerField, int, error)
	RangeContainers(func(ContainerField, PropertySet) error) error
}

type fieldFactory interface {
	buildField(schema fieldContext, value protoContext) Field
}

func newFieldFactory(schema j5schema.FieldSchema, field protoreflect.FieldDescriptor) (fieldFactory, error) {
	switch st := schema.(type) {
	case *j5schema.ObjectField:
		if field.Kind() != protoreflect.MessageKind {
			return nil, fmt.Errorf("ObjectField is kind %s", field.Kind())
		}
		return &objectFieldFactory{schema: st}, nil

	case *j5schema.OneofField:
		if field.Kind() != protoreflect.MessageKind {
			return nil, fmt.Errorf("OneofField is kind %s", field.Kind())
		}
		return &oneofFieldFactory{schema: st}, nil

	case *j5schema.EnumField:
		if field.Kind() != protoreflect.EnumKind {
			return nil, fmt.Errorf("EnumField is kind %s", field.Kind())
		}
		return &enumFieldFactory{schema: st}, nil

	case *j5schema.ScalarSchema:
		if st.WellKnownTypeName != "" {
			if field.Kind() != protoreflect.MessageKind {
				return nil, fmt.Errorf("ScalarField is proto kind %s, want message for %T", field.Kind(), st.Proto.Type)
			}
			if string(field.Message().FullName()) != string(st.WellKnownTypeName) {
				return nil, fmt.Errorf("ScalarField message is %s, want %s for %T", field.Message().FullName(), st.WellKnownTypeName, st.Proto.Type)
			}
		} else if field.Kind() != st.Kind {
			return nil, fmt.Errorf("ScalarField is proto kind %s, want schema %q for %T", field.Kind(), st.Kind, st.Proto.Type)
		}
		return &scalarFieldFactory{schema: st}, nil

	default:
		return nil, fmt.Errorf("unsupported schema type %T", schema)
	}
}
