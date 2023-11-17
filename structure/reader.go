package structure

import (
	"io"
	"os"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func ReadFileDescriptorSet(src string) (*descriptorpb.FileDescriptorSet, error) {
	descriptors := &descriptorpb.FileDescriptorSet{}

	if src == "-" {
		protoData, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
		if err := proto.Unmarshal(protoData, descriptors); err != nil {
			return nil, err
		}
	} else {
		protoData, err := os.ReadFile(src)
		if err != nil {
			return nil, err
		}
		if err := proto.Unmarshal(protoData, descriptors); err != nil {
			return nil, err
		}
	}

	return descriptors, nil
}
