package j5reflect

import (
	"fmt"

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
}

type propertyContext struct {
	schema *j5schema.ObjectProperty
}

func (c propertyContext) nameInParent() string {
	return c.schema.JSONName
}

type fieldDefaults struct {
	fieldType FieldType
	context   fieldContext
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

func (f *fieldDefaults) setContext(c fieldContext) {
	f.context = c
}

type Field interface {
	Type() FieldType
	IsSet() bool
	SetDefault() error

	// NameInParent is the name this field has in the context it exists.
	// Object and Oneof: The property name
	// Map<string>x, the key
	// Arrays, the index as a string
	NameInParent() string
	setContext(c fieldContext)

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
}

type ArrayOfContainerField interface {
	MutableArrayField
	NewContainerElement() (PropertySet, error)
}

type fieldFactory interface {
	buildField(value protoValueContext) Field
}

type objectFieldFactory struct {
	schema *j5schema.ObjectField
}

func (f *objectFieldFactory) buildField(value protoValueContext) Field {
	return newObjectField(f.schema, value)
}

type oneofFieldFactory struct {
	schema *j5schema.OneofField
}

func (f *oneofFieldFactory) buildField(value protoValueContext) Field {
	return newOneofField(f.schema, value)
}

type enumFieldFactory struct {
	schema *j5schema.EnumField
}

func (f *enumFieldFactory) buildField(value protoValueContext) Field {
	return newEnumField(f.schema, value)
}

type scalarFieldFactory struct {
	schema *j5schema.ScalarSchema
}

func (f *scalarFieldFactory) buildField(value protoValueContext) Field {
	return newScalarField(f.schema, value)
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
