package swagger

import (
	"testing"

	"github.com/pentops/jsonapi/gen/v1/jsonapi_pb"
	"github.com/pentops/jsonapi/jsontest"
)

func TestConvertSchema(t *testing.T) {

	for _, tc := range []struct {
		name  string
		input *jsonapi_pb.Schema
		want  map[string]interface{}
	}{{
		name: "string",
		input: &jsonapi_pb.Schema{
			Description: "desc",
			Type: &jsonapi_pb.Schema_StringItem{
				StringItem: &jsonapi_pb.StringItem{
					Format:  Ptr("uuid"),
					Example: Ptr("example"),
					Rules: &jsonapi_pb.StringRules{
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
		input: &jsonapi_pb.Schema{
			Type: &jsonapi_pb.Schema_NumberItem{
				NumberItem: &jsonapi_pb.NumberItem{
					Format: "double",
					Rules: &jsonapi_pb.NumberRules{
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
		input: &jsonapi_pb.Schema{
			Type: &jsonapi_pb.Schema_EnumItem{
				EnumItem: &jsonapi_pb.EnumItem{
					Options: []*jsonapi_pb.EnumItem_Value{{
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
		name: "object",
		input: &jsonapi_pb.Schema{
			Type: &jsonapi_pb.Schema_ObjectItem{
				ObjectItem: &jsonapi_pb.ObjectItem{
					GoPackageName:   "Go Package",
					GoTypeName:      "Go Type",
					GrpcPackageName: "Grpc Package",

					ProtoFullName:    "long",
					ProtoMessageName: "short",

					Rules: &jsonapi_pb.ObjectRules{
						MinProperties: Ptr(uint64(1)),
						MaxProperties: Ptr(uint64(2)),
					},
					Properties: []*jsonapi_pb.ObjectProperty{{
						Name:             "foo",
						Required:         true,
						ProtoFieldName:   "foo",
						ProtoFieldNumber: 1,
						Schema: &jsonapi_pb.Schema{
							Type: &jsonapi_pb.Schema_StringItem{
								StringItem: &jsonapi_pb.StringItem{},
							},
						},
					}, {
						Name:             "bar",
						Required:         false,
						ProtoFieldName:   "bar",
						ProtoFieldNumber: 2,
						Schema: &jsonapi_pb.Schema{
							Type: &jsonapi_pb.Schema_StringItem{
								StringItem: &jsonapi_pb.StringItem{},
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
		name: "array",
		input: &jsonapi_pb.Schema{
			Type: &jsonapi_pb.Schema_ArrayItem{
				ArrayItem: &jsonapi_pb.ArrayItem{
					Items: &jsonapi_pb.Schema{
						Type: &jsonapi_pb.Schema_StringItem{
							StringItem: &jsonapi_pb.StringItem{},
						},
					},
					Rules: &jsonapi_pb.ArrayRules{
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
	}} {
		t.Run(tc.name, func(t *testing.T) {

			output, err := convertSchema(tc.input)
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
