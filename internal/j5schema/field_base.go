package j5schema

import "github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"

type FieldSchema interface {
	ToJ5Field() *schema_j5pb.Field

	TypeName() string

	ParentContext() (ContainerSchema, string)
	FullName() string

	// Mutable determines how reflection can access the Field
	// true for ObjectField and OneofField
	// false for ScalarField, EnumField
	// true for ArrayField and MapField
	// false for AnyField special case
	Mutable() bool
}

type ContainerSchema interface {
	FullName() string
}

type fieldContext struct {
	//propertyParent *ObjectProperty
	//arrayParent    *ArrayField
	//mapParent      *MapField
	parent       ContainerSchema
	nameInParent string
}

func (f *fieldContext) ParentContext() (ContainerSchema, string) {
	return f.parent, f.nameInParent
}

func (f *fieldContext) FullName() string {
	return f.parent.FullName() + "." + f.nameInParent
}

func baseTypeName(st schema_j5pb.IsField_Type) string {
	switch st.(type) {
	case *schema_j5pb.Field_Any:
		return "any"
	case *schema_j5pb.Field_Boolean:
		return "bool"
	case *schema_j5pb.Field_Bytes:
		return "bytes"
	case *schema_j5pb.Field_Array:
		return "array"
	case *schema_j5pb.Field_Object:
		return "object"
	case *schema_j5pb.Field_Map:
		return "map"
	case *schema_j5pb.Field_Enum:
		return "enum"
	case *schema_j5pb.Field_Oneof:
		return "oneof"
	case *schema_j5pb.Field_String_:
		return "string"
	case *schema_j5pb.Field_Date:
		return "date"
	case *schema_j5pb.Field_Timestamp:
		return "timestamp"
	case *schema_j5pb.Field_Decimal:
		return "decimal"
	case *schema_j5pb.Field_Float:
		return "float"
	case *schema_j5pb.Field_Integer:
		return "integer"
	case *schema_j5pb.Field_Key:
		return "key"
	default:
		return "unknown"
	}
}
