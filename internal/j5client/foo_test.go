package j5client

import (
	"context"
	"os"
	"testing"

	"github.com/pentops/j5/gen/j5/client/v1/client_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/internal/source"
	"github.com/pentops/j5/internal/structure"
	"github.com/pentops/j5/internal/testlib"
)

func TestFooSchema(t *testing.T) {

	ctx := context.Background()
	rootFS := os.DirFS("../../")
	thisRoot, err := source.ReadLocalSource(ctx, rootFS)
	if err != nil {
		t.Fatalf("ReadLocalSource: %v", err)
	}

	input, err := thisRoot.NamedInput("test")
	if err != nil {
		t.Fatalf("NamedInput: %v", err)
	}

	srcImg, err := input.SourceImage(ctx)
	if err != nil {
		t.Fatalf("SourceImage: %v", err)
	}

	sourceAPI, err := structure.APIFromImage(srcImg)
	if err != nil {
		t.Fatalf("APIFromImage: %v", err)
	}

	for _, pkg := range sourceAPI.Packages {
		t.Logf("Package: %s", pkg.Name)
		for name := range pkg.Schemas {
			t.Logf("Schema: %s", name)
		}
	}

	clientAPI, err := APIFromSource(sourceAPI)
	if err != nil {
		t.Fatalf("APIFromSource: %v", err)
	}

	t.Logf("ClientAPI: %v", clientAPI)

	want := wantAPI().Packages[0]

	got := clientAPI.Packages[0]
	got.Schemas = nil

	testlib.AssertEqualProto(t, want, got)
}

func wantAPI() *client_j5pb.API {

	objectRef := func(pkg, schema string) *schema_j5pb.Field {
		return &schema_j5pb.Field{
			Type: &schema_j5pb.Field_Object{
				Object: &schema_j5pb.ObjectField{
					Schema: &schema_j5pb.ObjectField_Ref{
						Ref: &schema_j5pb.Ref{
							Package: pkg,
							Schema:  schema,
						},
					},
				},
			},
		}
	}

	array := func(of *schema_j5pb.Field) *schema_j5pb.Field {
		return &schema_j5pb.Field{
			Type: &schema_j5pb.Field_Array{
				Array: &schema_j5pb.ArrayField{
					Items: of,
				},
			},
		}
	}

	getFoo := &client_j5pb.Method{
		Name:         "GetFoo",
		FullGrpcName: "/test.foo.v1.FooQueryService/GetFoo",
		HttpMethod:   client_j5pb.HTTPMethod_GET,
		HttpPath:     "/test/v1/foo/:id",
		Request: &client_j5pb.Method_Request{
			PathParameters: []*schema_j5pb.ObjectProperty{{
				Name:       "id",
				ProtoField: []int32{1},
				Schema: &schema_j5pb.Field{
					Type: &schema_j5pb.Field_String_{
						String_: &schema_j5pb.StringField{},
					},
				},
			}},
			QueryParameters: []*schema_j5pb.ObjectProperty{{
				Name:       "number",
				ProtoField: []int32{2},
				Schema: &schema_j5pb.Field{
					Type: &schema_j5pb.Field_Integer{
						Integer: &schema_j5pb.IntegerField{
							Format: schema_j5pb.IntegerField_FORMAT_INT64,
						},
					},
				},
			}, {
				Name:       "numbers",
				ProtoField: []int32{3},
				Schema: &schema_j5pb.Field{
					Type: &schema_j5pb.Field_Array{
						Array: &schema_j5pb.ArrayField{
							Items: &schema_j5pb.Field{
								Type: &schema_j5pb.Field_Float{
									Float: &schema_j5pb.FloatField{
										Format: schema_j5pb.FloatField_FORMAT_FLOAT32,
									},
								},
							},
						},
					},
				},
			}, {
				Name:       "ab",
				ProtoField: []int32{4},
				Schema: &schema_j5pb.Field{
					Type: &schema_j5pb.Field_Object{
						Object: &schema_j5pb.ObjectField{
							Schema: &schema_j5pb.ObjectField_Ref{
								Ref: &schema_j5pb.Ref{
									Package: "test.foo.v1.service",
									Schema:  "ABMessage",
								},
							},
						},
					},
				},
			}, {
				Name:       "multipleWord",
				ProtoField: []int32{5},
				Schema: &schema_j5pb.Field{
					Type: &schema_j5pb.Field_String_{
						String_: &schema_j5pb.StringField{},
					},
				},
			}},
		},
		ResponseBody: &schema_j5pb.Object{
			Name: "GetFooResponse",
			Properties: []*schema_j5pb.ObjectProperty{{
				Name:       "foo",
				ProtoField: []int32{1},
				Schema:     objectRef("test.foo.v1", "FooState"),
			}},
		},
	}

	listFoos := &client_j5pb.Method{
		Name:         "ListFoos",
		FullGrpcName: "/test.foo.v1.FooQueryService/ListFoos",
		HttpMethod:   client_j5pb.HTTPMethod_GET,

		HttpPath: "/test/v1/foos",
		Request: &client_j5pb.Method_Request{
			QueryParameters: []*schema_j5pb.ObjectProperty{{
				Name:       "page",
				ProtoField: []int32{100},
				Schema:     objectRef("j5.list.v1", "PageRequest"),
			}, {
				Name:       "query",
				ProtoField: []int32{101},
				Schema:     objectRef("j5.list.v1", "QueryRequest"),
			}},
			List: &client_j5pb.ListRequest{
				SearchableFields: []*client_j5pb.ListRequest_SearchField{{
					Name: "name",
				}, {
					Name: "bar.field",
				}},
				SortableFields: []*client_j5pb.ListRequest_SortField{{
					Name: "createdAt",
				}},
				FilterableFields: []*client_j5pb.ListRequest_FilterField{{
					Name:           "status",
					DefaultFilters: []string{"ACTIVE"},
				}, {
					Name: "createdAt",
				}, {
					Name: "bar.id",
				}},
			},
		},
		ResponseBody: &schema_j5pb.Object{
			Name: "ListFoosResponse",
			Properties: []*schema_j5pb.ObjectProperty{{
				Name:       "foos",
				ProtoField: []int32{1},
				Schema:     array(objectRef("test.foo.v1", "FooState")),
			}},
		},
	}

	listFooEvents := &client_j5pb.Method{
		Name:         "ListFooEvents",
		FullGrpcName: "/test.foo.v1.FooQueryService/ListFooEvents",
		HttpMethod:   client_j5pb.HTTPMethod_GET,
		HttpPath:     "/test/v1/foo/:id/events",
		Request: &client_j5pb.Method_Request{
			PathParameters: []*schema_j5pb.ObjectProperty{{
				Name:       "id",
				ProtoField: []int32{1},
				Schema: &schema_j5pb.Field{
					Type: &schema_j5pb.Field_Key{
						Key: &schema_j5pb.KeyField{
							Format: schema_j5pb.KeyFormat_KEY_FORMAT_UUID,
						},
					},
				},
			}},
			QueryParameters: []*schema_j5pb.ObjectProperty{{
				Name:       "page",
				ProtoField: []int32{100},
				Schema:     objectRef("j5.list.v1", "PageRequest"),
			}, {
				Name:       "query",
				ProtoField: []int32{101},
				Schema: &schema_j5pb.Field{
					Type: &schema_j5pb.Field_Object{
						Object: &schema_j5pb.ObjectField{
							Schema: &schema_j5pb.ObjectField_Ref{
								Ref: &schema_j5pb.Ref{
									Package: "j5.list.v1",
									Schema:  "QueryRequest",
								},
							},
						},
					},
				},
			}},
			List: &client_j5pb.ListRequest{
				// empty object because it is a list, but no fields.
			},
		},
		ResponseBody: &schema_j5pb.Object{
			Name: "ListFooEventsResponse",
			Properties: []*schema_j5pb.ObjectProperty{{
				Name:       "events",
				ProtoField: []int32{1},
				Schema:     array(objectRef("test.foo.v1", "FooEvent")),
			}},
		},
	}

	fooQueryService := &client_j5pb.Service{
		Name: "FooQueryService",
		Methods: []*client_j5pb.Method{
			getFoo,
			listFoos,
			listFooEvents,
		},
	}

	postFoo := &client_j5pb.Method{
		Name:         "PostFoo",
		FullGrpcName: "/test.foo.v1.FooCommandService/PostFoo",
		HttpMethod:   client_j5pb.HTTPMethod_POST,
		HttpPath:     "/test/v1/foo",
		Request: &client_j5pb.Method_Request{
			Body: &schema_j5pb.Object{
				Name: "PostFooRequest",
				Properties: []*schema_j5pb.ObjectProperty{{
					Name:       "id",
					ProtoField: []int32{1},
					Schema: &schema_j5pb.Field{
						Type: &schema_j5pb.Field_String_{
							String_: &schema_j5pb.StringField{},
						},
					},
				}},
			},
		},
		ResponseBody: &schema_j5pb.Object{
			Name: "PostFooResponse",
			Properties: []*schema_j5pb.ObjectProperty{{
				Name:       "foo",
				ProtoField: []int32{1},
				Schema:     objectRef("test.foo.v1", "FooState"),
			}},
		},
	}

	fooCommandService := &client_j5pb.Service{
		Name: "FooCommandService",
		Methods: []*client_j5pb.Method{
			postFoo,
		},
	}

	return &client_j5pb.API{
		Packages: []*client_j5pb.Package{{
			Name: "test.foo.v1",

			Services: []*client_j5pb.Service{},
			StateEntities: []*client_j5pb.StateEntity{{
				Name:         "foo",
				FullName:     "test.foo.v1/foo",
				SchemaName:   "test.foo.v1.FooState",
				PrimaryKey:   []string{"fooId"},
				QueryService: fooQueryService,
				CommandServices: []*client_j5pb.Service{
					fooCommandService,
				},
				Events: []*client_j5pb.StateEvent{{
					Name:        "created",
					FullName:    "test.foo.v1/foo.created",
					Description: "Comment on Created",
				}, {
					Name:        "updated",
					FullName:    "test.foo.v1/foo.updated",
					Description: "Comment on Updated",
				}},
			}},
		}},
	}
}
