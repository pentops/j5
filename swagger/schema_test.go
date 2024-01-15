package swagger

import (
	"encoding/json"
	"testing"

	"github.com/pentops/jsonapi/gen/v1/jsonapi_pb"
	"github.com/pentops/jsonapi/jsontest"
)

func TestConvertSchema(t *testing.T) {

	input := jsonapi_pb.Schema{}
	output, err := convertSchema(&input)
	if err != nil {
		t.Fatal(err)
	}

	jsonOutput, err := json.Marshal(output)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(jsonOutput))
}

func TestSchemaJSONMarshal(t *testing.T) {

	object := &Schema{
		SchemaItem: &SchemaItem{
			Type: &ObjectItem{
				Required: []string{"id"},
				Properties: map[string]*ObjectProperty{
					"id": {
						Schema: Schema{
							SchemaItem: &SchemaItem{
								Description: "desc",
								Type: StringItem{
									Format: "uuid",
								},
							},
						},
					},
					"number": {
						Schema: Schema{
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
						Schema: Schema{
							SchemaItem: &SchemaItem{
								Type: &ObjectItem{
									Required: []string{"foo"},
									Properties: map[string]*ObjectProperty{
										"foo": {
											Schema: Schema{
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
						Schema: Schema{
							Ref: Ptr("#/definitions/foo"),
						},
					},
					"oneof": {
						Schema: Schema{
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
