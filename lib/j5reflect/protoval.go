package j5reflect

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type protoContext interface {
	isSet() bool

	getValue() (protoreflect.Value, bool)
	getMutableValue(createIfNotSet bool) (protoreflect.Value, error)
	setValue(protoreflect.Value) error
}

// protoPair is a field within a message. The message will exist, but the field
// may be empty/unset/nil
type protoPair struct {
	parentMessage protoreflect.Message
	fieldInParent protoreflect.FieldDescriptor
	isSubset      bool // for a 'virtual' oneof, this is true, as the message is shared with the 'parent' of the oneof.
}

var _ protoContext = (*protoPair)(nil)

func newProtoPair(msg protoreflect.Message, field protoreflect.FieldDescriptor) *protoPair {
	if msg == nil || !msg.IsValid() {
		panic("msg is nil/invalid")
	}
	if field == nil {
		panic("field is nil")
	}
	return &protoPair{
		parentMessage: msg,
		fieldInParent: field,
		isSubset:      false, // the only way to get this is to use cloneForSubset.
	}
}

func (pp *protoPair) isSet() bool {
	return pp.parentMessage.Has(pp.fieldInParent)
}

func (pp *protoPair) getValue() (protoreflect.Value, bool) {
	if !pp.isSet() {
		return protoreflect.Value{}, false
	}
	val := pp.parentMessage.Get(pp.fieldInParent)
	if !val.IsValid() {
		return protoreflect.Value{}, false
	}
	return val, true
}

func (pp *protoPair) setValue(val protoreflect.Value) error {
	if !val.IsValid() {
		pp.parentMessage.Clear(pp.fieldInParent)
		return nil
	}
	pp.parentMessage.Set(pp.fieldInParent, val)
	return nil
}

func (pp *protoPair) fieldDescriptor() protoreflect.FieldDescriptor {
	return pp.fieldInParent
}

func (pp *protoPair) getMutableValue(createIfNotSet bool) (protoreflect.Value, error) {
	if !pp.isSet() {
		if !createIfNotSet {
			return protoreflect.Value{}, fmt.Errorf("field %s is not set", pp.fieldInParent.FullName())
		}
	}
	return pp.parentMessage.Mutable(pp.fieldInParent), nil
}
