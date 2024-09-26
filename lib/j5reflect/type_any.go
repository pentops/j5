package j5reflect

import (
	"fmt"
	"strings"

	"github.com/pentops/j5/j5types/any_j5t"
	"github.com/pentops/j5/lib/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

/*** Interface ***/

type AnyField interface {
	Field

	SetJ5Any(val *any_j5t.Any) error
	GetJ5Any() (*any_j5t.Any, error)
	SetProtoAny(val *anypb.Any) error
	GetProtoAny() (*anypb.Any, error)
}

/*** Implementation ***/

type anyField struct {
	fieldDefaults
	fieldContext

	schema    *j5schema.AnyField
	value     protoreflect.Message
	valueType protoreflect.FullName
}

var _ AnyField = (*anyField)(nil)

func (field *anyField) IsSet() bool {
	// any has message by this point
	return true
}

var j5AnyType protoreflect.FullName
var j5AnyTypeField protoreflect.FieldDescriptor
var j5AnyProtoField protoreflect.FieldDescriptor
var j5AnyJ5JSONField protoreflect.FieldDescriptor

var pbAnyType protoreflect.FullName
var pbAnyTypeField protoreflect.FieldDescriptor
var pbAnyValueField protoreflect.FieldDescriptor

const anyPrefix = "type.googleapis.com/"

func init() {
	desc := (&any_j5t.Any{}).ProtoReflect().Descriptor()
	j5AnyType = desc.FullName()
	j5AnyTypeField = desc.Fields().ByName("type_name")
	j5AnyProtoField = desc.Fields().ByName("proto")
	j5AnyJ5JSONField = desc.Fields().ByName("j5_json")

	pbAnyDesc := (&anypb.Any{}).ProtoReflect().Descriptor()
	pbAnyType = pbAnyDesc.FullName()
	pbAnyTypeField = pbAnyDesc.Fields().ByName("type_url")
	pbAnyValueField = pbAnyDesc.Fields().ByName("value")
}

func (field *anyField) SetJ5Any(val *any_j5t.Any) error {
	return field.set(val.TypeName, val.Proto, val.J5Json)
}

func (field *anyField) SetProtoAny(val *anypb.Any) error {
	typeName := strings.TrimPrefix(val.TypeUrl, anyPrefix)
	return field.set(typeName, val.Value, nil)
}

func (field *anyField) set(typeName string, proto, j5json []byte) error {
	switch field.valueType {
	case pbAnyType:
		field.value.Set(pbAnyTypeField, protoreflect.ValueOfString(anyPrefix+typeName))
		if proto == nil {
			return fmt.Errorf("proto is required for PB Any type %s", typeName)
		}
		field.value.Set(pbAnyValueField, protoreflect.ValueOfBytes(proto))
		return nil
	case j5AnyType:
		field.value.Set(j5AnyTypeField, protoreflect.ValueOfString(typeName))
		if proto != nil {
			field.value.Set(j5AnyProtoField, protoreflect.ValueOfBytes(proto))
		}
		if j5json != nil {
			field.value.Set(j5AnyJ5JSONField, protoreflect.ValueOfBytes(j5json))
		}
		return nil
	default:
		return fmt.Errorf("unsupported Any type %s", field.valueType)
	}
}

func (field *anyField) GetJ5Any() (*any_j5t.Any, error) {
	switch field.valueType {
	case pbAnyType:
		typeName := strings.TrimPrefix(field.value.Get(pbAnyTypeField).String(), anyPrefix)
		return &any_j5t.Any{
			TypeName: typeName,
			Proto:    field.value.Get(pbAnyValueField).Bytes(),
		}, nil
	case j5AnyType:
		return &any_j5t.Any{
			TypeName: field.value.Get(j5AnyTypeField).String(),
			Proto:    field.value.Get(j5AnyProtoField).Bytes(),
			J5Json:   field.value.Get(j5AnyJ5JSONField).Bytes(),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported Any type %s", field.valueType)
	}
}

func (field *anyField) GetProtoAny() (*anypb.Any, error) {
	switch field.valueType {
	case pbAnyType:
		return &anypb.Any{
			TypeUrl: field.value.Get(pbAnyTypeField).String(),
			Value:   field.value.Get(pbAnyValueField).Bytes(),
		}, nil
	case j5AnyType:
		typeName := field.value.Get(j5AnyTypeField).String()
		if !field.value.Has(j5AnyProtoField) {
			return nil, fmt.Errorf("cannot convert from j5 any (%s) to proto without proto encoding", typeName)
		}
		return &anypb.Any{
			TypeUrl: anyPrefix + typeName,
			Value:   field.value.Get(j5AnyProtoField).Bytes(),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported Any type %s", field.valueType)
	}

}

var _ AnyField = (*anyField)(nil)

type anyFieldFactory struct {
	schema *j5schema.AnyField
}

func (factory *anyFieldFactory) buildField(context fieldContext, value protoreflect.Message) Field {
	valueType := value.Descriptor().FullName()
	return &anyField{
		schema:       factory.schema,
		value:        value,
		fieldContext: context,
		valueType:    valueType,
	}
}
