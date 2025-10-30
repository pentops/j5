package export

import (
	"encoding/json"
	"testing"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/tidwall/gjson"
)

func TestPathMap(t *testing.T) {
	dd := Document{
		Paths: PathSet{
			&PathItem{
				&Operation{
					OperationHeader: OperationHeader{
						Method:      "get",
						Path:        "/foo",
						OperationID: "test",
					},
				},
			},
		},
	}

	jsonVal, err := json.MarshalIndent(dd, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if val := gjson.GetBytes(jsonVal, "paths./foo.get.operationId").String(); val != "test" {
		t.Fatalf("expected operationId to be 'test', got %s", val)
	}

}

func TestPathFormat(t *testing.T) {
	path := "/foo/:bar/baz/:qux"
	formattedPath, err := formatPathParameters(path, []*schema_j5pb.ObjectProperty{
		{Name: "bar"},
		{Name: "qux"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if formattedPath != "/foo/{bar}/baz/{qux}" {
		t.Fatalf("expected %q, got %q", "/foo/{bar}/baz/{qux}", formattedPath)
	}
}
