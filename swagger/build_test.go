package swagger

import (
	"encoding/json"
	"testing"

	"github.com/pentops/custom-proto-api/jsonapi"
	"github.com/pentops/o5-runtime-sidecar/testproto"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func filesToSwagger(t testing.TB, fileDescriptors ...*descriptorpb.FileDescriptorProto) *Document {
	t.Helper()
	services := testproto.FilesToServiceDescriptors(t, fileDescriptors...)
	swaggerDoc, err := Build(jsonapi.Options{
		ShortEnums: &jsonapi.ShortEnumsOption{},
	}, services)
	if err != nil {
		t.Fatal(err)
	}
	return swaggerDoc
}

const (
	pathMessage = 4
	pathField   = 2
)

func TestBuild(t *testing.T) {

	swaggerDoc := filesToSwagger(t, &descriptorpb.FileDescriptorProto{
		Name:    proto.String("test.proto"),
		Package: proto.String("test"),
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
				TypeName: proto.String(".test.Nested"),
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
				TypeName: proto.String(".test.TestEnum"),
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
	})

	bb, err := json.MarshalIndent(swaggerDoc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bb))

	if _, ok := swaggerDoc.GetSchema("#/components/schemas/test.TestRequest"); ok {
		t.Fatal("TestRequest should not be registered as a schema, but was")
	}

	refSchema, ok := swaggerDoc.GetSchema("#/components/schemas/test.Nested")
	if !ok {
		t.Fatal("schema not found")
	}

	if tn := refSchema.ItemType.TypeName(); tn != "object" {
		t.Fatalf("unexpected type: %s", tn)
	}

	if refSchema.Description != "Message Comment" {
		t.Errorf("unexpected description: '%s'", refSchema.Description)
	}

	asObject := refSchema.ItemType.(jsonapi.ObjectItem)
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
	enumType, ok := fEnum.SchemaItem.ItemType.(jsonapi.EnumItem)
	if !ok {
		t.Fatalf("unexpected type: %T", fEnum.SchemaItem.ItemType)
	}

	if enumType.Enum[0] != "UNSPECIFIED" {
		t.Errorf("unexpected enum value: '%s'", enumType.Enum[0])
	}
	if enumType.Enum[1] != "FOO" {
		t.Errorf("unexpected enum value: '%s'", enumType.Enum[1])
	}

}
