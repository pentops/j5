package structure

import (
	"encoding/json"
	"fmt"
	"testing"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/pentops/flowtest/jsontest"
	"github.com/pentops/jsonapi/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/jsonapi/gen/j5/ext/v1/ext_j5pb"
	"github.com/pentops/jsonapi/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/jsonapi/gen/test/foo/v1/foo_testpb"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func buildFieldSchema(t *testing.T, field *descriptorpb.FieldDescriptorProto, validate *validate.FieldConstraints) *jsontest.Asserter {
	ss := NewSchemaSet(&config_j5pb.CodecOptions{
		ShortEnums: &config_j5pb.ShortEnumOptions{
			StrictUnmarshal: true,
		},
		WrapOneof: true,
	})
	proto := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("test.proto"),
		Package: proto.String("test"),
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String("TestMessage"),
			Field: []*descriptorpb.FieldDescriptorProto{
				fieldWithValidateExtension(field, validate),
			},
		}},
	}
	schemaItem, err := ss.BuildSchemaObject(msgDesscriptorToReflection(t, proto))
	if err != nil {
		t.Fatal(err.Error())
	}
	obj, ok := schemaItem.Type.(*schema_j5pb.Schema_ObjectItem)
	if !ok {
		t.Fatalf("expected object item, got %T", schemaItem.Type)
	}
	prop := obj.ObjectItem.Properties[0]

	bt, err := protojson.Marshal(prop)
	if err != nil {
		t.Fatal(err.Error())
	}

	dd, err := jsontest.NewAsserter(json.RawMessage(bt))
	if err != nil {
		t.Fatal(err.Error())
	}
	return dd
}

func TestStringSchemaTypes(t *testing.T) {
	for _, tt := range []struct {
		name       string
		constraint *validate.StringRules
		expected   map[string]interface{}
	}{{
		name: "length constraints",
		constraint: &validate.StringRules{
			MinLen: proto.Uint64(1),
			MaxLen: proto.Uint64(10),
		},
		expected: map[string]interface{}{
			"schema.stringItem.rules.minLength": "1",
			"schema.stringItem.rules.maxLength": "10",
		},
	}, {
		name: "pattern constraint",
		constraint: &validate.StringRules{
			Pattern: proto.String("^[a-z]+$"),
		},
		expected: map[string]interface{}{
			"schema.stringItem.rules.pattern": "^[a-z]+$",
		},
	}, {
		name: "uuid constraint",
		constraint: &validate.StringRules{
			WellKnown: &validate.StringRules_Uuid{
				Uuid: true,
			},
		},
		expected: map[string]interface{}{
			"schema.stringItem.format": "uuid",
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			dd := buildFieldSchema(t, &descriptorpb.FieldDescriptorProto{
				Name:   proto.String("test_field"),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
				Number: proto.Int32(1),
			}, &validate.FieldConstraints{
				Type: &validate.FieldConstraints_String_{
					String_: tt.constraint,
				},
			})
			dd.Print(t)
			for path, expected := range tt.expected {
				dd.AssertEqual(t, path, expected)
			}
		})
	}
}

func TestSchemaTypesSimple(t *testing.T) {
	for _, tt := range []struct {
		name     string
		proto    *descriptorpb.FieldDescriptorProto
		validate *validate.FieldConstraints
		expected map[string]interface{}
	}{{
		name: "int32",
		proto: &descriptorpb.FieldDescriptorProto{
			Name:   proto.String("test_field"),
			Type:   descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum(),
			Number: proto.Int32(1),
		},
		validate: &validate.FieldConstraints{
			Type: &validate.FieldConstraints_Int32{
				Int32: &validate.Int32Rules{
					LessThan: &validate.Int32Rules_Lt{
						Lt: 10,
					},
					GreaterThan: &validate.Int32Rules_Gte{
						Gte: 1,
					},
				},
			},
		},
		expected: map[string]interface{}{
			"schema.integerItem.format":                 "int32",
			"schema.integerItem.rules.minimum":          "1",
			"schema.integerItem.rules.maximum":          "10",
			"schema.integerItem.rules.exclusiveMaximum": true,
		},
	}, {
		name: "int64",
		proto: &descriptorpb.FieldDescriptorProto{
			Name:   proto.String("test_field"),
			Type:   descriptorpb.FieldDescriptorProto_TYPE_INT64.Enum(),
			Number: proto.Int32(1),
		},
		validate: &validate.FieldConstraints{
			Type: &validate.FieldConstraints_Int64{
				Int64: &validate.Int64Rules{
					LessThan: &validate.Int64Rules_Lt{
						Lt: 10,
					},
					GreaterThan: &validate.Int64Rules_Gte{
						Gte: 1,
					},
				},
			},
		},
		expected: map[string]interface{}{
			"schema.integerItem.format":                 "int64",
			"schema.integerItem.rules.minimum":          "1",
			"schema.integerItem.rules.maximum":          "10",
			"schema.integerItem.rules.exclusiveMaximum": true,
		},
	}, {
		name: "uint32",
		proto: &descriptorpb.FieldDescriptorProto{
			Name:   proto.String("test_field"),
			Type:   descriptorpb.FieldDescriptorProto_TYPE_UINT32.Enum(),
			Number: proto.Int32(1),
		},
		validate: &validate.FieldConstraints{
			Type: &validate.FieldConstraints_Uint32{
				Uint32: &validate.UInt32Rules{
					LessThan: &validate.UInt32Rules_Lt{
						Lt: 10,
					},
					GreaterThan: &validate.UInt32Rules_Gte{
						Gte: 1,
					},
				},
			},
		},
		expected: map[string]interface{}{
			"schema.integerItem.format":                 "uint32",
			"schema.integerItem.rules.minimum":          "1",
			"schema.integerItem.rules.maximum":          "10",
			"schema.integerItem.rules.exclusiveMaximum": true,
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			dd := buildFieldSchema(t, tt.proto, tt.validate)
			dd.Print(t)
			for path, expected := range tt.expected {
				dd.AssertEqual(t, path, expected)
			}
		})
	}
}

func TestTestProtoSchemaTypes(t *testing.T) {
	ss := NewSchemaSet(&config_j5pb.CodecOptions{
		ShortEnums: &config_j5pb.ShortEnumOptions{
			StrictUnmarshal: true,
		},
		WrapOneof: true,
	})

	fooDesc := (&foo_testpb.PostFooRequest{}).ProtoReflect().Descriptor()

	t.Log(protojson.Format(protodesc.ToDescriptorProto(fooDesc)))

	schemaItem, err := ss.BuildSchemaObject(fooDesc)
	if err != nil {
		t.Fatal(err.Error())
	}

	obj := schemaItem.Type.(*schema_j5pb.Schema_ObjectItem)
	assertProperty := func(name string, expected map[string]interface{}) {
		for _, prop := range obj.ObjectItem.Properties {
			if prop.Name == name {
				t.Run(name, func(t *testing.T) {
					dd, err := jsontest.NewAsserter(prop)
					if err != nil {
						t.Fatal(err.Error())
					}
					dd.Print(t)
					dd.AssertEqualSet(t, "", expected)
				})
				return
			}
		}
		t.Errorf("property %q not found", name)
	}

	assertProperty("sString", map[string]interface{}{
		"protoFieldName":   "s_string",
		"protoFieldNumber": 1,
	})

	assertProperty("oString", map[string]interface{}{
		"protoFieldName":     "o_string",
		"protoFieldNumber":   2,
		"explicitlyOptional": true,
	})

	assertProperty("rString", map[string]interface{}{
		"protoFieldName":                    "r_string",
		"protoFieldNumber":                  3,
		"schema.arrayItem.items.stringItem": map[string]interface{}{},
	})

	assertProperty("mapStringString", map[string]interface{}{
		"protoFieldName":                       "map_string_string",
		"protoFieldNumber":                     15,
		"schema.mapItem.itemSchema.stringItem": map[string]interface{}{},
	})
}

func TestSchemaTypesComplex(t *testing.T) {
	ss := NewSchemaSet(&config_j5pb.CodecOptions{
		ShortEnums: &config_j5pb.ShortEnumOptions{
			StrictUnmarshal: true,
		},
		WrapOneof: true,
	})

	for _, tt := range []struct {
		name         string
		proto        *descriptorpb.FileDescriptorProto
		expected     map[string]interface{}
		expectedRefs map[string]map[string]interface{}
	}{{
		name: "empty message",
		proto: &descriptorpb.FileDescriptorProto{
			Name:    proto.String("test.proto"),
			Package: proto.String("test"),
			MessageType: []*descriptorpb.DescriptorProto{{
				Name: proto.String("TestMessage"),
			}},
		},
		expected: map[string]interface{}{
			"objectItem.protoMessageName": "TestMessage",
			//"objectItem.properties":       jsontest.LenEqual(0),
		},
	}, {
		name: "array field",
		proto: &descriptorpb.FileDescriptorProto{
			Name:    proto.String("test.proto"),
			Package: proto.String("test"),
			MessageType: []*descriptorpb.DescriptorProto{{
				Name: proto.String("TestMessage"),
				Field: []*descriptorpb.FieldDescriptorProto{{
					Name:   proto.String("test_field"),
					Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					Number: proto.Int32(1),
					Label:  descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
				}},
			}},
		},
		expected: map[string]interface{}{
			"objectItem.protoMessageName":  "TestMessage",
			"objectItem.properties.0.name": "testField",
		},
	}, {
		name: "flatten",
		proto: &descriptorpb.FileDescriptorProto{
			Name:    proto.String("test.proto"),
			Package: proto.String("test"),
			MessageType: []*descriptorpb.DescriptorProto{{
				Name: proto.String("TestMessage"),
				Field: []*descriptorpb.FieldDescriptorProto{
					fieldWithExtension(&descriptorpb.FieldDescriptorProto{
						Name:     proto.String("test_field"),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String("test.ChildMessage"),
						Number:   proto.Int32(1),
					}, ext_j5pb.E_Field, &ext_j5pb.FieldOptions{
						Type: &ext_j5pb.FieldOptions_Message{
							Message: &ext_j5pb.MessageFieldOptions{
								Flatten: true,
							},
						},
					})},
			}, {
				Name: proto.String("ChildMessage"),
				Field: []*descriptorpb.FieldDescriptorProto{{
					Name:   proto.String("child_field"),
					Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					Number: proto.Int32(1),
				}},
			}},
		},
		expected: map[string]interface{}{
			"objectItem.protoMessageName":               "TestMessage",
			"objectItem.properties.0.name":              "childField",
			"objectItem.properties.0.schema.stringItem": map[string]interface{}{},
		},
	}, {
		name: "map<string>string",
		proto: &descriptorpb.FileDescriptorProto{
			// Proto compiler creates an array of Key Value pairs for a
			// map[string]string
			Name:    proto.String("test.proto"),
			Package: proto.String("test"),
			MessageType: []*descriptorpb.DescriptorProto{{
				Name: proto.String("TestMessage"),
				Field: []*descriptorpb.FieldDescriptorProto{{
					Name:     proto.String("test_field"),
					Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
					TypeName: proto.String("test.TestMessage.TestFieldEntry"),
					Number:   proto.Int32(1),
					Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
				}},
				NestedType: []*descriptorpb.DescriptorProto{{
					Name: proto.String("TestFieldEntry"),
					Field: []*descriptorpb.FieldDescriptorProto{{
						Name:     proto.String("key"),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Number:   proto.Int32(1),
						JsonName: proto.String("key"),
					}, {
						Name:     proto.String("value"),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Number:   proto.Int32(2),
						JsonName: proto.String("value"),
					}},
					Options: &descriptorpb.MessageOptions{
						MapEntry: proto.Bool(true),
					},
				}},
			}},
		},
		expected: map[string]interface{}{
			// Outer Wrapper
			"objectItem.protoMessageName":                                  "TestMessage",
			"objectItem.properties.0.name":                                 "testField",
			"objectItem.properties.0.schema.mapItem.itemSchema.stringItem": map[string]interface{}{},
		},
	}, {
		name: "enum field",
		proto: &descriptorpb.FileDescriptorProto{
			Name:    proto.String("test.proto"),
			Package: proto.String("test"),
			MessageType: []*descriptorpb.DescriptorProto{{
				Name: proto.String("TestMessage"),
				Field: []*descriptorpb.FieldDescriptorProto{
					fieldWithValidateExtension(&descriptorpb.FieldDescriptorProto{
						Name:     proto.String("test_field"),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_ENUM.Enum(),
						TypeName: proto.String("test.TestEnum"),
						Number:   proto.Int32(1),
					}, &validate.FieldConstraints{
						Type: &validate.FieldConstraints_Enum{
							Enum: &validate.EnumRules{
								DefinedOnly: proto.Bool(true),
								NotIn:       []int32{0},
							},
						},
					}),
				},
			}},
			EnumType: []*descriptorpb.EnumDescriptorProto{{
				Name: proto.String("TestEnum"),
				Value: []*descriptorpb.EnumValueDescriptorProto{{
					Name:   proto.String("TEST_ENUM_UNSPECIFIED"),
					Number: proto.Int32(0),
				}, {
					Name:   proto.String("TEST_ENUM_FOO"),
					Number: proto.Int32(1),
				}, {
					Name:   proto.String("TEST_ENUM_BAR"),
					Number: proto.Int32(2),
				}},
			}},
		},
		expected: map[string]interface{}{
			"objectItem.properties.0.schema.ref": "test.TestEnum",
		},
		expectedRefs: map[string]map[string]interface{}{
			"test.TestEnum": {
				"enumItem.options.0.name": "FOO",
				"enumItem.options.1.name": "BAR",
			},
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			schemaItem, err := ss.BuildSchemaObject(msgDesscriptorToReflection(t, tt.proto))
			if err != nil {
				t.Fatal(err.Error())
			}

			dd, err := jsontest.NewAsserter(schemaItem)
			if err != nil {
				t.Fatal(err.Error())
			}

			dd.Print(t)

			for path, expected := range tt.expected {
				dd.AssertEqual(t, path, expected)
			}

			for path, expectSet := range tt.expectedRefs {
				schema, ok := ss.Schemas[path]
				if !ok {
					t.Fatalf("schema %q not found", path)
				}
				ddRef, err := jsontest.NewAsserter(schema)
				if err != nil {
					t.Fatal(err.Error())
				}
				ddRef.Print(t)
				for path, expected := range expectSet {
					ddRef.AssertEqual(t, path, expected)
				}
			}

		})
	}
}

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

func msgDesscriptorToReflection(t testing.TB, fileDescriptor *descriptorpb.FileDescriptorProto) protoreflect.MessageDescriptor {
	t.Helper()
	files, err := protodesc.NewFiles(&descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{fileDescriptor},
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	desc, err := files.FindDescriptorByName(protoreflect.FullName(fmt.Sprintf("%s.%s", *fileDescriptor.Package, fileDescriptor.MessageType[0].GetName())))
	if err != nil {
		t.Fatal(err.Error())
	}

	descMsg, ok := desc.(protoreflect.MessageDescriptor)
	if !ok {
		t.Fatal("not a message descriptor")
	}

	return descMsg
}

func TestCommentBuilder(t *testing.T) {

	for _, tc := range []struct {
		name     string
		leading  string
		trailing string
		expected string
	}{{
		name:     "leading",
		leading:  "comment",
		expected: "comment",
	}, {
		name:     "fallback",
		expected: "fallback",
	}, {
		name:     "both",
		leading:  "leading",
		trailing: "trailing",
		expected: "leading\ntrailing",
	}, {
		name:     "multiline",
		leading:  "line1\n  line2",
		trailing: "line3\n  line4",
		expected: "line1\nline2\nline3\nline4",
	}, {
		name:     "multiline commented",
		leading:  "#line1\nline2",
		expected: "line2",
	}, {
		name:     "commented fallback",
		leading:  "#line1",
		expected: "fallback",
	}} {
		t.Run(tc.name, func(t *testing.T) {
			sl := protoreflect.SourceLocation{
				LeadingComments:  tc.leading,
				TrailingComments: tc.trailing,
			}

			got := buildComment(sl, "fallback")
			if got != tc.expected {
				t.Errorf("expected comment: '%s', got '%s'", tc.expected, got)
			}

		})
	}
}
