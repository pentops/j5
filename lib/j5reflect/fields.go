package j5reflect

import (
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/internal/j5schema"
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
	// themselves and true. Maybe I should have fixated on java instead.
	AsArray() (ArrayField, bool)
	AsScalar() (ScalarField, bool)
	AsContainer() (ContainerField, bool)
	AsArrayOfContainer() (ArrayOfContainerField, bool)
	AsArrayOfScalar() (ArrayOfScalarField, bool)
}

type fieldContext interface {
	nameInParent() string
	indexInParent() int

	fieldSchema() schema_j5pb.IsField_Type
	propertySchema() *schema_j5pb.ObjectProperty
	typeName() string
	fullTypeName() string
	protoPath() []string
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

func (fieldDefaults) AsArray() (ArrayField, bool) {
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
