package jsonapi

import (
	"encoding/json"
	"fmt"
	"testing"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/pentops/jsonapi/testproto/gen/testpb"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func buildFieldSchema(t *testing.T, field *descriptorpb.FieldDescriptorProto, validate *validate.FieldConstraints) *DynamicJSON {
	ss := NewSchemaSet(Options{
		ShortEnums: &ShortEnumsOption{
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
	obj, ok := schemaItem.ItemType.(*ObjectItem)
	if !ok {
		t.Fatalf("expected object item, got %T", schemaItem.ItemType)
	}
	prop := obj.Properties[0]

	dd, err := MarshalDynamic(prop)
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
			"type":      "string",
			"minLength": float64(1),
			"maxLength": float64(10),
		},
	}, {
		name: "pattern constraint",
		constraint: &validate.StringRules{
			Pattern: proto.String("^[a-z]+$"),
		},
		expected: map[string]interface{}{
			"type":    "string",
			"pattern": "^[a-z]+$",
		},
	}, {
		name: "uuid constraint",
		constraint: &validate.StringRules{
			WellKnown: &validate.StringRules_Uuid{
				Uuid: true,
			},
		},
		expected: map[string]interface{}{
			"type":   "string",
			"format": "uuid",
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
			"type":             "integer",
			"format":           "int32",
			"minimum":          float64(1),
			"maximum":          float64(10),
			"exclusiveMaximum": true,
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
			"type":             "integer",
			"format":           "int64",
			"minimum":          float64(1),
			"maximum":          float64(10),
			"exclusiveMaximum": true,
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
			"type":             "integer",
			"format":           "uint32",
			"minimum":          float64(1),
			"maximum":          float64(10),
			"exclusiveMaximum": true,
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

	ss := NewSchemaSet(Options{
		ShortEnums: &ShortEnumsOption{
			StrictUnmarshal: true,
		},
		WrapOneof: true,
	})

	fooDesc := (&testpb.PostFooRequest{}).ProtoReflect().Descriptor()

	t.Log(protojson.Format(protodesc.ToDescriptorProto(fooDesc)))

	schemaItem, err := ss.BuildSchemaObject(fooDesc)
	if err != nil {
		t.Fatal(err.Error())
	}

	dd, err := MarshalDynamic(schemaItem)
	if err != nil {
		t.Fatal(err.Error())
	}

	assertProperty := func(name string, expected map[string]interface{}) {
		t.Run(name, func(t *testing.T) {
			dd.PrintAt(t, fmt.Sprintf("properties.%s", name))
			dd.AssertEqualSet(t, fmt.Sprintf("properties.%s", name), expected)
		})

	}

	assertProperty("sString", map[string]interface{}{
		"type":             "string",
		"x-proto-name":     "s_string",
		"x-proto-number":   1,
		"x-proto-optional": false,
	})

	assertProperty("oString", map[string]interface{}{
		"type":             "string",
		"x-proto-name":     "o_string",
		"x-proto-number":   2,
		"x-proto-optional": true,
	})

	assertProperty("rString", map[string]interface{}{
		"type":             "array",
		"x-proto-name":     "r_string",
		"x-proto-number":   3,
		"x-proto-optional": false,
		"items.type":       "string",
	})

	assertProperty("mapStringString", map[string]interface{}{
		"type":                      "object",
		"x-proto-name":              "map_string_string",
		"x-proto-number":            15,
		"x-proto-optional":          false,
		"additionalProperties.type": "string",
		"x-key-property.type":       "string",
	})

}

func TestSchemaTypesComplex(t *testing.T) {

	ss := NewSchemaSet(Options{
		ShortEnums: &ShortEnumsOption{
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
			"x-proto-name": "TestMessage",
			"type":         "object",
			"properties":   LenEqual(0),
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
			"x-proto-name":              "TestMessage",
			"type":                      "object",
			"properties.testField.type": "array",
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
			"x-proto-name":              "TestMessage",
			"type":                      "object",
			"properties.testField.type": "object",
			"properties.testField.additionalProperties.type": "string",
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
			"x-proto-name":              "TestMessage",
			"type":                      "object",
			"properties":                LenEqual(1),
			"properties.testField.$ref": "test.TestEnum",
		},
		expectedRefs: map[string]map[string]interface{}{
			"test.TestEnum": {
				"enum.0": "FOO",
				"enum.1": "BAR",
			},
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			schemaItem, err := ss.BuildSchemaObject(msgDesscriptorToReflection(t, tt.proto))
			if err != nil {
				t.Fatal(err.Error())
			}

			dd, err := MarshalDynamic(schemaItem)
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
				ddRef, err := MarshalDynamic(schema)
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
	if field.Options == nil {
		field.Options = &descriptorpb.FieldOptions{}
	}

	proto.SetExtension(field.Options, validate.E_Field, constraints)
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

func TestSchemaJSONMarshal(t *testing.T) {

	object := SchemaItem{
		ItemType: &ObjectItem{
			debug: "a",
			Properties: []*ObjectProperty{{
				Name: "id",
				SchemaItem: SchemaItem{
					Description: "desc",
					ItemType: StringItem{
						Format:      "uuid",
						StringRules: StringRules{},
					},
				},
				Required: true,
			}, {
				Name: "number",
				SchemaItem: SchemaItem{
					ItemType: NumberItem{
						Format: "double",
						NumberRules: NumberRules{
							Minimum: Value(0.0),
							Maximum: Value(100.0),
						},
					},
				},
			}, {
				Name: "object",
				SchemaItem: SchemaItem{
					ItemType: &ObjectItem{
						debug: "b",
						Properties: []*ObjectProperty{{
							Name:     "foo",
							Required: true,
							SchemaItem: SchemaItem{
								ItemType: StringItem{},
							},
						}},
					},
				},
			}, {
				Name: "ref",
				SchemaItem: SchemaItem{
					Ref: "#/definitions/foo",
				},
			}, {
				Name: "oneof",
				SchemaItem: SchemaItem{
					OneOf: []SchemaItem{
						{
							ItemType: StringItem{},
						},
						{
							Ref: "#/foo/bar",
						},
					},
				},
			},
			},
		},
	}

	out, err := MarshalDynamic(object)
	if err != nil {
		t.Error(err)
	}

	out.Print(t)
	out.AssertEqual(t, "type", "object")
	out.AssertEqual(t, "properties.id.type", "string")
	out.AssertEqual(t, "properties.id.format", "uuid")
	out.AssertEqual(t, "properties.id.description", "desc")
	out.AssertEqual(t, "required.0", "id")

	out.AssertEqual(t, "properties.number.type", "number")
	out.AssertEqual(t, "properties.number.format", "double")
	out.AssertEqual(t, "properties.number.minimum", 0.0)
	out.AssertEqual(t, "properties.number.maximum", 100.0)
	out.AssertNotSet(t, "properties.number.exclusiveMinimum")

	out.AssertEqual(t, "properties.object.properties.foo.type", "string")

	out.AssertEqual(t, "properties.ref.$ref", "#/definitions/foo")

}

type DynamicJSON struct {
	JSON string
}

func MarshalDynamic(v interface{}) (*DynamicJSON, error) {
	val, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return &DynamicJSON{JSON: string(val)}, nil
}

func (d *DynamicJSON) Print(t testing.TB) {
	t.Log(string(d.JSON))
}

func (d *DynamicJSON) PrintAt(t testing.TB, path string) {
	val := gjson.Get(d.JSON, path)
	if val.Exists() {
		t.Log(val.String())
	} else {
		t.Log("path not found")
	}
}

func (d *DynamicJSON) Get(path string) (interface{}, bool) {
	val := gjson.Get(d.JSON, path)
	if val.Exists() {
		return val.Value(), true
	}
	return nil, false
}

type LenEqual int

func (d *DynamicJSON) AssertEqual(t testing.TB, path string, value interface{}) {
	t.Helper()
	actual, ok := d.Get(path)
	if !ok {
		t.Errorf("path %q not found", path)
		return
	}

	switch value.(type) {
	case LenEqual:
		actualSlice, ok := actual.([]interface{})
		if ok {
			if len(actualSlice) != int(value.(LenEqual)) {
				t.Errorf("expected %d, got %d", value, len(actualSlice))
			}
			return
		}
		actualMap, ok := actual.(map[string]interface{})
		if ok {
			if len(actualMap) != int(value.(LenEqual)) {
				t.Errorf("expected %d, got %d", value, len(actualMap))
			}
			return
		}
		t.Errorf("expected len(%d), got non len object %T", value, actual)
	default:
		assert.EqualValues(t, value, actual, "at path %q", path)
	}
}

func (d *DynamicJSON) AssertNotSet(t testing.TB, path string) {
	_, ok := d.Get(path)
	if ok {
		t.Errorf("path %q was set", path)
	}
}

func (d *DynamicJSON) AssertEqualSet(t testing.TB, path string, expected map[string]interface{}) {
	t.Helper()
	for key, expectSet := range expected {
		pathKey := key
		if path != "" {
			pathKey = fmt.Sprintf("%s.%s", path, key)
		}

		d.AssertEqual(t, pathKey, expectSet)
	}
}
