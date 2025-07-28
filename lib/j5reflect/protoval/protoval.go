package protoval

import (
	"sync"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type Value interface {
	IsSet() bool
	GetValue() (protoreflect.Value, bool)
	SetValue(protoreflect.Value) error
	Create()

	AsMessage() (MessageValue, bool)
	AsList() (ListValue, bool)
	AsMap() (MapValue, bool)

	FieldDescriptor() (protoreflect.FieldDescriptor, bool)
}

type MessageValue interface {
	Value
	MaybeMessageValue() (protoreflect.Message, bool)
	MessageValue() protoreflect.Message
	ChildField(protoreflect.FieldDescriptor) (Value, error)

	MessageDescriptor() protoreflect.MessageDescriptor
}

func NewRootMessageValue(msg protoreflect.Message) MessageValue {
	desc := msg.Descriptor()
	return newProtoContextMessage(newMessageValue(msg), desc)
}

type unimplementedValueTypes struct{}

func (unimplementedValueTypes) AsMessage() (MessageValue, bool) {
	return nil, true
}

func (unimplementedValueTypes) AsList() (ListValue, bool) {
	return nil, false
}

func (unimplementedValueTypes) AsMap() (MapValue, bool) {
	return nil, false
}

type valueWrapper struct {
	valueContext
}

func newProtoImpl(context valueContext) *valueWrapper {
	return &valueWrapper{
		valueContext: context,
	}
}

func (pi *valueWrapper) AsMessage() (MessageValue, bool) {
	desc, ok := pi.MessageDescriptor()
	if !ok {
		return nil, false
	}
	return newProtoContextMessage(pi, desc), true
}

func (pi *valueWrapper) AsList() (ListValue, bool) {
	fieldDesc, ok := pi.FieldDescriptor()
	if !ok {
		return nil, false
	}
	if !fieldDesc.IsList() {
		return nil, false
	}

	return newListWrapper(pi), true
}

func (pi *valueWrapper) AsMap() (MapValue, bool) {
	fieldDesc, ok := pi.FieldDescriptor()
	if !ok {
		return nil, false
	}
	if !fieldDesc.IsMap() {
		return nil, false
	}

	return newMapWrapper(pi), true
}

type protoContextMessage struct {
	valueContext
	descriptor protoreflect.MessageDescriptor
	unimplementedValueTypes
}

func newProtoContextMessage(context valueContext, descriptor protoreflect.MessageDescriptor) *protoContextMessage {
	return &protoContextMessage{
		valueContext: context,
		descriptor:   descriptor,
	}
}

func (pcm *protoContextMessage) MessageDescriptor() protoreflect.MessageDescriptor {
	return pcm.descriptor
}

func (pcm *protoContextMessage) MaybeMessageValue() (protoreflect.Message, bool) {
	val, ok := pcm.GetValue()
	if !ok {
		return nil, false
	}
	return val.Message(), true
}

func (pcm *protoContextMessage) MessageValue() protoreflect.Message {
	val, ok := pcm.GetValue()
	if ok {
		return val.Message()
	}

	mutableValue := pcm.newMutable()
	return mutableValue.Message()
}

func (pcm *protoContextMessage) ChildField(field protoreflect.FieldDescriptor) (Value, error) {
	childValue := newProtoPair(pcm, field)
	return newProtoImpl(childValue), nil
}

func (pcm *protoContextMessage) AsMessage() (MessageValue, bool) {
	return pcm, true
}

type ListValue interface {
	Value
	ItemMessageDescriptor() (protoreflect.MessageDescriptor, bool)
	ItemFieldDescriptor() protoreflect.FieldDescriptor
	Len() int
	Truncate(newLen int)
	AppendMessage() (MessageValue, int)
	AppendValue(value protoreflect.Value) (Value, int)
	ValueAt(idx int) (Value, bool)
}

type listWrapper struct {
	valueContext
	unimplementedValueTypes
	lock sync.Mutex
}

var _ ListValue = (*listWrapper)(nil)

func newListWrapper(context valueContext) *listWrapper {
	return &listWrapper{
		valueContext: context,
	}
}

func (lw *listWrapper) ItemFieldDescriptor() protoreflect.FieldDescriptor {
	fieldDesc, ok := lw.FieldDescriptor()
	if !ok {
		panic("listWrapper called ItemFieldDescriptor without a valid field descriptor")
	}
	return fieldDesc
}

func (lw *listWrapper) ItemMessageDescriptor() (protoreflect.MessageDescriptor, bool) {
	fieldDesc, ok := lw.FieldDescriptor()
	if !ok {
		return nil, false
	}
	if !fieldDesc.IsList() {
		// by now this should be impossible
		return nil, false
	}
	itemDesc := fieldDesc.Message()
	if itemDesc == nil {
		return nil, false
	}
	return itemDesc, true
}

func (lw *listWrapper) Len() int {
	val, ok := lw.GetValue()
	if !ok {
		return 0
	}
	if !val.IsValid() {
		return 0
	}
	return val.List().Len()
}

func (lw *listWrapper) Truncate(newLen int) {
	val, ok := lw.GetValue()
	if !ok {
		return
	}
	if !val.IsValid() {
		return
	}
	lw.lock.Lock()
	defer lw.lock.Unlock()
	list := val.List()
	if newLen < 0 || newLen > list.Len() {
		return
	}
	list.Truncate(newLen)
}

func (lw *listWrapper) AppendMessage() (MessageValue, int) {
	lw.lock.Lock()
	defer lw.lock.Unlock()
	val := lw.newMutable().List()
	idx := val.Len()
	elem := val.AppendMutable().Message()
	return NewRootMessageValue(elem), idx
}

func (lw *listWrapper) AppendValue(value protoreflect.Value) (Value, int) {
	lw.lock.Lock()
	defer lw.lock.Unlock()
	val := lw.newMutable().List()
	idx := val.Len()
	val.Append(value)
	item := &protoListValue{
		list:   val,
		index:  idx,
		parent: lw,
	}
	return newProtoImpl(item), idx
}

func (lw *listWrapper) ValueAt(idx int) (Value, bool) {
	lw.lock.Lock()
	defer lw.lock.Unlock()
	val := lw.newMutable().List()
	if idx < 0 || idx >= val.Len() {
		return nil, false
	}
	item := &protoListValue{
		list:   val,
		index:  idx,
		parent: lw,
	}
	return newProtoImpl(item), true
}

type MapValue interface {
	Value
	ItemMessageDescriptor() (protoreflect.MessageDescriptor, bool)
	ItemFieldDescriptor() protoreflect.FieldDescriptor
	ValueAt(key string) Value
	RangeValues(cb func(key string, value Value) bool)
}

type mapWrapper struct {
	valueContext
	unimplementedValueTypes
}

var _ MapValue = (*mapWrapper)(nil)

func newMapWrapper(context valueContext) *mapWrapper {
	return &mapWrapper{
		valueContext: context,
	}
}

func (mw *mapWrapper) ItemFieldDescriptor() protoreflect.FieldDescriptor {
	fieldDesc, ok := mw.FieldDescriptor()
	if !ok {
		panic("mapWrapper called ItemFieldDescriptor without a valid field descriptor")
	}
	return fieldDesc.MapValue()
}

func (mw *mapWrapper) ItemMessageDescriptor() (protoreflect.MessageDescriptor, bool) {
	fieldDesc, ok := mw.FieldDescriptor()
	if !ok {
		return nil, false
	}
	if !fieldDesc.IsMap() {
		// by now this should be impossible
		return nil, false
	}
	itemDesc := fieldDesc.MapValue().Message()
	if itemDesc == nil {
		return nil, false
	}
	return itemDesc, true
}

func (mw *mapWrapper) ValueAt(key string) Value {
	val := mw.newMutable().Map()
	item := &protoMapValue{
		mapValue: val,
		parent:   mw,
		key:      protoreflect.MapKey(protoreflect.ValueOfString(key)),
	}

	return newProtoImpl(item)
}

func (mw *mapWrapper) RangeValues(cb func(key string, value Value) bool) {
	parentMap := mw.newMutable().Map()
	parentMap.Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
		keyStr := key.Value().String()
		item := &protoMapValue{
			mapValue: parentMap,
			parent:   mw,
			key:      key,
		}
		return cb(keyStr, newProtoImpl(item))
	})
}
