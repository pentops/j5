package j5client

import (
	"testing"

	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/reflect/protodesc"

	"github.com/pentops/flowtest/prototest"
	"github.com/pentops/golib/gl"
	"github.com/pentops/j5/gen/j5/client/v1/client_j5pb"
	"github.com/pentops/j5/internal/gen/test/foo/v1/foo_testspb"
	"github.com/pentops/j5/lib/j5schema"
)

func TestTestListRequest(t *testing.T) {
	ss := j5schema.NewSchemaCache()

	fooDesc := (&foo_testspb.FooListResponse{}).ProtoReflect().Descriptor()

	t.Log(prototext.Format(protodesc.ToDescriptorProto(fooDesc)))

	schemaItem, err := ss.Schema(fooDesc)
	if err != nil {
		t.Fatal(err.Error())
	}

	listRequest, err := buildListRequest(schemaItem)
	if err != nil {
		t.Fatal(err.Error())
	}

	want := &client_j5pb.ListRequest{
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
	}

	prototest.AssertEqualProto(t, want, listRequest)
}
