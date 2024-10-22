package j5schema

import (
	"encoding/json"
	"testing"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/pentops/flowtest/jsontest"
	"github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"

	"github.com/pentops/j5/gen/test/schema/v1/schema_testpb"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

func buildFieldSchema(t *testing.T, field *descriptorpb.FieldDescriptorProto, validate *validate.FieldConstraints) *jsontest.Asserter {
	ss := NewSchemaCache()
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
	reflectRoot, err := ss.Schema(msgDesscriptorToReflection(t, proto))
	if err != nil {
		t.Fatal(err.Error())
	}
	schemaItem := reflectRoot.ToJ5Root()
	obj, ok := schemaItem.Type.(*schema_j5pb.RootSchema_Object)
	if !ok {
		t.Fatalf("expected object item, got %T", schemaItem.Type)
	}
	prop := obj.Object.Properties[0]

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
			"schema.string.rules.minLength": "1",
			"schema.string.rules.maxLength": "10",
		},
	}, {
		name: "pattern constraint",
		constraint: &validate.StringRules{
			Pattern: proto.String("^[a-z]+$"),
		},
		expected: map[string]interface{}{
			"schema.string.rules.pattern": "^[a-z]+$",
		},
	}, {
		name: "uuid constraint",
		constraint: &validate.StringRules{
			WellKnown: &validate.StringRules_Uuid{
				Uuid: true,
			},
		},
		expected: map[string]interface{}{
			"schema.key.format": jsontest.IsOneofKey("uuid"),
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
			"schema.integer.format":                 schema_j5pb.IntegerField_FORMAT_INT32.String(),
			"schema.integer.rules.minimum":          "1",
			"schema.integer.rules.maximum":          "10",
			"schema.integer.rules.exclusiveMaximum": true,
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
			"schema.integer.format":                 schema_j5pb.IntegerField_FORMAT_INT64.String(),
			"schema.integer.rules.minimum":          "1",
			"schema.integer.rules.maximum":          "10",
			"schema.integer.rules.exclusiveMaximum": true,
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
			"schema.integer.format":                 schema_j5pb.IntegerField_FORMAT_UINT32.String(),
			"schema.integer.rules.minimum":          "1",
			"schema.integer.rules.maximum":          "10",
			"schema.integer.rules.exclusiveMaximum": true,
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

	ss := NewSchemaCache()

	fooDesc := (&schema_testpb.FullSchema{}).ProtoReflect().Descriptor()

	t.Log(protojson.Format(protodesc.ToDescriptorProto(fooDesc)))

	reflectRoot, err := ss.Schema(fooDesc)
	if err != nil {
		t.Fatal(err.Error())
	}

	schemaItem := reflectRoot.ToJ5Root()

	obj := schemaItem.Type.(*schema_j5pb.RootSchema_Object)
	assertProperty := func(name string, expected map[string]interface{}) {
		t.Helper()
		for _, prop := range obj.Object.Properties {
			if prop.Name == name {
				t.Run(name, func(t *testing.T) {
					t.Helper()
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
		"protoField": jsontest.Array[float64]{1},
	})

	assertProperty("oString", map[string]interface{}{
		"protoField":         jsontest.Array[float64]{2},
		"explicitlyOptional": true,
	})

	assertProperty("rString", map[string]interface{}{
		"protoField":                jsontest.Array[float64]{3},
		"schema.array.items.string": map[string]interface{}{},
	})

	assertProperty("mapStringString", map[string]interface{}{
		"protoField":                   jsontest.Array[float64]{36},
		"schema.map.itemSchema.string": map[string]interface{}{},
	})

	assertProperty("flattened", map[string]interface{}{
		"protoField":            jsontest.Array[float64]{42},
		"schema.object.flatten": true,
	})
}

func TestSchemaTypesComplex(t *testing.T) {

	type testCase struct {
		proto        *descriptorpb.FileDescriptorProto
		expected     map[string]interface{}
		expectedRefs map[string]map[string]interface{}
	}

	runTestCase := func(t *testing.T, tt testCase) {
		t.Helper()
		ss := NewSchemaCache()
		reflectRoot, err := ss.Schema(msgDesscriptorToReflection(t, tt.proto))
		if err != nil {
			t.Fatal(err.Error())
		}
		schemaItem := reflectRoot.ToJ5Root()

		dd, err := jsontest.NewAsserter(schemaItem)
		if err != nil {
			t.Fatal(err.Error())
		}

		dd.Print(t)

		for path, expected := range tt.expected {
			dd.AssertEqual(t, path, expected)
		}

		testPkg := ss.packages["test"]

		for path, expectSet := range tt.expectedRefs {
			ref, ok := testPkg.Schemas[path]
			if !ok {
				t.Fatalf("schema %q not found", path)
			}
			schema := ref.To
			if schema == nil {
				t.Fatalf("schema %q not linked", path)
			}
			schemaJ5 := schema.ToJ5Root()

			ddRef, err := jsontest.NewAsserter(schemaJ5)
			if err != nil {
				t.Fatal(err.Error())
			}
			ddRef.Print(t)
			for path, expected := range expectSet {
				ddRef.AssertEqual(t, path, expected)
			}
		}

	}

	t.Run("empty message", func(t *testing.T) {
		runTestCase(t, testCase{
			proto: &descriptorpb.FileDescriptorProto{
				Name:    proto.String("test.proto"),
				Package: proto.String("test"),
				MessageType: []*descriptorpb.DescriptorProto{{
					Name: proto.String("TestMessage"),
				}},
			},
			expected: map[string]interface{}{
				"object.name": "TestMessage",
				//"object.properties":       jsontest.LenEqual(0),
			},
		})
	})

	t.Run("array field", func(t *testing.T) {
		runTestCase(t, testCase{
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
				"object.name":              "TestMessage",
				"object.properties.0.name": "testField",
			},
		})
	})

	t.Run("flatten", func(t *testing.T) {
		runTestCase(t, testCase{
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
				"object.name":                               "TestMessage",
				"object.properties.0.name":                  "testField",
				"object.properties.0.schema.object.flatten": true,
			},
		})
	})

	t.Run("exposedOneof", func(t *testing.T) {
		runTestCase(t, testCase{
			proto: &descriptorpb.FileDescriptorProto{
				Name:    proto.String("test.proto"),
				Package: proto.String("test"),
				MessageType: []*descriptorpb.DescriptorProto{{
					Name: proto.String("TestMessage"),
					OneofDecl: []*descriptorpb.OneofDescriptorProto{{
						Name: proto.String("expose_me"),
						Options: extend(&descriptorpb.OneofOptions{}, ext_j5pb.E_Oneof, &ext_j5pb.OneofOptions{
							Expose: true,
						}),
					}},
					Field: []*descriptorpb.FieldDescriptorProto{{
						Name:       proto.String("test_field"),
						Type:       descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Number:     proto.Int32(1),
						OneofIndex: proto.Int32(0),
					}},
				}},
			},
			expected: map[string]interface{}{
				"object.name":              "TestMessage",
				"object.properties.0.name": "exposeMe",
				"object.properties.0.schema.oneof.ref.package": "test",
				"object.properties.0.schema.oneof.ref.schema":  "TestMessage_expose_me",
				"object.properties.0.protoField":               jsontest.NotSet{},
			},
			expectedRefs: map[string]map[string]interface{}{
				"TestMessage_expose_me": {
					"oneof.properties.0.name":       "testField",
					"oneof.properties.0.protoField": jsontest.Array[float64]{1},
				},
			},
		})
	})
	t.Run("map<string>string", func(t *testing.T) {
		runTestCase(t, testCase{
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
				"object.name":              "TestMessage",
				"object.properties.0.name": "testField",
				"object.properties.0.schema.map.itemSchema.string": map[string]interface{}{},
			},
		})
	})

	t.Run("enum field", func(t *testing.T) {
		runTestCase(t, testCase{
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
					Options: extend(&descriptorpb.EnumOptions{}, ext_j5pb.E_Enum, &ext_j5pb.EnumOptions{
						NoDefault: true,
					}),
				}},
			},
			expected: map[string]interface{}{
				"object.properties.0.schema.enum.ref.package": "test",
				"object.properties.0.schema.enum.ref.schema":  "TestEnum",
			},
			expectedRefs: map[string]map[string]interface{}{
				"TestEnum": {
					"enum.options.0.name": "FOO",
					"enum.options.1.name": "BAR",
				},
			},
		})
	})

	simpleFields := func(fields ...*descriptorpb.FieldDescriptorProto) *descriptorpb.FileDescriptorProto {
		return &descriptorpb.FileDescriptorProto{
			Name:    proto.String("test.proto"),
			Package: proto.String("test"),
			MessageType: []*descriptorpb.DescriptorProto{{
				Name:  proto.String("TestMessage"),
				Field: fields,
			}},
		}
	}

	t.Run("any Proto", func(t *testing.T) {
		base := simpleFields(&descriptorpb.FieldDescriptorProto{
			Name:     proto.String("pbany"),
			Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
			TypeName: proto.String("google.protobuf.Any"),
			Number:   proto.Int32(1),
			Options: extend(&descriptorpb.FieldOptions{}, ext_j5pb.E_Field, &ext_j5pb.FieldOptions{
				Type: &ext_j5pb.FieldOptions_Any{
					Any: &ext_j5pb.AnyField{
						OnlyDefined: true,
						Types:       []string{"foo.v1.foo", "foo.v1.bar"},
					},
				},
			}),
		})

		base.Dependency = []string{"google/protobuf/any.proto"}
		runTestCase(t, testCase{
			proto: base,
			expected: map[string]interface{}{
				"object.properties.0.schema.any.onlyDefined": true,
				"object.properties.0.schema.any.types":       []interface{}{"foo.v1.foo", "foo.v1.bar"},
			},
		})

	})

	t.Run("any J5", func(t *testing.T) {
		base := simpleFields(&descriptorpb.FieldDescriptorProto{
			Name:     proto.String("pbany"),
			Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
			TypeName: proto.String("j5.types.any.v1.Any"),
			Number:   proto.Int32(1),
			Options: extend(&descriptorpb.FieldOptions{}, ext_j5pb.E_Field, &ext_j5pb.FieldOptions{
				Type: &ext_j5pb.FieldOptions_Any{
					Any: &ext_j5pb.AnyField{
						OnlyDefined: true,
						Types:       []string{"foo.v1.foo", "foo.v1.bar"},
					},
				},
			}),
		})

		base.Dependency = []string{"j5/types/any/v1/any.proto"}
		runTestCase(t, testCase{
			proto: base,
			expected: map[string]interface{}{
				"object.properties.0.schema.any.onlyDefined": true,
				"object.properties.0.schema.any.types":       []interface{}{"foo.v1.foo", "foo.v1.bar"},
			},
		})

	})

	t.Run("any member", func(t *testing.T) {
		base := &descriptorpb.FileDescriptorProto{
			Name:    proto.String("test.proto"),
			Package: proto.String("test"),
			MessageType: []*descriptorpb.DescriptorProto{{
				Name:  proto.String("TestMessage"),
				Field: []*descriptorpb.FieldDescriptorProto{},
				Options: extend(&descriptorpb.MessageOptions{}, ext_j5pb.E_Message, &ext_j5pb.MessageOptions{
					Type: &ext_j5pb.MessageOptions_Object{
						Object: &ext_j5pb.ObjectMessageOptions{
							AnyMember: []string{"foo"},
						},
					},
				}),
			}},
		}

		base.Dependency = []string{"j5/ext/v1/annotations.proto"}
		runTestCase(t, testCase{
			proto: base,
			expected: map[string]interface{}{
				"object.anyMember": []interface{}{"foo"},
			},
		})

	})

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

func extend[T proto.Message](v T, extensionType protoreflect.ExtensionType, extensionValue interface{}) T {
	proto.SetExtension(v, extensionType, extensionValue)
	return v
}

func msgDesscriptorToReflection(t testing.TB, fileDescriptor *descriptorpb.FileDescriptorProto) protoreflect.MessageDescriptor {
	t.Helper()
	file, err := protodesc.NewFile(fileDescriptor, protoregistry.GlobalFiles)
	if err != nil {
		t.Fatal(err.Error())
	}

	return file.Messages().Get(0)

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
