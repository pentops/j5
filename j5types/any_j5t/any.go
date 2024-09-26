package any_j5t

import "google.golang.org/protobuf/proto"

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
