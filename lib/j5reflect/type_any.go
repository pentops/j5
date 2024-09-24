package j5reflect

import (
	"github.com/pentops/j5/lib/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

/*** Interface ***/

type AnyField interface {
	Field

	SetProtoAny(val *anypb.Any)
	GetProtoAny() *anypb.Any
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

var anyTypeField protoreflect.FieldDescriptor
var anyValueField protoreflect.FieldDescriptor

func init() {
	desc := (&anypb.Any{}).ProtoReflect().Descriptor()
	anyTypeField = desc.Fields().ByName("type_url")
	anyValueField = desc.Fields().ByName("value")
}

func (field *anyField) SetProtoAny(val *anypb.Any) {
	field.value.Set(anyTypeField, protoreflect.ValueOfString(val.TypeUrl))
	field.value.Set(anyValueField, protoreflect.ValueOfBytes(val.Value))
}

func (field *anyField) GetProtoAny() *anypb.Any {
	return &anypb.Any{
		TypeUrl: field.value.Get(anyTypeField).String(),
		Value:   field.value.Get(anyValueField).Bytes(),
	}
}

var _ AnyField = (*anyField)(nil)

type anyFieldFactory struct {
	schema *j5schema.AnyField
}

func (factory *anyFieldFactory) buildField(context fieldContext, value protoreflect.Message) Field {
	return &anyField{
		schema:       factory.schema,
		value:        value,
		fieldContext: context,
	}
}
