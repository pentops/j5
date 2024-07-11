package structure

import (
	"strings"
	"testing"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/pentops/flowtest/prototest"
	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/internal/j5reflect"
	"github.com/stretchr/testify/assert"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func fieldWithValidateExtension(field *descriptorpb.FieldDescriptorProto, constraints *validate.FieldConstraints) *descriptorpb.FieldDescriptorProto {
	return fieldWithExtension(field, validate.E_Field, constraints)
}

func fieldWithExtension(field *descriptorpb.FieldDescriptorProto, extensionType protoreflect.ExtensionType, extensionValue interface{}) *descriptorpb.FieldDescriptorProto {
	if field.Options == nil {
		field.Options = &descriptorpb.FieldOptions{}
	}

	proto.SetExtension(field.Options, extensionType, extensionValue)
	return field
}

const (
	pathMessage = 4
	pathField   = 2
)

func testImage() *source_j5pb.SourceImage {
	return &source_j5pb.SourceImage{
		Packages: []*config_j5pb.PackageConfig{{
			Label: "Test",
			Name:  "test.v1",
		}},
		File: []*descriptorpb.FileDescriptorProto{{
			Syntax: proto.String("proto3"),
			Options: &descriptorpb.FileOptions{
				GoPackage: proto.String("github.com/pentops/j5/test_pb"),
			},
			Name:    proto.String("test.proto"),
			Package: proto.String("test.v1"),
			Service: []*descriptorpb.ServiceDescriptorProto{{
				Name: proto.String("TestService"),
				Method: []*descriptorpb.MethodDescriptorProto{
					prototest.BuildHTTPMethod("Test", &annotations.HttpRule{
						Pattern: &annotations.HttpRule_Get{
							Get: "/test/{path_field}",
						},
					}),
				},
			}},
			MessageType: []*descriptorpb.DescriptorProto{{
				Name: proto.String("TestRequest"),
				Field: []*descriptorpb.FieldDescriptorProto{
					fieldWithValidateExtension(&descriptorpb.FieldDescriptorProto{
						Name:   proto.String("path_field"),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Number: proto.Int32(1),
					}, &validate.FieldConstraints{
						Required: true,
					}), {
						Name:   proto.String("query_field"),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Number: proto.Int32(2),
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
					TypeName: proto.String(".test.v1.Referenced"),
				}},
			}, {
				Name: proto.String("Referenced"),
				Field: []*descriptorpb.FieldDescriptorProto{{
					Name:   proto.String("field_1"),
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
		}},
	}
}

func testAPI() *schema_j5pb.API {

	return &schema_j5pb.API{
		Packages: []*schema_j5pb.Package{{
			Label: "Test",
			Name:  "test.v1",
			Services: []*schema_j5pb.Service{{
				Name: "TestService",
				Methods: []*schema_j5pb.Method{{
					FullGrpcName: "/test.v1.TestService/Test",
					Name:         "Test",
					HttpMethod:   schema_j5pb.HTTPMethod_HTTP_METHOD_GET,
					HttpPath:     "/test/:pathField",
					ResponseBody: &schema_j5pb.RootSchema{
						Type: &schema_j5pb.RootSchema_Object{
							Object: &schema_j5pb.Object{
								Name: "TestResponse",
								//Rules:            &schema_j5pb.ObjectRules{},
								Properties: []*schema_j5pb.ObjectProperty{{
									Name:        "testField",
									Description: "",
									Required:    false,
									WriteOnly:   false,
									Schema: &schema_j5pb.Schema{
										Type: &schema_j5pb.Schema_String_{
											String_: &schema_j5pb.StringField{},
										},
									},
									ProtoField: []int32{1},
								}, {
									Name:        "msg",
									Description: "",
									Required:    false,
									Schema: &schema_j5pb.Schema{
										Type: &schema_j5pb.Schema_Object{
											Object: &schema_j5pb.ObjectField{
												Schema: &schema_j5pb.ObjectField_Ref{
													Ref: &schema_j5pb.Ref{
														Package: "test.v1",
														Schema:  "Referenced",
													},
												},
											},
										},
									},
									ProtoField: []int32{2},
								}},
							},
						},
					},
					Request: &schema_j5pb.Method_Request{
						PathParameters: []*schema_j5pb.ObjectProperty{{
							Name:        "pathField",
							Description: "",
							Required:    true,
							Schema: &schema_j5pb.Schema{
								Type: &schema_j5pb.Schema_String_{
									String_: &schema_j5pb.StringField{},
								},
							},
							ProtoField: []int32{1},
						}},
						QueryParameters: []*schema_j5pb.ObjectProperty{{
							Name:     "queryField",
							Required: false,
							Schema: &schema_j5pb.Schema{
								Type: &schema_j5pb.Schema_String_{
									String_: &schema_j5pb.StringField{},
								},
							},
							ProtoField: []int32{2},
						}},
					},
				}},
			}},
			Schemas: map[string]*schema_j5pb.RootSchema{
				"test.v1.Referenced": {
					Type: &schema_j5pb.RootSchema_Object{
						Object: &schema_j5pb.Object{
							Name:        "Referenced",
							Description: "Message Comment",
							Properties: []*schema_j5pb.ObjectProperty{{
								Name:        "field1",
								Description: "Field Comment",
								Schema: &schema_j5pb.Schema{
									Type: &schema_j5pb.Schema_String_{
										String_: &schema_j5pb.StringField{},
									},
								},
								ProtoField: []int32{1},
							}, {
								Name: "enum",
								Schema: &schema_j5pb.Schema{
									Type: &schema_j5pb.Schema_Enum{
										Enum: &schema_j5pb.EnumField{
											Schema: &schema_j5pb.EnumField_Ref{
												Ref: &schema_j5pb.Ref{
													Package: "test.v1",
													Schema:  "TestEnum",
												},
											},
										},
									},
								},
								ProtoField: []int32{3},
							}},
						},
					},
				},
				"test.v1.TestEnum": {
					Type: &schema_j5pb.RootSchema_Enum{
						Enum: &schema_j5pb.Enum{
							Name:   "TestEnum",
							Prefix: "TEST_ENUM_",
							Options: []*schema_j5pb.Enum_Value{{
								Name:   "UNSPECIFIED",
								Number: 0,
							}, {
								Name:   "FOO",
								Number: 1,
							}},
						},
					},
				},
			},
		}},
	}
}

func TestBuildPath(t *testing.T) {

	sourceImage := testImage()

	apiReflection, err := ReflectFromSource(sourceImage)
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Run("Reflect Direct", func(t *testing.T) {
		if len(apiReflection.Packages) != 1 {
			t.Fatalf("unexpected packages: %d", len(apiReflection.Packages))
		}
		pkg := apiReflection.Packages[0]
		assert.Equal(t, "test.v1", pkg.Name)
		if len(pkg.Services) != 1 {
			t.Fatalf("unexpected services: %d", len(pkg.Services))
		}
		service := pkg.Services[0]
		if len(service.Methods) != 1 {
			t.Fatalf("unexpected methods: %d", len(service.Methods))
		}
		method := service.Methods[0]

		resObject, ok := method.Response.(*j5reflect.ObjectSchema)
		if !ok {
			t.Fatalf("unexpected type: %T", resObject)
		}

		if len(resObject.Properties) != 2 {
			t.Fatalf("unexpected properties: %d", len(resObject.Properties))
		}

		prop := resObject.Properties[1]
		fieldSchema, ok := prop.Schema.(*j5reflect.ObjectFieldSchema)
		if !ok {
			t.Fatalf("unexpected type: %T", prop.Schema)
		}

		// The field is a ref, as all are in reflection.
		assert.Equal(t, "test.v1", fieldSchema.Ref.Package)
		assert.Equal(t, "Referenced", fieldSchema.Ref.Schema)

	})

	apiDescriptor, err := apiReflection.ToJ5Proto()
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Run("Source to Descriptor", func(t *testing.T) {
		assert.Len(t, apiDescriptor.Packages, 1)
		want := testAPI()
		want.Metadata = apiDescriptor.Metadata
		assertEqualProto(t, want, apiDescriptor)
	})

	t.Run("Re-Convert", func(t *testing.T) {

		reflectionFromBuilt, err := j5reflect.APIFromDesc(apiDescriptor)
		if err != nil {
			t.Fatal(err.Error())
		}
		assert.Len(t, reflectionFromBuilt.Packages, 1)
		builtDescriptor, err := reflectionFromBuilt.ToJ5Proto()
		if err != nil {
			t.Fatal(err.Error())
		}
		assert.Len(t, builtDescriptor.Packages, 1)
		builtDescriptor.Metadata = apiDescriptor.Metadata

		assertEqualProto(t, apiDescriptor, builtDescriptor)

	})

	t.Run("Specific Cases", func(t *testing.T) {

		// Packages are controlled by this package, should equal in full. Schema
		// tests are in the jsonapi package.
		schemas := apiDescriptor.Packages[0].Schemas
		assert.Equal(t, "test.v1", apiDescriptor.Packages[0].Name)

		for k := range schemas {
			t.Logf("schema: %s", k)
		}
		t.Logf("%d schemas", len(schemas))

		if _, ok := schemas["TestRequest"]; ok {
			t.Fatal("TestRequest should not be registered as a schema, but was")
		}

		refSchema, ok := schemas["test.v1.Referenced"]
		if !ok {
			t.Fatal("schema 'test.v1.Referenced' not found")
		}

		asObject := refSchema.GetObject()
		if asObject == nil {
			t.Fatalf("schema is not an object but a %T", refSchema.Type)
		}

		if asObject.Description != "Message Comment" {
			t.Errorf("unexpected description: '%s'", asObject.Description)
		}

		if len(asObject.Properties) != 2 {
			t.Fatalf("unexpected properties: %d", len(asObject.Properties))
		}

		f1 := asObject.Properties[0]
		if f1.Name != "field1" {
			t.Errorf("unexpected field name: '%s'", f1.Name)
		}

		if f1.Description != "Field Comment" {
			t.Errorf("unexpected description: '%s'", f1.Description)
		}

		fEnum := asObject.Properties[1]
		if fEnum.Name != "enum" {
			t.Errorf("unexpected field name: '%s'", fEnum.Name)
		}

		ref := fEnum.Schema.GetEnum().GetRef()
		if ref.Schema != "TestEnum" || ref.Package != "test.v1" {
			refStr := protojson.Format(ref)
			t.Fatalf("ref is %s", refStr)
		}

		schemaEnum, ok := schemas["test.v1.TestEnum"]
		if !ok {
			t.Fatalf("schema not found: '%s'", ref)
		}

		enumType := schemaEnum.GetEnum()
		if enumType == nil {
			t.Fatalf("unexpected type: %T", fEnum.Schema.Type)
		}

		if enumType.Options[0].Name != "UNSPECIFIED" {
			t.Errorf("unexpected enum value: '%s'", enumType.Options[0])
		}
		if enumType.Options[1].Name != "FOO" {
			t.Errorf("unexpected enum value: '%s'", enumType.Options[1])
		}

	})
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
			t.Logf("    W: %s", wantLine)
			t.Logf("    G: %s", gotLine)
			t.Log(strings.Repeat(`/\`, 10))
			break
		} else {
			t.Logf("   OK: %s", wantLine)
		}
		lineA++
		lineB++
	}
	if lineA < len(wantLines) {
		matched = false
		t.Logf("Remaining Want: \n%s", strings.Join(wantLines[lineA:], "  \n"))
	}
	if lineB < len(gotLines) {
		matched = false
		t.Logf("Remaining Got: \n%s", strings.Join(gotLines[lineB:], "  \n"))
	}

	if !matched {
		t.Errorf("unexpected JSON")
	}

}
