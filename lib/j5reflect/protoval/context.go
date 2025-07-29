package protoval

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type valueContext interface {
	IsSet() bool
	Create()
	GetValue() (protoreflect.Value, bool)
	SetValue(protoreflect.Value) error
	newMutable() protoreflect.Value

	MessageDescriptor() (protoreflect.MessageDescriptor, bool)
	FieldDescriptor() (protoreflect.FieldDescriptor, bool)
}

type parentContext interface {
	MaybeMessageValue() (protoreflect.Message, bool)
	MessageValue() protoreflect.Message
}

// protoPair is a field in a message. Fields in proto can't exist on their own.
// The field can be any type, including a scalar, message, list or map.
type protoPair struct {
	parentMessage parentContext
	fieldInParent protoreflect.FieldDescriptor
}

var _ valueContext = (*protoPair)(nil)

func newProtoPair(msg parentContext, field protoreflect.FieldDescriptor) *protoPair {
	return &protoPair{
		parentMessage: msg,
		fieldInParent: field,
	}
}

func (pp *protoPair) IsSet() bool {
	pm, ok := pp.parentMessage.MaybeMessageValue()
	if !ok {
		return false
	}
	if !pm.Has(pp.fieldInParent) {
		return false
	}
	val := pm.Get(pp.fieldInParent)
	return val.IsValid()
}

func (pp *protoPair) Create() {
	mv := pp.parentMessage.MessageValue()
	if mv.Has(pp.fieldInParent) {
		return
	}
	defaultValue := mv.NewField(pp.fieldInParent)
	mv.Set(pp.fieldInParent, defaultValue)
}

func (pp *protoPair) GetValue() (protoreflect.Value, bool) {
	pm, ok := pp.parentMessage.MaybeMessageValue()
	if !ok {
		return protoreflect.Value{}, false
	}
	if !pm.Has(pp.fieldInParent) {
		return protoreflect.Value{}, false
	}
	val := pm.Get(pp.fieldInParent)
	if !val.IsValid() {
		return protoreflect.Value{}, false
	}
	return val, true
}

func (pp *protoPair) SetValue(val protoreflect.Value) error {
	pm := pp.parentMessage.MessageValue()
	if !val.IsValid() {
		pm.Clear(pp.fieldInParent)
		return nil
	}
	pm.Set(pp.fieldInParent, val)
	return nil
}

func (pp *protoPair) newMutable() protoreflect.Value {
	pm := pp.parentMessage.MessageValue()
	if !pm.IsValid() {
		panic("protoContextMessage parent is invalid")
	}
	val := pm.Mutable(pp.fieldInParent)
	if !val.IsValid() {
		panic("mutable field returned invalid value")
	}
	return val
}

func (pp *protoPair) FieldDescriptor() (protoreflect.FieldDescriptor, bool) {
	return pp.fieldInParent, true
}

func (pp *protoPair) MessageDescriptor() (protoreflect.MessageDescriptor, bool) {
	if pp.fieldInParent.Kind() != protoreflect.MessageKind {
		return nil, false
	}
	if pp.fieldInParent.Message() == nil {
		return nil, false
	}
	return pp.fieldInParent.Message(), true
}

// messageValue represents a concrete message which already exists, rather than
// as a property in a recursive parent.
type messageValue struct {
	value      protoreflect.Message
	descriptor protoreflect.MessageDescriptor
}

var _ valueContext = (*messageValue)(nil)

func newMessageValue(value protoreflect.Message) *messageValue {
	return &messageValue{
		value:      value,
		descriptor: value.Descriptor(),
	}
}
func (mv *messageValue) IsSet() bool {
	return true
}

func (mv *messageValue) Create() {
	// concrete value already exists
}

func (mv *messageValue) GetValue() (protoreflect.Value, bool) {
	return protoreflect.ValueOfMessage(mv.value), true
}

func (mv *messageValue) SetValue(val protoreflect.Value) error {
	return fmt.Errorf("cannot set a value to a message value, use setMessage instead")
}

func (mv *messageValue) newMutable() protoreflect.Value {
	if !mv.value.IsValid() {
		panic("messageValue is invalid")
	}
	return protoreflect.ValueOfMessage(mv.value)
}

func (mv *messageValue) MessageDescriptor() (protoreflect.MessageDescriptor, bool) {
	return mv.descriptor, true
}

func (mv *messageValue) FieldDescriptor() (protoreflect.FieldDescriptor, bool) {
	return nil, false
}

/*

// protoMapValue represents the value under a key of a map.
type protoMapValue struct {
	mapVal protoreflect.Map
	key    protoreflect.MapKey

	unimplementedProtoContext
}

var _ protoContext = (*protoMapValue)(nil)

func (pmv *protoMapValue) isSet() bool {
	_, ok := pmv.getValue()
	return ok
}

func (pmv *protoMapValue) setValue(val protoreflect.Value) error {
	if !val.IsValid() {
		pmv.mapVal.Clear(pmv.key)
		return nil
	}
	pmv.mapVal.Set(pmv.key, val)
	return nil
}

func (pmv *protoMapValue) getValue() (protoreflect.Value, bool) {
	itemVal := pmv.mapVal.Get(pmv.key)
	return itemVal, itemVal.IsValid()
}

func (pmv *protoMapValue) setDefaultValue() error {
	return nil
}

func (pmv *protoMapValue) getMutableValue(createIfNotSet bool) (protoMessageContext, error) {
	return newMessageValue(pmv.mapVal.Get(pmv.key).Message()), nil
}

*/

// protoListValue wraps a scalar/leaf type array, keeping pointer to the parent
// and the location within the parent where the object exists to make it
// semi-mutable.
type protoListValue struct {
	list   protoreflect.List
	parent valueContext

	//parentField protoreflect.FieldDescriptor
	index int
}

var _ valueContext = (*protoListValue)(nil)

func (plv *protoListValue) IsSet() bool {
	_, ok := plv.GetValue()
	return ok
}

func (plv *protoListValue) Create() {
	// list values are always set, so no need to create
}

func (plv *protoListValue) SetValue(val protoreflect.Value) error {
	if !val.IsValid() {
		return fmt.Errorf("cannot set a nil value to a list val")
	}
	plv.list.Set(plv.index, val)
	return nil
}

func (plv *protoListValue) GetValue() (protoreflect.Value, bool) {
	itemVal := plv.list.Get(plv.index)
	return itemVal, itemVal.IsValid()
}

func (plv *protoListValue) newMutable() protoreflect.Value {
	return plv.list.Get(plv.index)
}

func (plv *protoListValue) FieldDescriptor() (protoreflect.FieldDescriptor, bool) {
	return plv.parent.FieldDescriptor()
}

func (plv *protoListValue) MessageDescriptor() (protoreflect.MessageDescriptor, bool) {
	return plv.parent.MessageDescriptor()
}

// protoMapValue represents a value in a map, which is a key-value pair.
type protoMapValue struct {
	mapValue protoreflect.Map
	parent   valueContext
	key      protoreflect.MapKey
}

var _ valueContext = (*protoMapValue)(nil)

func (pmv *protoMapValue) IsSet() bool {
	_, ok := pmv.GetValue()
	return ok
}

func (pmv *protoMapValue) Create() {
	// map values are always set, so no need to create
}

func (pmv *protoMapValue) SetValue(val protoreflect.Value) error {
	if !val.IsValid() {
		return fmt.Errorf("cannot set a nil value to a map val")
	}
	pmv.mapValue.Set(pmv.key, val)
	return nil
}

func (pmv *protoMapValue) GetValue() (protoreflect.Value, bool) {
	itemVal := pmv.mapValue.Get(pmv.key)
	return itemVal, itemVal.IsValid()
}

func (pmv *protoMapValue) newMutable() protoreflect.Value {
	if !pmv.mapValue.Has(pmv.key) {
		pmv.mapValue.Set(pmv.key, pmv.mapValue.NewValue())
	}
	return pmv.mapValue.Get(pmv.key)
}

func (pmv *protoMapValue) FieldDescriptor() (protoreflect.FieldDescriptor, bool) {
	return pmv.parent.FieldDescriptor()
}

func (pmv *protoMapValue) MessageDescriptor() (protoreflect.MessageDescriptor, bool) {
	return pmv.parent.MessageDescriptor()
}
