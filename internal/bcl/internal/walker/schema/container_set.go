package schema

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
)

type Container interface {
	Path() []string
	Spec() BlockSpec
	Name() string
	SchemaName() string
}

type schemaFlags struct {
	canAttribute bool
	canBlock     bool
}

func (sf schemaFlags) GoString() string {
	return fmt.Sprintf("schema: Attr %t, Block: %t}", sf.canAttribute, sf.canBlock)
}

func schemaCan(st schema_j5pb.IsField_Type) schemaFlags {
	switch st.(type) {
	case *schema_j5pb.Field_Bool,
		*schema_j5pb.Field_Bytes,
		*schema_j5pb.Field_String_,
		*schema_j5pb.Field_Date,
		*schema_j5pb.Field_Timestamp,
		*schema_j5pb.Field_Decimal,
		*schema_j5pb.Field_Float,
		*schema_j5pb.Field_Integer,
		*schema_j5pb.Field_Key:
		return schemaFlags{canAttribute: true, canBlock: false}

	case *schema_j5pb.Field_Any:
		return schemaFlags{canAttribute: false, canBlock: false}

	case *schema_j5pb.Field_Array:
		return schemaFlags{canAttribute: false, canBlock: true}

	case *schema_j5pb.Field_Object:
		return schemaFlags{canAttribute: false, canBlock: true}

	case *schema_j5pb.Field_Map:
		return schemaFlags{canAttribute: false, canBlock: true}

	case *schema_j5pb.Field_Enum:
		return schemaFlags{canAttribute: true, canBlock: false}

	case *schema_j5pb.Field_Oneof:
		return schemaFlags{canAttribute: false, canBlock: true}
	default:
		return schemaFlags{}
	}
}
