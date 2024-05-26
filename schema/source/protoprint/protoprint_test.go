package protoprint

import (
	"os"
	"strings"
	"testing"

	"github.com/pentops/jsonapi/gen/test/foo/v1/foo_testpb"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestSimplePrint(t *testing.T) {

	input := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("test.proto"),
		Syntax:  proto.String("proto3"),
		Package: proto.String("test.v1"),
		Dependency: []string{
			"google/protobuf/empty.proto",
			"google/protobuf/timestamp.proto",
		},
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String("go_package_value"),
		},
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String("Test"),

			Field: []*descriptorpb.FieldDescriptorProto{{
				Name:   proto.String("f1"),
				Number: proto.Int32(1),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
			}, {
				Name:           proto.String("f2"),
				Number:         proto.Int32(2),
				Label:          descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				Type:           descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
				Proto3Optional: proto.Bool(true),
			}, {
				Name:   proto.String("f3"),
				Number: proto.Int32(3),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
			}, {
				Name:     proto.String("f4"),
				Number:   proto.Int32(4),
				Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				TypeName: proto.String(".google.protobuf.Empty"),
				Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
			}, {
				Name:     proto.String("f5"),
				Number:   proto.Int32(5),
				Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
				Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
				TypeName: proto.String(".test.v1.Test.F5Entry"),
			}},

			NestedType: []*descriptorpb.DescriptorProto{{
				Name: proto.String("F5Entry"),
				Field: []*descriptorpb.FieldDescriptorProto{{
					Name:   proto.String("key"),
					Number: proto.Int32(1),
					Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				}, {
					Name:   proto.String("value"),
					Number: proto.Int32(2),
					Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				}},
				Options: &descriptorpb.MessageOptions{
					MapEntry: proto.Bool(true),
				},
			}},
		}},

		Service: []*descriptorpb.ServiceDescriptorProto{{
			Name: proto.String("TestService"),
			Method: []*descriptorpb.MethodDescriptorProto{{
				Name:       proto.String("TestMethod"),
				InputType:  proto.String(".test.v1.Test"),
				OutputType: proto.String("google.protobuf.Empty"),
				Options:    &descriptorpb.MethodOptions{},
			}},
		}},
	}

	proto.SetExtension(input.Service[0].Method[0].Options, annotations.E_Http, &annotations.HttpRule{
		Pattern: &annotations.HttpRule_Post{
			Post: "/test/v1/foo",
		},
		Body: "*",
	})

	output, err := printFile(input)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(prototext.Format(input))

	expected := []string{
		`syntax = "proto3";`,
		"",
		`package test.v1;`,
		"",
		`import "google/protobuf/empty.proto";`,
		`import "google/protobuf/timestamp.proto";`,
		``,
		`option go_package = "go_package_value";`,
		"",
		`service TestService {`,
		`  rpc TestMethod(Test) returns (google.protobuf.Empty) {`,
		`    option (google.api.http) = {`,
		`      post: "/test/v1/foo"`,
		`      body: "*"`,
		`    };`,
		`  }`,
		`}`,
		"",
		`message Test {`,
		`  string f1 = 1;`,
		`  optional string f2 = 2;`,
		`  repeated string f3 = 3;`,
		`  google.protobuf.Empty f4 = 4;`,
		`  map<string, string> f5 = 5;`,
		`}`,
		"",
	}

	outputLines := strings.Split(string(output), "\n")
	assertEqualLines(t, expected, outputLines)
}

func assertEqualLines(t *testing.T, wantLines, gotLines []string) {

	for idx, line := range gotLines {
		t.Logf("got %03d: '%s'", idx, line)
		if idx >= len(wantLines) {
			t.Errorf("   EXTRA")
			continue
		}
		if wantLines[idx] != line {
			t.Errorf("   want: '%s'", wantLines[idx])
		}
	}
}

func TestExampleReal(t *testing.T) {
	inputDesc := (&foo_testpb.GetFooRequest{}).ProtoReflect().Descriptor().ParentFile()

	inputMsg := protodesc.ToFileDescriptorProto(inputDesc)

	t.Logf(prototext.Format(inputMsg))

	output, err := printFile(inputMsg)
	if err != nil {
		t.Fatal(err)
	}

	realFile, err := os.ReadFile("../../../proto/test/test/foo/v1/test.proto")
	if err != nil {
		t.Fatal(err)
	}

	realLines := strings.Split(string(realFile), "\n")
	gotLines := strings.Split(string(output), "\n")

	assertEqualLines(t, realLines, gotLines)

}
