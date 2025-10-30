package j5client

import (
	"github.com/pentops/golib/gl"
	"github.com/pentops/j5/gen/j5/auth/v1/auth_j5pb"
	"github.com/pentops/j5/gen/j5/client/v1/client_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
)

func wantAPI() *client_j5pb.API {
	objectRef := tObjectRef
	array := tArrayOf

	authJWT := &auth_j5pb.MethodAuthType{
		Type: &auth_j5pb.MethodAuthType_JwtBearer{
			JwtBearer: &auth_j5pb.MethodAuthType_JWTBearer{},
		},
	}
	authNone := &auth_j5pb.MethodAuthType{
		Type: &auth_j5pb.MethodAuthType_None_{
			None: &auth_j5pb.MethodAuthType_None{},
		},
	}

	getFoo := &client_j5pb.Method{
		Method: &schema_j5pb.Method{
			Name:         "GetFoo",
			Auth:         authNone,
			FullGrpcName: "/test.foo.v1.FooQueryService/GetFoo",
			HttpMethod:   schema_j5pb.HTTPMethod_GET,
			HttpPath:     "/test/v1/foo/:id",
			MethodType: &schema_j5pb.MethodType{
				Type: &schema_j5pb.MethodType_StateQuery_{
					StateQuery: &schema_j5pb.MethodType_StateQuery{
						EntityName: "test.foo.v1/foo",
						QueryPart:  schema_j5pb.StateQueryPart_GET,
					},
				},
			},
		},
		Request: &client_j5pb.Method_Request{
			PathParameters: []*schema_j5pb.ObjectProperty{{
				Name:     "id",
				Required: true,
				Schema: &schema_j5pb.Field{
					Type: &schema_j5pb.Field_String_{
						String_: &schema_j5pb.StringField{},
					},
				},
			}},
			QueryParameters: []*schema_j5pb.ObjectProperty{{
				Name: "number",
				Schema: &schema_j5pb.Field{
					Type: &schema_j5pb.Field_Integer{
						Integer: &schema_j5pb.IntegerField{
							Format: schema_j5pb.IntegerField_FORMAT_INT64,
						},
					},
				},
			}, {
				Name: "numbers",
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
				Name: "ab",
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
				Name: "multipleWord",
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
				Name:   "foo",
				Schema: objectRef("test.foo.v1", "FooState"),
			}},
		},
	}

	listFoos := &client_j5pb.Method{
		Method: &schema_j5pb.Method{
			Name:         "ListFoos",
			Auth:         authJWT,
			FullGrpcName: "/test.foo.v1.FooQueryService/ListFoos",
			HttpMethod:   schema_j5pb.HTTPMethod_GET,

			HttpPath: "/test/v1/foos",
			MethodType: &schema_j5pb.MethodType{
				Type: &schema_j5pb.MethodType_StateQuery_{
					StateQuery: &schema_j5pb.MethodType_StateQuery{
						EntityName: "test.foo.v1/foo",
						QueryPart:  schema_j5pb.StateQueryPart_LIST,
					},
				},
			},
		},
		Request: &client_j5pb.Method_Request{
			QueryParameters: []*schema_j5pb.ObjectProperty{{
				Name:   "page",
				Schema: objectRef("j5.list.v1", "PageRequest"),
			}, {
				Name:   "query",
				Schema: objectRef("j5.list.v1", "QueryRequest"),
			}},
			List: &client_j5pb.ListRequest{
				SearchableFields: []*client_j5pb.ListRequest_SearchField{{
					Name: "name",
				}, {
					Name: "bar.field",
				}},
				SortableFields: []*client_j5pb.ListRequest_SortField{{
					Name:        "createdAt",
					DefaultSort: gl.Ptr(client_j5pb.ListRequest_SortField_DIRECTION_DESC),
				}},
				FilterableFields: []*client_j5pb.ListRequest_FilterField{{
					Name: "fooId",
				}, {
					Name:           "status",
					DefaultFilters: []string{"ACTIVE"},
				}, {
					Name: "bar.id",
				}, {
					Name: "createdAt",
				}},
			},
		},
		ResponseBody: &schema_j5pb.Object{
			Name: "ListFoosResponse",
			Properties: []*schema_j5pb.ObjectProperty{{
				Name:   "foos",
				Schema: array(objectRef("test.foo.v1", "FooState")),
			}, {
				Name:   "page",
				Schema: objectRef("j5.list.v1", "PageResponse"),
			}},
		},
	}

	listFooEvents := &client_j5pb.Method{
		Method: &schema_j5pb.Method{
			Name:         "ListFooEvents",
			Auth:         authNone,
			FullGrpcName: "/test.foo.v1.FooQueryService/ListFooEvents",
			HttpMethod:   schema_j5pb.HTTPMethod_GET,
			HttpPath:     "/test/v1/foo/:id/events",
			MethodType: &schema_j5pb.MethodType{
				Type: &schema_j5pb.MethodType_StateQuery_{
					StateQuery: &schema_j5pb.MethodType_StateQuery{
						EntityName: "test.foo.v1/foo",
						QueryPart:  schema_j5pb.StateQueryPart_LIST_EVENTS,
					},
				},
			},
		},
		Request: &client_j5pb.Method_Request{
			PathParameters: []*schema_j5pb.ObjectProperty{{
				Name:     "id",
				Required: true,
				Schema: &schema_j5pb.Field{
					Type: &schema_j5pb.Field_Key{
						Key: &schema_j5pb.KeyField{
							Format: &schema_j5pb.KeyFormat{
								Type: &schema_j5pb.KeyFormat_Uuid{
									Uuid: &schema_j5pb.KeyFormat_UUID{},
								},
							},
						},
					},
				},
			}},
			QueryParameters: []*schema_j5pb.ObjectProperty{{
				Name:   "page",
				Schema: objectRef("j5.list.v1", "PageRequest"),
			}, {
				Name: "query",
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
				FilterableFields: []*client_j5pb.ListRequest_FilterField{{
					Name: "fooId",
				}, {
					Name: "timestamp",
				}},
				SortableFields: []*client_j5pb.ListRequest_SortField{{
					Name:        "timestamp",
					DefaultSort: gl.Ptr(client_j5pb.ListRequest_SortField_DIRECTION_DESC),
				}},
			},
		},
		ResponseBody: &schema_j5pb.Object{
			Name: "ListFooEventsResponse",
			Properties: []*schema_j5pb.ObjectProperty{{
				Name:   "events",
				Schema: array(objectRef("test.foo.v1", "FooEvent")),
			}, {
				Name:   "page",
				Schema: objectRef("j5.list.v1", "PageResponse"),
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
		Method: &schema_j5pb.Method{
			Name:         "PostFoo",
			Auth:         authJWT,
			FullGrpcName: "/test.foo.v1.FooCommandService/PostFoo",
			HttpMethod:   schema_j5pb.HTTPMethod_POST,
			HttpPath:     "/test/v1/foo",
		},
		Request: &client_j5pb.Method_Request{
			Body: &schema_j5pb.Object{
				Name: "PostFooRequest",
				Properties: []*schema_j5pb.ObjectProperty{{
					Name: "id",
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
				Name:   "foo",
				Schema: objectRef("test.foo.v1", "FooState"),
			}},
		},
	}

	fooCommandService := &client_j5pb.Service{
		Name: "FooCommandService",
		Methods: []*client_j5pb.Method{
			postFoo,
		},
	}

	downloadFoo := &client_j5pb.Method{
		Method: &schema_j5pb.Method{
			Name:         "DownloadRaw",
			FullGrpcName: "/test.foo.v1.FooDownloadService/DownloadRaw",
			HttpMethod:   schema_j5pb.HTTPMethod_GET,
			HttpPath:     "/test/v1/foo/:id/raw",
		},
		Request: &client_j5pb.Method_Request{
			PathParameters: []*schema_j5pb.ObjectProperty{{
				Name:     "id",
				Required: true,
				Schema: &schema_j5pb.Field{
					Type: &schema_j5pb.Field_String_{
						String_: &schema_j5pb.StringField{},
					},
				},
			}},
		},
		ResponseBody: nil,
	}

	fooDownloadService := &client_j5pb.Service{
		Name: "FooDownloadService",
		Methods: []*client_j5pb.Method{
			downloadFoo,
		},
	}

	return &client_j5pb.API{
		Packages: []*client_j5pb.Package{{
			Name: "test.foo.v1",

			Services: []*client_j5pb.Service{
				fooDownloadService,
			},
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
