package j5client

import (
	"github.com/pentops/golib/gl"
	"github.com/pentops/j5/gen/j5/auth/v1/auth_j5pb"
	"github.com/pentops/j5/gen/j5/client/v1/client_j5pb"
	"github.com/pentops/j5/gen/j5/list/v1/list_j5pb"
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

	getFoo := &client_j5pb.Method{
		Method: &schema_j5pb.Method{
			Name:         "FooGet",
			Auth:         authJWT,
			FullGrpcName: "/test.foo.v1.FooQueryService/FooGet",
			HttpMethod:   schema_j5pb.HTTPMethod_GET,
			HttpPath:     "/test/foo/v1/foo/q/:fooId",
			MethodType: &schema_j5pb.MethodType{
				Type: &schema_j5pb.MethodType_StateQuery_{
					StateQuery: &schema_j5pb.MethodType_StateQuery{
						EntityName: "foo",
						QueryPart:  schema_j5pb.StateQueryPart_GET,
					},
				},
			},
		},
		Request: &client_j5pb.Method_Request{
			PathParameters: []*schema_j5pb.ObjectProperty{{
				Name:     "fooId",
				Required: true,
				EntityKey: &schema_j5pb.EntityKey{
					Primary: true,
				},
				Schema: &schema_j5pb.Field{
					Type: &schema_j5pb.Field_Key{
						Key: &schema_j5pb.KeyField{
							Entity: &schema_j5pb.KeyField_DeprecatedEntityKey{
								Type: &schema_j5pb.KeyField_DeprecatedEntityKey_PrimaryKey{
									PrimaryKey: true,
								},
							},
							Format: &schema_j5pb.KeyFormat{
								Type: &schema_j5pb.KeyFormat_Uuid{
									Uuid: &schema_j5pb.KeyFormat_UUID{},
								},
							},
							ListRules: &list_j5pb.KeyRules{
								Filtering: &list_j5pb.FilteringConstraint{
									Filterable: true,
								},
							},
						},
					},
				},
			}},
			QueryParameters: []*schema_j5pb.ObjectProperty{},
		},
		ResponseBody: &schema_j5pb.Object{
			Name: "FooGetResponse",
			Properties: []*schema_j5pb.ObjectProperty{{
				Name:     "foo",
				Required: true,
				Schema:   objectRef("test.foo.v1", "FooState"),
			}},
		},
	}

	listFoos := &client_j5pb.Method{
		Method: &schema_j5pb.Method{
			Name:         "FooList",
			Auth:         authJWT,
			FullGrpcName: "/test.foo.v1.FooQueryService/FooList",
			HttpMethod:   schema_j5pb.HTTPMethod_GET,

			HttpPath: "/test/foo/v1/foo/q",
			MethodType: &schema_j5pb.MethodType{
				Type: &schema_j5pb.MethodType_StateQuery_{
					StateQuery: &schema_j5pb.MethodType_StateQuery{
						EntityName: "foo",
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
				SearchableFields: []*client_j5pb.ListRequest_SearchField{
					{
						Name: "data.name",
					},
					{
						Name: "data.bar.field",
					},
				},
				SortableFields: []*client_j5pb.ListRequest_SortField{
					{
						Name:        "metadata.createdAt",
						DefaultSort: gl.Ptr(client_j5pb.ListRequest_SortField_DIRECTION_DESC),
					},
					{
						Name: "metadata.updatedAt",
					},
					{
						Name:        "data.createdAt",
						DefaultSort: gl.Ptr(client_j5pb.ListRequest_SortField_DIRECTION_DESC),
					},
				},
				FilterableFields: []*client_j5pb.ListRequest_FilterField{
					{
						Name: "fooId",
					},
					{
						Name: "barId",
					},
					{
						Name: "data.bar.id",
					},
					{
						Name: "data.createdAt",
					},
					{
						Name:           "status",
						DefaultFilters: []string{"ACTIVE"},
					},
				},
			},
		},
		ResponseBody: &schema_j5pb.Object{
			Name: "FooListResponse",
			Properties: []*schema_j5pb.ObjectProperty{
				{
					Name:   "foo",
					Schema: array(objectRef("test.foo.v1", "FooState")),
				},
				{
					Name:   "page",
					Schema: objectRef("j5.list.v1", "PageResponse"),
				},
			},
		},
	}

	listFooEvents := &client_j5pb.Method{
		Method: &schema_j5pb.Method{
			Name:         "FooEvents",
			Auth:         authJWT,
			FullGrpcName: "/test.foo.v1.FooQueryService/FooEvents",
			HttpMethod:   schema_j5pb.HTTPMethod_GET,
			HttpPath:     "/test/foo/v1/foo/q/:fooId/events",
			MethodType: &schema_j5pb.MethodType{
				Type: &schema_j5pb.MethodType_StateQuery_{
					StateQuery: &schema_j5pb.MethodType_StateQuery{
						EntityName: "foo",
						QueryPart:  schema_j5pb.StateQueryPart_LIST_EVENTS,
					},
				},
			},
		},
		Request: &client_j5pb.Method_Request{
			PathParameters: []*schema_j5pb.ObjectProperty{{
				Name: "fooId",
				EntityKey: &schema_j5pb.EntityKey{
					Primary: true,
				},
				Required: true,
				Schema: &schema_j5pb.Field{
					Type: &schema_j5pb.Field_Key{
						Key: &schema_j5pb.KeyField{
							Format: &schema_j5pb.KeyFormat{
								Type: &schema_j5pb.KeyFormat_Uuid{
									Uuid: &schema_j5pb.KeyFormat_UUID{},
								},
							},
							Entity: &schema_j5pb.KeyField_DeprecatedEntityKey{
								Type: &schema_j5pb.KeyField_DeprecatedEntityKey_PrimaryKey{
									PrimaryKey: true,
								},
							},
							ListRules: &list_j5pb.KeyRules{
								Filtering: &list_j5pb.FilteringConstraint{
									Filterable: true,
								},
							},
						},
					},
				},
			}},
			QueryParameters: []*schema_j5pb.ObjectProperty{
				{
					Name:   "page",
					Schema: objectRef("j5.list.v1", "PageRequest"),
				},
				{
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
				},
			},
			List: &client_j5pb.ListRequest{
				FilterableFields: []*client_j5pb.ListRequest_FilterField{
					{
						Name: "metadata.timestamp",
					},
					{
						Name: "fooId",
					},
					{
						Name: "barId",
					},
					{
						Name: "event.!type",
					},
				},
				SortableFields: []*client_j5pb.ListRequest_SortField{
					{
						Name:        "metadata.timestamp",
						DefaultSort: gl.Ptr(client_j5pb.ListRequest_SortField_DIRECTION_DESC),
					},
				},
			},
		},
		ResponseBody: &schema_j5pb.Object{
			Name: "FooEventsResponse",
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
			HttpPath:     "/test/foo/v1/foo/c",
		},
		Request: &client_j5pb.Method_Request{
			Body: &schema_j5pb.Object{
				Name: "PostFooRequest",
				Properties: []*schema_j5pb.ObjectProperty{{
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
			HttpPath:     "/test/foo/v1/foo/:id/raw",
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
				Events: []*client_j5pb.StateEvent{
					{
						Name:        "created",
						FullName:    "test.foo.v1/foo.created",
						Description: "Comment on Created",
					},
					{
						Name:        "updated",
						FullName:    "test.foo.v1/foo.updated",
						Description: "Comment on Updated",
					},
				},
			}},
		}},
	}
}
