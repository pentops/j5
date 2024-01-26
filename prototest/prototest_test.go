package prototest

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestSingletonDescriptor(t *testing.T) {

	// When parsing the entire file set, the descriptors for imported types are
	// equivalent but not equal to the descriptors for the same types in the
	// global registry.

	// This test ensures, through the string wrapper as an example, that a
	// directly constructed wrapperspb.StringValue is the same as one constructed
	// via the reflection on the on-the-fly parsed descriptor.

	runTest := func(t *testing.T, msgDesc protoreflect.MessageDescriptor) {
		stringFieldDesc := msgDesc.Fields().ByName("string")

		wrapperVal := wrapperspb.String("value")

		stringValDynamic := dynamicpb.NewMessage(stringFieldDesc.Message())
		stringValDynamic.Set(stringValDynamic.Descriptor().Fields().ByName("value"), protoreflect.ValueOfString("value"))

		if !proto.Equal(stringValDynamic, wrapperVal) {
			t.Errorf("dynamic and concrete not equal")
		}

		if stringValDynamic.ProtoReflect().Descriptor() != wrapperVal.ProtoReflect().Descriptor() {
			t.Fatal("descriptors of the two values not equal")
		}

	}

	t.Run("target", func(t *testing.T) {
		// This test sets the target baseline, proving the test is possible
		// (which of course it is now that it passes, but that wasn't always as obvious)
		fileDesc, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
			Name:    proto.String("test.proto"),
			Package: proto.String("test"),
			Syntax:  proto.String("proto3"),
			Dependency: []string{
				"google/protobuf/wrappers.proto",
			},
			MessageType: []*descriptorpb.DescriptorProto{{
				Name: proto.String("Wrapper"),
				Field: []*descriptorpb.FieldDescriptorProto{{
					Name:     proto.String("string"),
					Number:   proto.Int32(1),
					Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
					TypeName: proto.String(".google.protobuf.StringValue"),
				}},
			}},
		}, protoregistry.GlobalFiles)

		if err != nil {
			t.Fatal(err)
		}

		msgDesc := fileDesc.Messages().ByName("Wrapper")
		if msgDesc == nil {
			t.Fatal("no foo message")
		}
		runTest(t, msgDesc)

	})

	t.Run("using library", func(t *testing.T) {

		rs := DescriptorsFromSource(t, map[string]string{
			"test.proto": `
		syntax = "proto3";

		package test;

		import "google/protobuf/wrappers.proto";

		message Wrapper {
			google.protobuf.StringValue string = 1;
		}
		`,
		})
		msgDesc := rs.MessageByName(t, "test.Wrapper")
		runTest(t, msgDesc)

	})

}
