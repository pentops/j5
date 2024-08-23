package j5reflect

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type protoValueContext interface {
	getValue() protoreflect.Value
	kind() protoreflect.Kind
	isSet() bool
	setValue(protoreflect.Value)
	clearValue() error
	getOrCreate() protoreflect.Value
	getOrCreateMutable() protoreflect.Value
	getOrCreateChildMessage() (*protoMessageWrapper, error)
}

type realProtoMessageField struct {
	parent        *protoMessageWrapper
	fieldInParent protoreflect.FieldDescriptor
}

var _ protoValueContext = &realProtoMessageField{}

func (mf *realProtoMessageField) kind() protoreflect.Kind {
	return mf.fieldInParent.Kind()
}

func (mf *realProtoMessageField) getValue() protoreflect.Value {
	if !mf.isSet() {
		return protoreflect.Value{}
	}
	return mf.parent.protoReflectMessage.Get(mf.fieldInParent)
}

func (mf *realProtoMessageField) setValue(val protoreflect.Value) {
	mf.parent.ensureExists()
	if !val.IsValid() {
		mf.parent.protoReflectMessage.Clear(mf.fieldInParent)
		return
	}

	mf.parent.protoReflectMessage.Set(mf.fieldInParent, val)
}

func (mf *realProtoMessageField) clearValue() error {
	mf.parent.ensureExists()
	mf.parent.protoReflectMessage.Clear(mf.fieldInParent)
	return nil
}

func (mf *realProtoMessageField) getOrCreate() protoreflect.Value {
	mf.parent.ensureExists()
	if mf.parent.protoReflectMessage.Has(mf.fieldInParent) {
		return mf.parent.protoReflectMessage.Get(mf.fieldInParent)
	}
	return mf.parent.protoReflectMessage.NewField(mf.fieldInParent)
}

func (mf *realProtoMessageField) getOrCreateMutable() protoreflect.Value {
	mf.parent.ensureExists()
	return mf.parent.protoReflectMessage.Mutable(mf.fieldInParent)
}

func (mf *realProtoMessageField) getOrCreateChildMessage() (*protoMessageWrapper, error) {
	mf.parent.ensureExists()

	if mf.fieldInParent.Kind() != protoreflect.MessageKind {
		return nil, fmt.Errorf("field is not a message but has nested types")
	}

	desc := mf.fieldInParent.Message()
	if desc == nil {
		return nil, fmt.Errorf("field is a message but has no descriptor")
	}

	return mf.parent.fieldAsWrapper(mf.fieldInParent)
}

func (mf *realProtoMessageField) isSet() bool {
	if !mf.parent.exists() {
		return false
	}
	return mf.parent.protoReflectMessage.Has(mf.fieldInParent)
}

// vitualProtoField is a non-proto field which appears in client schemas, it is
// a j5 schema around a sub-set of fields in a parent message
type virtualProtoField struct {
	msg      *protoMessageWrapper
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

func (vf *virtualProtoField) getOrCreateChildMessage() (*protoMessageWrapper, error) {
	return nil, fmt.Errorf("cannot get child message of a virtual field")
}

func (vf *virtualProtoField) getValue() protoreflect.Value {
	return protoreflect.Value{}
}

func (vf *virtualProtoField) setValue(val protoreflect.Value) {
	// There is nothing logical which could be set.
}

func (vf *virtualProtoField) clearValue() error {
	// TODO: Clear the children?
	return fmt.Errorf("cannot clear a virtual field")
}

func (vf *virtualProtoField) isSet() bool {
	// The field has no actual value, so we check if any child is set. For
	// oneofs, this will be the one field
	for _, prop := range vf.children.asSlice {
		if prop.IsSet() {
			return true
		}
	}
	return false
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

func (mfv *protoValueWrapper) isSet() bool {
	return true
}

func (mfv *protoValueWrapper) getOrCreate() protoreflect.Value {
	return mfv.value
}

func (mfv *protoValueWrapper) getOrCreateMutable() protoreflect.Value {
	return mfv.value
}

func (mfv *protoValueWrapper) getOrCreateChildMessage() (*protoMessageWrapper, error) {
	return newRootMessageValue(mfv.value.Message())
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

func (lfv *protoListItem) clearValue() error {
	return fmt.Errorf("cannot clear a list item")
}

type protoMapItem struct {
	protoValueWrapper
	key   protoreflect.MapKey
	prMap protoreflect.Map
}

func (mfv *protoMapItem) setValue(val protoreflect.Value) {
	mfv.prMap.Set(mfv.key, val)
}

func (mfv *protoMapItem) clearValue() error {
	mfv.prMap.Clear(mfv.key)
	return nil
}

type protoMessageWrapper struct {
	descriptor protoreflect.MessageDescriptor

	// may be nil if the value is not yet set.
	protoReflectMessage protoreflect.Message

	fieldInParent protoreflect.FieldDescriptor
	parent        *protoMessageWrapper
}

func newRootMessageValue(msg protoreflect.Message) (*protoMessageWrapper, error) {
	if msg == nil {
		return nil, fmt.Errorf("message is nil")
	}

	return &protoMessageWrapper{
		descriptor:          msg.Descriptor(),
		protoReflectMessage: msg,
		parent:              nil,
		fieldInParent:       nil,
	}, nil
}

func (pmw *protoMessageWrapper) ensureExists() {
	if pmw.protoReflectMessage != nil {
		return
	}
	msg := pmw.parent.getOrCreate(pmw.fieldInParent)
	pmw.protoReflectMessage = msg
}

func (pmw *protoMessageWrapper) exists() bool {
	return pmw.protoReflectMessage != nil
}

func (pmw *protoMessageWrapper) getOrCreate(field protoreflect.FieldDescriptor) protoreflect.Message {
	pmw.ensureExists()
	if pmw.protoReflectMessage.Has(field) {
		return pmw.protoReflectMessage.Get(field).Message()
	}
	return pmw.protoReflectMessage.Mutable(field).Message()
}

func (pmw *protoMessageWrapper) fieldAsWrapper(fd protoreflect.FieldDescriptor) (*protoMessageWrapper, error) {
	desc := fd.Message()
	if desc == nil {
		return nil, fmt.Errorf("field %s is a message but has no descriptor", fd.FullName())
	}

	// The 'child' message is field 'fd' in the parent message 'pmw'
	child := &protoMessageWrapper{
		descriptor: desc,

		fieldInParent: fd,
		parent:        pmw,
	}

	//child.logName("new child node")

	if pmw.protoReflectMessage != nil {
		if pmw.protoReflectMessage.Has(fd) {
			child.protoReflectMessage = pmw.protoReflectMessage.Get(fd).Message()
		}
	}

	return child, nil
}

/*
func (pmw *protoMessageWrapper) debugPrint(as string) {
	fmt.Printf("%s\n | field %s \n |  in   %s\n |  is a %s\n", as, pmw.fieldInParent.FullName(), pmw.parent.descriptor.FullName(), pmw.descriptor.FullName())
	if !strings.HasPrefix(string(pmw.fieldInParent.FullName()), string(pmw.parent.descriptor.FullName())) {
		panic("field and message names do not match")
	}
}*/

func (pmw *protoMessageWrapper) fieldByNumber(field protoreflect.FieldNumber) (*realProtoMessageField, error) {
	if pmw.descriptor == nil {
		return nil, fmt.Errorf("descriptor is nil in fieldByNumber")
	}
	fd := pmw.descriptor.Fields().ByNumber(field)
	if fd == nil {
		return nil, fmt.Errorf("field is nil")
	}

	leafNode := &realProtoMessageField{
		parent:        pmw,
		fieldInParent: fd,
	}

	return leafNode, nil
}

func (pmw *protoMessageWrapper) virtualField(children *propSet) *virtualProtoField {
	return &virtualProtoField{
		msg:      pmw,
		children: children,
	}
}
