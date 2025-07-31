package j5reflect

import (
	"fmt"
	"strings"

	"github.com/pentops/j5/j5types/any_j5t"
	"github.com/pentops/j5/lib/j5reflect/protoval"
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

type MapOfAnyField interface {
	MapField
}

type ArrayOfAnyField interface {
	ArrayField
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
	isSet() bool
}

var _ AnyField = (*anyField)(nil)

func (field *anyField) IsSet() bool {
	if field.implType == nil {
		return false
	}
	return field.implType.isSet()
}

func (field *anyField) SetDefaultValue() error {
	// Default value for Any is not defined, so we do nothing here.
	return nil
}

func (field *anyField) AsAny() (AnyField, bool) {
	return field, true
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
	value        protoval.MessageValue
	typeUrlField protoreflect.FieldDescriptor
	valueField   protoreflect.FieldDescriptor
}

func newPbAnyImpl(value protoval.MessageValue) anyImpl {
	desc := value.MessageDescriptor()
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
	impl.value.MessageValue().Set(impl.typeUrlField, protoreflect.ValueOfString(anyPrefix+val.TypeName))
	if val.Proto == nil {
		return fmt.Errorf("proto is required for PB Any type %s", val.TypeName)
	}
	impl.value.MessageValue().Set(impl.valueField, protoreflect.ValueOfBytes(val.Proto))
	return nil
}

func (impl *pbAnyImpl) isSet() bool {
	return impl.value.IsSet()
}

func (impl *pbAnyImpl) getAny() (*any_j5t.Any, error) {
	typeUrl := impl.value.MessageValue().Get(impl.typeUrlField).String()
	typeName := strings.TrimPrefix(typeUrl, anyPrefix)
	return &any_j5t.Any{
		TypeName: typeName,
		Proto:    impl.value.MessageValue().Get(impl.valueField).Bytes(),
	}, nil
}

type exitAnyImpl struct {
	value     protoval.MessageValue
	valueType string
}

func newExitAnyImpl(valueType string, value protoval.MessageValue) anyImpl {
	return &exitAnyImpl{
		value:     value,
		valueType: valueType,
	}
}

func (impl *exitAnyImpl) isSet() bool {
	return impl.value.IsSet()
}

func (impl *exitAnyImpl) setAny(val *any_j5t.Any) error {
	return fmt.Errorf("setAny not implemented for exit Any type %s", impl.valueType)
}

func (impl *exitAnyImpl) getAny() (*any_j5t.Any, error) {
	return nil, fmt.Errorf("getAny not implemented for exit Any type %s", impl.valueType)
}

type j5AnyImpl struct {
	value         protoval.MessageValue
	typeNameField protoreflect.FieldDescriptor
	protoField    protoreflect.FieldDescriptor
	j5JsonField   protoreflect.FieldDescriptor
}

func newJ5AnyImpl(value protoval.MessageValue) anyImpl {
	desc := value.MessageDescriptor()
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
	impl.value.MessageValue().Set(impl.typeNameField, protoreflect.ValueOfString(val.TypeName))
	if val.Proto != nil {
		impl.value.MessageValue().Set(impl.protoField, protoreflect.ValueOfBytes(val.Proto))
	}
	if val.J5Json != nil {
		impl.value.MessageValue().Set(impl.j5JsonField, protoreflect.ValueOfBytes(val.J5Json))
	}
	return nil
}

func (impl *j5AnyImpl) isSet() bool {
	return impl.value.IsSet()
}

func (impl *j5AnyImpl) getAny() (*any_j5t.Any, error) {
	if !impl.value.IsSet() {
		return nil, nil
	}
	val := impl.value.MessageValue()
	typeName := val.Get(impl.typeNameField).String()
	out := &any_j5t.Any{
		TypeName: typeName,
	}
	if val.Has(impl.protoField) {
		out.Proto = val.Get(impl.protoField).Bytes()
	}

	if val.Has(impl.j5JsonField) {
		out.J5Json = val.Get(impl.j5JsonField).Bytes()
	}
	return out, nil
}

var _ AnyField = (*anyField)(nil)

type anyFieldFactory struct {
	schema *j5schema.AnyField
}

func (factory *anyFieldFactory) buildField(context fieldContext, value protoval.Value) Field {
	msgVal, ok := value.AsMessage()
	if !ok {
		panic(fmt.Sprintf("expected a message value for Any field, got %s", value))
	}

	valueType := msgVal.MessageDescriptor().FullName()
	var impl anyImpl
	switch valueType {
	case "google.protobuf.Any":
		impl = newPbAnyImpl(msgVal)
	case "j5.types.any.v1.Any":
		impl = newJ5AnyImpl(msgVal)
	default:
		impl = newExitAnyImpl(string(valueType), msgVal)
	}

	return &anyField{
		schema:       factory.schema,
		fieldContext: context,
		implType:     impl,
	}
}

/*** Implement Array Of Any ***/

type arrayOfAnyField struct {
	mutableArrayField
}

var _ ArrayOfAnyField = (*arrayOfAnyField)(nil)

/*** Implement Map Of Any ***/

type mapOfAnyField struct {
	mutableMapField
}

var _ MapOfAnyField = (*mapOfAnyField)(nil)
