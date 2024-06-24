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
			Type: &schema_j5pb.Schema_String_{
				String_: &schema_j5pb.String{
					Format:  Ptr("uuid"),
					Example: Ptr("example"),
					Rules: &schema_j5pb.String_Rules{
						Pattern:   Ptr("regex-pattern"),
						MinLength: Ptr(uint64(1)),
						MaxLength: Ptr(uint64(2)),
					},
				},
			},
		},
		want: map[string]interface{}{
			"type":      "string",
			"example":   "example",
			"format":    "uuid",
			"pattern":   "regex-pattern",
			"minLength": 1,
			"maxLength": 2,
		},
	}, {
		name: "number",
		input: &schema_j5pb.Schema{
			Type: &schema_j5pb.Schema_Float{
				Float: &schema_j5pb.Float{
					Format: schema_j5pb.Float_FORMAT_FLOAT64,
					Rules: &schema_j5pb.Float_Rules{
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
			Type: &schema_j5pb.Schema_Enum{
				Enum: &schema_j5pb.Enum{
					Options: []*schema_j5pb.Enum_Value{{
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
				Ref: &schema_j5pb.Ref{
					Package: "package.v1",
					Schema:  "Foo",
				},
			},
		},
		want: map[string]interface{}{
			"$ref": "#/definitions/package.v1.Foo",
		},
	}, {
		name: "object",
		input: &schema_j5pb.Schema{
			Type: &schema_j5pb.Schema_Object{
				Object: &schema_j5pb.Object{
					Name:        "short",
					Description: "description",
					Rules: &schema_j5pb.Object_Rules{
						MinProperties: Ptr(uint64(1)),
						MaxProperties: Ptr(uint64(2)),
					},
					Properties: []*schema_j5pb.ObjectProperty{{
						Name:     "foo",
						Required: true,
						Schema: &schema_j5pb.Schema{
							Type: &schema_j5pb.Schema_String_{
								String_: &schema_j5pb.String{},
							},
						},
					}, {
						Name:     "bar",
						Required: false,
						Schema: &schema_j5pb.Schema{
							Type: &schema_j5pb.Schema_String_{
								String_: &schema_j5pb.String{},
							},
						},
					}},
				},
			},
		},
		want: map[string]interface{}{
			"type":                "object",
			"description":         "description",
			"x-name":              "short",
			"required.0":          "foo",
			"properties.foo.type": "string",
			"properties.bar.type": "string",
		},
	}, {
		name: "oneof",
		input: &schema_j5pb.Schema{
			Type: &schema_j5pb.Schema_Oneof{
				Oneof: &schema_j5pb.Oneof{
					Name: "short",

					Properties: []*schema_j5pb.ObjectProperty{{
						Name: "foo",
						Schema: &schema_j5pb.Schema{
							Type: &schema_j5pb.Schema_String_{
								String_: &schema_j5pb.String{},
							},
						},
					}, {
						Name: "bar",
						Schema: &schema_j5pb.Schema{
							Type: &schema_j5pb.Schema_String_{
								String_: &schema_j5pb.String{},
							},
						},
					}},
				},
			},
		},
		want: map[string]interface{}{
			"type":                "object",
			"x-name":              "short",
			"properties.foo.type": "string",
			"properties.bar.type": "string",
		},
	}, {
		name: "array",
		input: &schema_j5pb.Schema{
			Type: &schema_j5pb.Schema_Array{
				Array: &schema_j5pb.Array{
					Items: &schema_j5pb.Schema{
						Type: &schema_j5pb.Schema_String_{
							String_: &schema_j5pb.String{},
						},
					},
					Rules: &schema_j5pb.Array_Rules{
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
			Type: &schema_j5pb.Schema_Map{
				Map: &schema_j5pb.Map{
					ItemSchema: &schema_j5pb.Schema{
						Type: &schema_j5pb.Schema_String_{
							String_: &schema_j5pb.String{},
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
				Any: &schema_j5pb.Any{},
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
								Type: StringItem{
									Format: Some("uuid"),
								},
							},
						},
					},
					"number": {
						Schema: &Schema{
							SchemaItem: &SchemaItem{
								Type: FloatItem{
									Format:  "double",
									Minimum: Value(0.0),
									Maximum: Value(100.0),
								},
							},
						},
					},
					"namedObject": {
						Schema: &Schema{
							SchemaItem: &SchemaItem{
								Type: &ObjectItem{
									Name:        "namedObject",
									Description: "desc",
									Required:    []string{"foo"},
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
	out.AssertEqual(t, "required.0", "id")

	out.AssertEqual(t, "properties.number.type", "number")
	out.AssertEqual(t, "properties.number.format", "double")
	out.AssertEqual(t, "properties.number.minimum", 0.0)
	out.AssertEqual(t, "properties.number.maximum", 100.0)
	out.AssertNotSet(t, "properties.number.exclusiveMinimum")

	out.AssertEqual(t, "properties.namedObject.properties.foo.type", "string")
	out.AssertEqual(t, "properties.namedObject.x-name", "namedObject")
	out.AssertEqual(t, "properties.namedObject.description", "desc")

	out.AssertEqual(t, "properties.ref.$ref", "#/definitions/foo")
}
