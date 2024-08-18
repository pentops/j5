package j5reflect

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type protoValueContext interface {
	getValue() protoreflect.Value
	kind() protoreflect.Kind
	IsSet() bool
	setValue(protoreflect.Value)
	getOrCreate() protoreflect.Value
	getOrCreateMutable() protoreflect.Value
}

type realProtoMessageField struct {
	msg   *protoMessage
	field protoreflect.FieldDescriptor
}

var _ protoValueContext = &realProtoMessageField{}

func (mf *realProtoMessageField) kind() protoreflect.Kind {
	return mf.field.Kind()
}

func (mf *realProtoMessageField) getValue() protoreflect.Value {
	if !mf.IsSet() {
		return protoreflect.Value{}
	}
	return mf.msg.protoReflectMessage.Get(mf.field)
}

func (mf *realProtoMessageField) setValue(val protoreflect.Value) {
	mf.msg.ensureExists()
	mf.msg.protoReflectMessage.Set(mf.field, val)
}

func (mf *realProtoMessageField) getOrCreate() protoreflect.Value {
	mf.msg.ensureExists()
	if mf.msg.protoReflectMessage.Has(mf.field) {
		return mf.msg.protoReflectMessage.Get(mf.field)
	}
	return mf.msg.protoReflectMessage.NewField(mf.field)
}

func (mf *realProtoMessageField) getOrCreateMutable() protoreflect.Value {
	mf.msg.ensureExists()
	return mf.msg.protoReflectMessage.Mutable(mf.field)
}

func (mf *realProtoMessageField) IsSet() bool {
	if !mf.msg.exists() {
		return false
	}
	return mf.msg.protoReflectMessage.Has(mf.field)
}

// vitualProtoField is a non-proto field which appears in client schemas, it is
// a j5 schema around a sub-set of fields in a parent message
type virtualProtoField struct {
	msg      *protoMessage
	children *propSet
}

var _ protoValueContext = &virtualProtoField{}

func (vf *virtualProtoField) kind() protoreflect.Kind {
	return protoreflect.MessageKind
}

func (vf *virtualProtoField) getOrCreate() protoreflect.Value {
	// the virtual field doesn't exist, but the parent message where the fields
	// live needs to exist to set values.
	vf.msg.ensureExists()
	return protoreflect.Value{}
}

func (vf *virtualProtoField) getOrCreateMutable() protoreflect.Value {
	return vf.getOrCreate()
}

func (vf *virtualProtoField) getValue() protoreflect.Value {
	return protoreflect.Value{}
}

func (vf *virtualProtoField) setValue(val protoreflect.Value) {
	// There is nothing logical which could be set.
}

func (vf *virtualProtoField) IsSet() bool {
	// The field has no actual value, so we check if any child is set. For
	// oneofs, this will be the one field
	return vf.children.AnySet()
}

// protoValueWrapper is used where the value must have already been created to
// exist, unlike a proto field, it does not need to 'walk up the tree' to make
// sure the fields are created.
// Used in List and Map.
type protoValueWrapper struct {
	value   protoreflect.Value
	prField protoreflect.FieldDescriptor
}

func (mfv *protoValueWrapper) kind() protoreflect.Kind {
	return mfv.prField.Kind()
}

func (mfv *protoValueWrapper) IsSet() bool {
	return true
}

func (mfv *protoValueWrapper) getOrCreate() protoreflect.Value {
	return mfv.value
}

func (mfv *protoValueWrapper) getOrCreateMutable() protoreflect.Value {
	return mfv.value
}

func (mfv *protoValueWrapper) getValue() protoreflect.Value {
	return mfv.value
}

// protoListItem is a single item in a repeated field
type protoListItem struct {
	protoValueWrapper
	idx    int
	prList protoreflect.List
}

var _ protoValueContext = &protoListItem{}

func (lfv *protoListItem) setValue(val protoreflect.Value) {
	lfv.prList.Set(lfv.idx, val)
}

type protoMapItem struct {
	protoValueWrapper
	key   protoreflect.MapKey
	prMap protoreflect.Map
}

func (mfv *protoMapItem) setValue(val protoreflect.Value) {
	mfv.prMap.Set(mfv.key, val)
}

type valueParent interface {
	getOrCreate() protoreflect.Value
}

type protoMessage struct {
	protoReflectMessage protoreflect.Message
	descriptor          protoreflect.MessageDescriptor
	parent              valueParent
}

func newChildMessageValue(parent valueParent, value protoreflect.Message) (*protoMessage, error) {
	if parent == nil {
		return nil, fmt.Errorf("parent is nil")
	}
	if value == nil || !value.IsValid() {
		return nil, fmt.Errorf("value is nil")
	}

	msg := &protoMessage{
		descriptor:          value.Descriptor(),
		protoReflectMessage: value,
		parent:              parent,
	}

	return msg, nil
}

func newRootMessageValue(msg protoreflect.Message, descriptor protoreflect.MessageDescriptor) (*protoMessage, error) {
	if msg == nil {
		return nil, fmt.Errorf("message is nil")
	}
	if descriptor == nil {
		return nil, fmt.Errorf("descriptor is nil")
	}

	return &protoMessage{
		protoReflectMessage: msg,
		descriptor:          descriptor,
	}, nil
}

func (mv *protoMessage) ensureExists() {
	if mv.protoReflectMessage != nil {
		return
	}
	msg := mv.parent.getOrCreate()
	mv.protoReflectMessage = msg.Message()
}

func (mv *protoMessage) exists() bool {
	return mv.protoReflectMessage != nil
}

func (mv *protoMessage) getOrCreate() protoreflect.Value {
	mv.ensureExists()
	return protoreflect.ValueOfMessage(mv.protoReflectMessage)
}

func (mv *protoMessage) childByNumber(field protoreflect.FieldNumber) (*protoMessage, error) {
	fd := mv.protoReflectMessage.Descriptor().Fields().ByNumber(field)
	if fd.Kind() != protoreflect.MessageKind {
		return nil, fmt.Errorf("field %s is not a message but has nested types", fd.FullName())
	}
	desc := fd.Message()
	if desc == nil {
		return nil, fmt.Errorf("field %s is a message but has no descriptor", fd.FullName())
	}

	child := &protoMessage{
		descriptor: desc,
		parent:     mv,
	}
	if mv.protoReflectMessage.Has(fd) {
		child.protoReflectMessage = mv.protoReflectMessage.Get(fd).Message()
	}

	return child, nil
}

func (mv *protoMessage) fieldByNumber(field protoreflect.FieldNumber) (*realProtoMessageField, error) {
	fd := mv.descriptor.Fields().ByNumber(field)

	if fd == nil {
		return nil, fmt.Errorf("field is nil")
	}

	return &realProtoMessageField{
		msg:   mv,
		field: fd,
	}, nil
}

func (mv *protoMessage) virtualField(children *propSet) *virtualProtoField {
	return &virtualProtoField{
		msg:      mv,
		children: children,
	}
}
