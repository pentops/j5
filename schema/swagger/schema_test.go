package swagger

import (
	"testing"

	"github.com/pentops/flowtest/jsontest"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
)

func TestConvertSchema(t *testing.T) {

	for _, tc := range []struct {
		name  string
		input *schema_j5pb.Schema
		want  map[string]interface{}
	}{{
		name: "string",
		input: &schema_j5pb.Schema{
			Description: "desc",
			Type: &schema_j5pb.Schema_StringItem{
				StringItem: &schema_j5pb.StringItem{
					Format:  Ptr("uuid"),
					Example: Ptr("example"),
					Rules: &schema_j5pb.StringRules{
						Pattern:   Ptr("regex-pattern"),
						MinLength: Ptr(uint64(1)),
						MaxLength: Ptr(uint64(2)),
					},
				},
			},
		},
		want: map[string]interface{}{
			"type":        "string",
			"example":     "example",
			"format":      "uuid",
			"pattern":     "regex-pattern",
			"minLength":   1,
			"maxLength":   2,
			"description": "desc",
		},
	}, {
		name: "number",
		input: &schema_j5pb.Schema{
			Type: &schema_j5pb.Schema_NumberItem{
				NumberItem: &schema_j5pb.NumberItem{
					Format: "double",
					Rules: &schema_j5pb.NumberRules{
						Minimum:          Ptr(0.0),
						Maximum:          Ptr(100.0),
						ExclusiveMinimum: Ptr(true),
						ExclusiveMaximum: Ptr(false),
					},
				},
			},
		},
		want: map[string]interface{}{
			"type":             "number",
			"format":           "double",
			"minimum":          0.0,
			"maximum":          100.0,
			"exclusiveMinimum": true,
			"exclusiveMaximum": false,
		},
	}, {
		name: "enum",
		input: &schema_j5pb.Schema{
			Type: &schema_j5pb.Schema_EnumItem{
				EnumItem: &schema_j5pb.EnumItem{
					Options: []*schema_j5pb.EnumItem_Value{{
						Name:        "FOO",
						Description: "Foo Description",
					}, {
						Name:        "BAR",
						Description: "Bar Description",
					}},
				}},
		},
		want: map[string]interface{}{
			// json schema doesn't have an actual 'enum' type, enum is just an
			// extension on any other type. Our enums are always strings.
			"type":                 "string",
			"x-enum.0.name":        "FOO",
			"x-enum.0.description": "Foo Description",
			"x-enum.1.name":        "BAR",
			"x-enum.1.description": "Bar Description",
			"enum.0":               "FOO",
			"enum.1":               "BAR",
		},
	}, {
		name: "ref",
		input: &schema_j5pb.Schema{
			Type: &schema_j5pb.Schema_Ref{
				Ref: "#/definitions/foo",
			},
		},
		want: map[string]interface{}{
			"$ref": "#/definitions/foo",
		},
	}, {
		name: "object",
		input: &schema_j5pb.Schema{
			Type: &schema_j5pb.Schema_ObjectItem{
				ObjectItem: &schema_j5pb.ObjectItem{
					GoPackageName:   "Go Package",
					GoTypeName:      "Go Type",
					GrpcPackageName: "Grpc Package",

					ProtoFullName:    "long",
					ProtoMessageName: "short",

					Rules: &schema_j5pb.ObjectRules{
						MinProperties: Ptr(uint64(1)),
						MaxProperties: Ptr(uint64(2)),
					},
					Properties: []*schema_j5pb.ObjectProperty{{
						Name:             "foo",
						Required:         true,
						ProtoFieldName:   "foo",
						ProtoFieldNumber: 1,
						Schema: &schema_j5pb.Schema{
							Type: &schema_j5pb.Schema_StringItem{
								StringItem: &schema_j5pb.StringItem{},
							},
						},
					}, {
						Name:             "bar",
						Required:         false,
						ProtoFieldName:   "bar",
						ProtoFieldNumber: 2,
						Schema: &schema_j5pb.Schema{
							Type: &schema_j5pb.Schema_StringItem{
								StringItem: &schema_j5pb.StringItem{},
							},
						},
					}},
				},
			},
		},
		want: map[string]interface{}{
			"type":                          "object",
			"x-proto-name":                  "short",
			"x-proto-full-name":             "long",
			"required.0":                    "foo",
			"properties.foo.type":           "string",
			"properties.foo.x-proto-name":   "foo",
			"properties.foo.x-proto-number": 1,
			"properties.bar.type":           "string",
		},
	}, {
		name: "oneof",
		input: &schema_j5pb.Schema{
			Type: &schema_j5pb.Schema_OneofWrapper{
				OneofWrapper: &schema_j5pb.OneofWrapperItem{
					GoPackageName:   "Go Package",
					GoTypeName:      "Go Type",
					GrpcPackageName: "Grpc Package",

					ProtoFullName:    "long",
					ProtoMessageName: "short",

					Properties: []*schema_j5pb.ObjectProperty{{
						Name:             "foo",
						ProtoFieldName:   "foo",
						ProtoFieldNumber: 1,
						Schema: &schema_j5pb.Schema{
							Type: &schema_j5pb.Schema_StringItem{
								StringItem: &schema_j5pb.StringItem{},
							},
						},
					}, {
						Name:             "bar",
						ProtoFieldName:   "bar",
						ProtoFieldNumber: 2,
						Schema: &schema_j5pb.Schema{
							Type: &schema_j5pb.Schema_StringItem{
								StringItem: &schema_j5pb.StringItem{},
							},
						},
					}},
				},
			},
		},
		want: map[string]interface{}{
			"type":                        "object",
			"x-proto-name":                "short",
			"x-proto-full-name":           "long",
			"properties.foo.type":         "string",
			"properties.foo.x-proto-name": "foo",
			"properties.bar.type":         "string",
			"properties.bar.x-proto-name": "bar",
		},
	}, {
		name: "array",
		input: &schema_j5pb.Schema{
			Type: &schema_j5pb.Schema_ArrayItem{
				ArrayItem: &schema_j5pb.ArrayItem{
					Items: &schema_j5pb.Schema{
						Type: &schema_j5pb.Schema_StringItem{
							StringItem: &schema_j5pb.StringItem{},
						},
					},
					Rules: &schema_j5pb.ArrayRules{
						MinItems:    Ptr(uint64(1)),
						MaxItems:    Ptr(uint64(2)),
						UniqueItems: Ptr(true),
					},
				},
			},
		},
		want: map[string]interface{}{
			"type":        "array",
			"items.type":  "string",
			"minItems":    1,
			"maxItems":    2,
			"uniqueItems": true,
		},
	}, {
		name: "map",
		input: &schema_j5pb.Schema{
			Type: &schema_j5pb.Schema_MapItem{
				MapItem: &schema_j5pb.MapItem{
					ItemSchema: &schema_j5pb.Schema{
						Type: &schema_j5pb.Schema_StringItem{
							StringItem: &schema_j5pb.StringItem{},
						},
					},
				},
			},
		},
		want: map[string]interface{}{
			"type":                      "object",
			"additionalProperties.type": "string",
			"x-key-property.type":       "string",
		},
	}, {
		name: "any",
		input: &schema_j5pb.Schema{
			Type: &schema_j5pb.Schema_Any{
				Any: &schema_j5pb.AnySchemmaItem{},
			},
		},
		want: map[string]interface{}{
			"type":                 "object",
			"additionalProperties": true,
		},
	}} {
		t.Run(tc.name, func(t *testing.T) {

			output, err := ConvertSchema(tc.input)
			if err != nil {
				t.Fatal(err)
			}

			// assertions in JSON as the implementation doesn't actually matter
			out, err := jsontest.NewAsserter(output)
			if err != nil {
				t.Fatal(err)
			}

			out.Print(t)
			out.AssertEqualSet(t, "", tc.want)

		})

	}
}

func TestSchemaJSONMarshal(t *testing.T) {

	object := &Schema{
		SchemaItem: &SchemaItem{
			Type: &ObjectItem{
				Required: []string{"id"},
				Properties: map[string]*ObjectProperty{
					"id": {
						Schema: &Schema{
							SchemaItem: &SchemaItem{
								Description: "desc",
								Type: StringItem{
									Format: Some("uuid"),
								},
							},
						},
					},
					"number": {
						Schema: &Schema{
							SchemaItem: &SchemaItem{
								Type: NumberItem{
									Format:  "double",
									Minimum: Value(0.0),
									Maximum: Value(100.0),
								},
							},
						},
					},
					"object": {
						Schema: &Schema{
							SchemaItem: &SchemaItem{
								Type: &ObjectItem{
									Required: []string{"foo"},
									Properties: map[string]*ObjectProperty{
										"foo": {
											Schema: &Schema{
												SchemaItem: &SchemaItem{
													Type: StringItem{},
												},
											},
										},
									},
								},
							},
						},
					},
					"ref": {
						Schema: &Schema{
							Ref: Ptr("#/definitions/foo"),
						},
					},
					"oneof": {
						Schema: &Schema{
							OneOf: []*Schema{{
								SchemaItem: &SchemaItem{
									Type: StringItem{},
								},
							}, {
								Ref: Ptr("#/foo/bar"),
							}},
						},
					},
				},
			},
		},
	}

	out, err := jsontest.NewAsserter(object)
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
