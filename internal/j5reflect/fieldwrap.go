package j5reflect

import (
	"google.golang.org/protobuf/reflect/protoreflect"
)

type protoValue interface {
	IsSet() bool
	Type() FieldType
	getValue() protoreflect.Value
	setValue(protoreflect.Value) error
}

type fieldWrapper struct {
	msg       protoreflect.Message
	field     protoreflect.FieldDescriptor
	fieldType FieldType
}

func newFieldWrapper(fieldType FieldType, msg protoreflect.Message, field protoreflect.FieldDescriptor) *fieldWrapper {
	return &fieldWrapper{
		msg:       msg,
		field:     field,
		fieldType: fieldType,
	}
}

func (fw *fieldWrapper) getValue() protoreflect.Value {
	return fw.msg.Get(fw.field)
}

func (fw *fieldWrapper) setValue(val protoreflect.Value) error {
	// parent, walk up, set everything
	return nil
}

func (wf *fieldWrapper) Type() FieldType {
	return FieldTypeEnum
}

func (wf *fieldWrapper) IsSet() bool {
	return wf.msg.Has(wf.field)
}

type listFieldValue struct {
	value     protoreflect.Value
	idx       int
	list      protoreflect.List
	fieldType FieldType
}

func (lfv *listFieldValue) IsSet() bool {
	return true
}

func (lfv *listFieldValue) Type() FieldType {
	return lfv.fieldType
}

func (lfv *listFieldValue) getValue() protoreflect.Value {
	return lfv.value
}

func (lfv *listFieldValue) setValue(val protoreflect.Value) error {
	lfv.list.Set(lfv.idx, val)
	return nil
}

type mapFieldValue struct {
	value     protoreflect.Value
	key       protoreflect.MapKey
	mapVal    protoreflect.Map
	fieldType FieldType
}

func (mfv *mapFieldValue) IsSet() bool {
	return true
}

func (mfv *mapFieldValue) Type() FieldType {
	return mfv.fieldType
}

func (mfv *mapFieldValue) getValue() protoreflect.Value {
	return mfv.value
}

func (mfv *mapFieldValue) setValue(val protoreflect.Value) error {
	return nil
}
