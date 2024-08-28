package j5reflect

import (
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/internal/j5schema"
)

type Field interface {
	IsSet() bool
	FieldContext

	// Fighting with go typing here, the implementations of these return
	// themselves and true. Maybe I should have fixated on java instead.
	AsArray() (ArrayField, bool)
	AsScalar() (ScalarField, bool)
	AsContainer() (ContainerField, bool)
	AsArrayOfContainer() (ArrayOfContainerField, bool)
	AsArrayOfScalar() (ArrayOfScalarField, bool)
	AsObject() (ObjectField, bool)
	AsOneof() (OneofField, bool)
	AsEnum() (EnumField, bool)
}

type FieldContext interface {
	// NameInParent is the name this field in the parent
	// Object and Oneof: The property name
	// Map<string>x, the key
	// Arrays, the index as a string
	NameInParent() string

	// IndexInParent returns -1 for non array fields
	IndexInParent() int
	PropertySchema() *schema_j5pb.ObjectProperty
	FieldSchema() schema_j5pb.IsField_Type
	TypeName() string
	FullTypeName() string
	ProtoPath() []string
}

type fieldDefaults struct {
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

func (fieldDefaults) AsObject() (ObjectField, bool) {
	return nil, false
}

func (fieldDefaults) AsOneof() (OneofField, bool) {
	return nil, false
}

func (fieldDefaults) AsEnum() (EnumField, bool) {
	return nil, false
}

type fieldContext interface {
	FieldContext

	// not exported
}

type propertyContext struct {
	walkedProtoPath []string
	schema          *j5schema.ObjectProperty
}

func (c propertyContext) NameInParent() string {
	return c.schema.JSONName
}

func (c propertyContext) IndexInParent() int {
	return -1
}

func (c propertyContext) PropertySchema() *schema_j5pb.ObjectProperty {
	return c.schema.ToJ5Proto()
}

func (c propertyContext) TypeName() string {
	return c.schema.Schema.TypeName()
}

func (c propertyContext) FieldSchema() schema_j5pb.IsField_Type {
	return c.schema.Schema.ToJ5Field().Type
}

func (c propertyContext) FullTypeName() string {
	return c.schema.FullName()
}

func (c propertyContext) ProtoPath() []string {
	return c.walkedProtoPath
}

func (c propertyContext) fieldSchema() schema_j5pb.IsField_Type {
	return c.schema.Schema.ToJ5Field().Type
}
