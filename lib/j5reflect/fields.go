package j5reflect

import (
	"github.com/pentops/j5/lib/j5schema"
)

type Field interface {
	IsSet() bool
	SetDefaultValue() error
	FieldContext

	// Fighting with go typing here, the implementations of these return
	// themselves and true. Maybe I should have fixated on java instead.

	AsScalar() (ScalarField, bool)
	AsEnum() (EnumField, bool)
	AsAny() (AnyField, bool)

	AsContainer() (ContainerField, bool)
	AsObject() (ObjectField, bool)
	AsOneof() (OneofField, bool)
	AsPolymorph() (PolymorphField, bool)

	AsArray() (ArrayField, bool)
	AsArrayOfScalar() (ArrayOfScalarField, bool)
	AsArrayOfContainer() (ArrayOfContainerField, bool)
	AsArrayOfOneof() (ArrayOfOneofField, bool)
	AsArrayOfObject() (ArrayOfObjectField, bool)

	AsMap() (MapField, bool)
	AsMapOfScalar() (MapOfScalarField, bool)
	AsMapOfContainer() (MapOfContainerField, bool)
	AsMapOfObject() (MapOfObjectField, bool)
	AsMapOfOneof() (MapOfOneofField, bool)
}

type FieldContext interface {
	// NameInParent is the name this field in the parent
	// Object and Oneof: The property name
	// Map<string>x, the key
	// Arrays, the index as a string
	NameInParent() string

	// IndexInParent returns -1 for non array fields
	IndexInParent() int
	PropertySchema() *j5schema.ObjectProperty
	FieldSchema() j5schema.FieldSchema
	TypeName() string
	FullTypeName() string
	ProtoPath() []string
}

// fieldDefaults is embedded into all field types to allow easy extension of the
// 'false' answers here. Individual types implement as 'true'
type fieldDefaults struct{}

func (fieldDefaults) AsScalar() (ScalarField, bool)                     { return nil, false }
func (fieldDefaults) AsEnum() (EnumField, bool)                         { return nil, false }
func (fieldDefaults) AsAny() (AnyField, bool)                           { return nil, false }
func (fieldDefaults) AsContainer() (ContainerField, bool)               { return nil, false }
func (fieldDefaults) AsObject() (ObjectField, bool)                     { return nil, false }
func (fieldDefaults) AsOneof() (OneofField, bool)                       { return nil, false }
func (fieldDefaults) AsPolymorph() (PolymorphField, bool)               { return nil, false }
func (fieldDefaults) AsArray() (ArrayField, bool)                       { return nil, false }
func (fieldDefaults) AsArrayOfContainer() (ArrayOfContainerField, bool) { return nil, false }
func (fieldDefaults) AsArrayOfScalar() (ArrayOfScalarField, bool)       { return nil, false }
func (fieldDefaults) AsArrayOfObject() (ArrayOfObjectField, bool)       { return nil, false }
func (fieldDefaults) AsArrayOfOneof() (ArrayOfOneofField, bool)         { return nil, false }
func (fieldDefaults) AsMap() (MapField, bool)                           { return nil, false }
func (fieldDefaults) AsMapOfScalar() (MapOfScalarField, bool)           { return nil, false }
func (fieldDefaults) AsMapOfContainer() (MapOfContainerField, bool)     { return nil, false }
func (fieldDefaults) AsMapOfObject() (MapOfObjectField, bool)           { return nil, false }
func (fieldDefaults) AsMapOfOneof() (MapOfOneofField, bool)             { return nil, false }

type fieldContext interface {
	FieldContext
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

func (c propertyContext) TypeName() string {
	return c.schema.Schema.TypeName()
}

func (c propertyContext) PropertySchema() *j5schema.ObjectProperty {
	return c.schema
}
func (c propertyContext) FieldSchema() j5schema.FieldSchema { //schema_j5pb.IsField_Type {
	return c.schema.Schema //.ToJ5Field().Type
}

func (c propertyContext) FullTypeName() string {
	return c.schema.FullName()
}

func (c propertyContext) ProtoPath() []string {
	return c.walkedProtoPath
}
