package j5reflect

import (
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/gen/test/foo/v1/foo_testspb"
)

func TestTestListRequest(t *testing.T) {

	ss := NewSchemaSet()

	fooDesc := (&foo_testspb.ListFoosResponse{}).ProtoReflect().Descriptor()

	t.Log(protojson.Format(protodesc.ToDescriptorProto(fooDesc)))

	schemaItem, err := ss.SchemaReflect(fooDesc)
	if err != nil {
		t.Fatal(err.Error())
	}

	listRequest, err := buildListRequest(schemaItem)
	if err != nil {
		t.Fatal(err.Error())
	}

	want := &schema_j5pb.ListRequest{
		SearchableFields: []*schema_j5pb.ListRequest_SearchField{{
			Name: "name",
		}, {
			Name: "bar.field",
		}},
		FilterableFields: []*schema_j5pb.ListRequest_FilterField{{
			Name: "bar.id",
		}},
	}

	if !proto.Equal(listRequest, want) {
		t.Logf("got: %s", protojson.Format(listRequest))
		t.Logf("want: %s", protojson.Format(want))
		t.Fatal("List method did not return expected ListRequest")
	}

}
