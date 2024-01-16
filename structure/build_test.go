package structure

import (
	"strings"
	"testing"

	"github.com/pentops/jsonapi/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/jsonapi/gen/j5/v1/schema_j5pb"
	"github.com/pentops/o5-runtime-sidecar/testproto"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	pathMessage = 4
	pathField   = 2
)

func TestBuild(t *testing.T) {

	descriptors := &descriptorpb.FileDescriptorProto{
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String("github.com/pentops/jsonapi/test_pb"),
		},
		Name:    proto.String("test.proto"),
		Package: proto.String("test.v1"),
		Service: []*descriptorpb.ServiceDescriptorProto{{
			Name: proto.String("TestService"),
			Method: []*descriptorpb.MethodDescriptorProto{
				testproto.BuildHTTPMethod("Test", &annotations.HttpRule{
					Pattern: &annotations.HttpRule_Get{
						Get: "/test/{test_field}",
					},
				}),
			},
		}},
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String("TestRequest"),
			Field: []*descriptorpb.FieldDescriptorProto{{
				Name:   proto.String("test_field"),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
				Number: proto.Int32(1),
			}},
		}, {
			Name: proto.String("TestResponse"),
			Field: []*descriptorpb.FieldDescriptorProto{{
				Name:   proto.String("test_field"),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
				Number: proto.Int32(1),
			}, {
				Name:     proto.String("msg"),
				Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
				Number:   proto.Int32(2),
				TypeName: proto.String(".test.v1.Nested"),
			}},
		}, {
			Name: proto.String("Nested"),
			Field: []*descriptorpb.FieldDescriptorProto{{
				Name:   proto.String("nested_field"),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
				Number: proto.Int32(1),
			}, {
				Name:     proto.String("enum"),
				Type:     descriptorpb.FieldDescriptorProto_TYPE_ENUM.Enum(),
				Number:   proto.Int32(3),
				TypeName: proto.String(".test.v1.TestEnum"),
			}},
		}},
		EnumType: []*descriptorpb.EnumDescriptorProto{{
			Name: proto.String("TestEnum"),
			Value: []*descriptorpb.EnumValueDescriptorProto{{
				Name:   proto.String("TEST_ENUM_UNSPECIFIED"),
				Number: proto.Int32(0),
			}, {
				Name:   proto.String("TEST_ENUM_FOO"),
				Number: proto.Int32(1),
			}},
		}},

		SourceCodeInfo: &descriptorpb.SourceCodeInfo{
			Location: []*descriptorpb.SourceCodeInfo_Location{{
				LeadingComments: proto.String("Message Comment"),
				Path:            []int32{pathMessage, 2}, // 4 = Message, 2 = Nested
				Span:            []int32{1, 1, 1},        // Single line comment
			}, {
				LeadingComments: proto.String("Field Comment"),
				Path:            []int32{pathMessage, 2, pathField, 0}, // 4 = Message, 2 = Nested, 1 = Field
				Span:            []int32{2, 1, 2},                      // Single line comment
			}},
		},
	}

	want := &schema_j5pb.API{
		Packages: []*schema_j5pb.Package{{
			Label: "Test",
			Name:  "test.v1",
			Methods: []*schema_j5pb.Method{{
				GrpcServiceName: "TestService",
				FullGrpcName:    "/test.v1.TestService/Test",
				GrpcMethodName:  "Test",
				HttpMethod:      "get",
				HttpPath:        "/test/:testField",
				ResponseBody: &schema_j5pb.Schema{
					Type: &schema_j5pb.Schema_ObjectItem{
						ObjectItem: &schema_j5pb.ObjectItem{
							ProtoFullName:    "test.v1.TestResponse",
							ProtoMessageName: "TestResponse",
							//Rules:            &schema_j5pb.ObjectRules{},
							GoPackageName:   "github.com/pentops/jsonapi/test_pb",
							GoTypeName:      "TestResponse",
							GrpcPackageName: "test.v1",
							Properties: []*schema_j5pb.ObjectProperty{{
								Name:               "testField",
								Description:        "",
								ProtoFieldName:     "test_field",
								ProtoFieldNumber:   1,
								ExplicitlyOptional: true,
								Required:           false,
								WriteOnly:          false,
								Schema: &schema_j5pb.Schema{
									Description: "",
									Type: &schema_j5pb.Schema_StringItem{
										StringItem: &schema_j5pb.StringItem{},
									},
								},
							}, {
								Name:               "msg",
								Description:        "",
								ProtoFieldName:     "msg",
								ProtoFieldNumber:   2,
								Required:           false,
								ExplicitlyOptional: true,
								Schema: &schema_j5pb.Schema{
									Type: &schema_j5pb.Schema_Ref{
										Ref: "test.v1.Nested",
									},
								},
							}},
						},
					},
				},
				PathParameters: []*schema_j5pb.Parameter{{
					Name:        "testField",
					Description: "",
					Required:    true,
					Schema: &schema_j5pb.Schema{
						Type: &schema_j5pb.Schema_StringItem{
							StringItem: &schema_j5pb.StringItem{},
						},
					},
				}},
			}},
		}},
	}

	built, err := BuildFromDescriptors(&source_j5pb.Config{
		Packages: []*source_j5pb.PackageConfig{{
			Label: "Test",
			Name:  "test.v1",
		}},
		Options: &source_j5pb.CodecOptions{
			ShortEnums: &source_j5pb.ShortEnumOptions{},
		},
	}, &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{descriptors},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Packages are controlled by this package, should equal in full. Schema
	// tests are in the jsonapi package.

	assertEqualProto(t, want.Packages[0], built.Packages[0])

	if _, ok := built.Schemas["test.v1.TestRequest"]; ok {
		t.Fatal("TestRequest should not be registered as a schema, but was")
	}

	refSchema, ok := built.Schemas["test.v1.Nested"]
	if !ok {
		t.Fatal("schema not found")
	}

	if refSchema.Description != "Message Comment" {
		t.Errorf("unexpected description: '%s'", refSchema.Description)
	}

	asObject := refSchema.GetObjectItem()
	if asObject == nil {
		t.Fatal("schema is not an object")
	}
	if len(asObject.Properties) != 2 {
		t.Fatalf("unexpected properties: %d", len(asObject.Properties))
	}

	f1 := asObject.Properties[0]
	if f1.Name != "nestedField" {
		t.Errorf("unexpected field name: '%s'", f1.Name)
	}

	if f1.Description != "Field Comment" {
		t.Errorf("unexpected description: '%s'", f1.Description)
	}

	fEnum := asObject.Properties[1]
	if fEnum.Name != "enum" {
		t.Errorf("unexpected field name: '%s'", fEnum.Name)
	}

	ref := fEnum.Schema.GetRef()
	if ref == "" {
		t.Fatal("ref is nil")
	}

	schemaEnum, ok := built.Schemas[strings.TrimPrefix(ref, "#/components/schemas/")]
	if !ok {
		t.Fatalf("schema not found: '%s'", ref)
	}

	enumType := schemaEnum.GetEnumItem()
	if enumType == nil {
		t.Fatalf("unexpected type: %T", fEnum.Schema.Type)
	}

	if enumType.Options[0].Name != "UNSPECIFIED" {
		t.Errorf("unexpected enum value: '%s'", enumType.Options[0])
	}
	if enumType.Options[1].Name != "FOO" {
		t.Errorf("unexpected enum value: '%s'", enumType.Options[1])
	}

}

func assertEqualProto(t *testing.T, want, got proto.Message) {
	t.Helper()
	wantJSON := protojson.Format(want)
	gotJSON := protojson.Format(got)
	if string(wantJSON) == string(gotJSON) {
		t.Log("STRINGS MATCH")
	}

	matched := true

	lineA := 0
	lineB := 0

	wantLines := strings.Split(string(wantJSON), "\n")
	gotLines := strings.Split(string(gotJSON), "\n")
	for {
		if lineA >= len(wantLines) || lineB >= len(gotLines) {
			break
		}
		wantLine := string(wantLines[lineA])
		gotLine := string(gotLines[lineB])
		if wantLine != gotLine {
			matched = false
			t.Logf("  !W: %s", wantLine)
			t.Logf("  !G: %s", gotLine)
			t.Log(strings.Repeat(`/\`, 10))
		} else {
			t.Logf("   OK: %s", wantLine)
		}
		lineA++
		lineB++
	}
	if lineA < len(wantLines) {
		matched = false
		t.Logf("  !A: %s", strings.Join(wantLines[lineA:], "\n"))
	}
	if lineB < len(gotLines) {
		matched = false
		t.Logf("  !B: %s", strings.Join(gotLines[lineB:], "\n"))
	}

	if !matched {
		t.Errorf("unexpected JSON")
	}

}
