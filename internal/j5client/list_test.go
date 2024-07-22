package j5client

import (
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"

	"github.com/pentops/j5/gen/j5/client/v1/client_j5pb"
	"github.com/pentops/j5/gen/test/foo/v1/foo_testspb"
	"github.com/pentops/j5/internal/j5reflect"
)

func TestTestListRequest(t *testing.T) {

	ss := j5reflect.NewSchemaCache()

	fooDesc := (&foo_testspb.ListFoosResponse{}).ProtoReflect().Descriptor()

	t.Log(protojson.Format(protodesc.ToDescriptorProto(fooDesc)))

	schemaItem, err := ss.Schema(fooDesc)
	if err != nil {
		t.Fatal(err.Error())
	}

	listRequest, err := buildListRequest(schemaItem)
	if err != nil {
		t.Fatal(err.Error())
	}

	want := &client_j5pb.ListRequest{
		SearchableFields: []*client_j5pb.ListRequest_SearchField{{
			Name: "name",
		}, {
			Name: "bar.field",
		}},
		FilterableFields: []*client_j5pb.ListRequest_FilterField{{
			Name: "bar.id",
		}},
	}

	if !proto.Equal(listRequest, want) {
		t.Logf("got: %s", protojson.Format(listRequest))
		t.Logf("want: %s", protojson.Format(want))
		t.Fatal("List method did not return expected ListRequest")
	}

}