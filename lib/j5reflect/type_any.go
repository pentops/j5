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

	schema *j5schema.AnyField

	implType anyImpl
}

type anyImpl interface {
	setAny(*any_j5t.Any) error
	getAny() (*any_j5t.Any, error)
}

var _ AnyField = (*anyField)(nil)

func (field *anyField) IsSet() bool {
	// any has message by this point
	return true
}

func (field *anyField) SetJ5Any(val *any_j5t.Any) error {
	return field.implType.setAny(val)
}

func (field *anyField) SetProtoAny(val *anypb.Any) error {
	typeName := strings.TrimPrefix(val.TypeUrl, anyPrefix)
	return field.implType.setAny(&any_j5t.Any{
		TypeName: typeName,
		Proto:    val.Value,
	})

}

func (field *anyField) GetJ5Any() (*any_j5t.Any, error) {
	val, err := field.implType.getAny()
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (field *anyField) GetProtoAny() (*anypb.Any, error) {
	val, err := field.implType.getAny()
	if err != nil {
		return nil, err
	}
	if val.Proto == nil {
		return nil, fmt.Errorf("proto is required for PB Any type %s", val.TypeName)
	}
	return &anypb.Any{
		TypeUrl: anyPrefix + val.TypeName,
		Value:   val.Proto,
	}, nil
}

type pbAnyImpl struct {
	value        protoreflect.Message
	typeUrlField protoreflect.FieldDescriptor
	valueField   protoreflect.FieldDescriptor
}

func newPbAnyImpl(value protoreflect.Message) anyImpl {
	desc := value.Descriptor()
	typeUrlField := desc.Fields().ByName("type_url")
	valueField := desc.Fields().ByName("value")
	return &pbAnyImpl{
		value:        value,
		typeUrlField: typeUrlField,
		valueField:   valueField,
	}
}

const anyPrefix = "type.googleapis.com/"

func (impl *pbAnyImpl) setAny(val *any_j5t.Any) error {
	impl.value.Set(impl.typeUrlField, protoreflect.ValueOfString(anyPrefix+val.TypeName))
	if val.Proto == nil {
		return fmt.Errorf("proto is required for PB Any type %s", val.TypeName)
	}
	impl.value.Set(impl.valueField, protoreflect.ValueOfBytes(val.Proto))
	return nil
}

func (impl *pbAnyImpl) getAny() (*any_j5t.Any, error) {
	typeUrl := impl.value.Get(impl.typeUrlField).String()
	typeName := strings.TrimPrefix(typeUrl, anyPrefix)
	return &any_j5t.Any{
		TypeName: typeName,
		Proto:    impl.value.Get(impl.valueField).Bytes(),
	}, nil
}

type j5AnyImpl struct {
	value         protoreflect.Message
	typeNameField protoreflect.FieldDescriptor
	protoField    protoreflect.FieldDescriptor
	j5JsonField   protoreflect.FieldDescriptor
}

func newJ5AnyImpl(value protoreflect.Message) anyImpl {
	desc := value.Descriptor()
	typeNameField := desc.Fields().ByName("type_name")
	protoField := desc.Fields().ByName("proto")
	j5JsonField := desc.Fields().ByName("j5_json")
	return &j5AnyImpl{
		value:         value,
		typeNameField: typeNameField,
		protoField:    protoField,
		j5JsonField:   j5JsonField,
	}
}

func (impl *j5AnyImpl) setAny(val *any_j5t.Any) error {
	impl.value.Set(impl.typeNameField, protoreflect.ValueOfString(val.TypeName))
	if val.Proto != nil {
		impl.value.Set(impl.protoField, protoreflect.ValueOfBytes(val.Proto))
	}
	if val.J5Json != nil {
		impl.value.Set(impl.j5JsonField, protoreflect.ValueOfBytes(val.J5Json))
	}
	return nil
}

func (impl *j5AnyImpl) getAny() (*any_j5t.Any, error) {
	typeName := impl.value.Get(impl.typeNameField).String()
	out := &any_j5t.Any{
		TypeName: typeName,
	}
	if impl.value.Has(impl.protoField) {
		out.Proto = impl.value.Get(impl.protoField).Bytes()
	}

	if impl.value.Has(impl.j5JsonField) {
		out.J5Json = impl.value.Get(impl.j5JsonField).Bytes()
	}
	return out, nil
}

var _ AnyField = (*anyField)(nil)

type anyFieldFactory struct {
	schema *j5schema.AnyField
}

func (factory *anyFieldFactory) buildField(context fieldContext, value protoreflect.Message) Field {
	valueType := value.Descriptor().FullName()
	var impl anyImpl
	switch valueType {
	case "google.protobuf.Any":
		impl = newPbAnyImpl(value)
	case "j5.types.any.v1.Any":
		impl = newJ5AnyImpl(value)
	default:
		panic(fmt.Sprintf("unsupported Any type %s", valueType))
	}

	return &anyField{
		schema:       factory.schema,
		fieldContext: context,
		implType:     impl,
	}
}
