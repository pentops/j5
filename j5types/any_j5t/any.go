package any_j5t

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

func FromProto(msg proto.Message) (*Any, error) {
	protoVal, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return &Any{
		TypeName: string(msg.ProtoReflect().Descriptor().FullName()),
		Proto:    protoVal,
	}, nil
}

func (a *Any) UnmarshalTo(msg proto.Message) error {
	if a.TypeName != string(msg.ProtoReflect().Descriptor().FullName()) {
		return fmt.Errorf("type mismatch: %s != %s", a.TypeName, msg.ProtoReflect().Descriptor().FullName())
	}

	return proto.Unmarshal(a.Proto, msg)
}
