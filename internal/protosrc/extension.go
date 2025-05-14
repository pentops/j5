package protosrc

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// GetExtension is equivalent to proto.GetExtension, but it works with
// dynamicpb.Message, required sometimes when using buf's protocompile linker
// (or our misuse of it here)
func GetExtension[T proto.Message](elem protoreflect.ProtoMessage, xt protoreflect.ExtensionType) (res T) {
	td := elem.ProtoReflect().Get(xt.TypeDescriptor())
	if xt.IsValidValue(td) {
		ext := xt.InterfaceOf(td)
		return ext.(T)
	}

	dynamicExt := td.Interface().(*dynamicpb.Message)

	mm, err := proto.Marshal(dynamicExt)
	if err != nil {
		panic(err)
	}

	rr := (*new(T)).ProtoReflect().New().Interface().(T)
	err = proto.Unmarshal(mm, rr)
	if err != nil {
		panic(err)
	}
	return rr
}
