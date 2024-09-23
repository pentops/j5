package j5reflect

import (
	"github.com/pentops/j5/lib/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

/*** Interface ***/

type AnyField interface {
	Field
}

/*** Implementation ***/

type anyField struct {
	fieldDefaults
	fieldContext

	schema *j5schema.AnyField
	value  protoreflect.Message
}

func (field *anyField) IsSet() bool {
	// any has message by this point
	return true
}

var _ AnyField = (*anyField)(nil)

type anyFieldFactory struct {
	schema *j5schema.AnyField
}

func (factory *anyFieldFactory) buildField(context fieldContext, value protoreflect.Message) Field {
	return &anyField{
		schema: factory.schema,
		value:  value,
	}
}
