package j5convert

import (
	"testing"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/gen/j5/sourcedef/v1/sourcedef_j5pb"
	"github.com/pentops/j5/internal/j5s/protobuild/errset"
)

func TestSummary(t *testing.T) {
	enumSchema := &sourcedef_j5pb.RootElement{
		Type: &sourcedef_j5pb.RootElement_Enum{
			Enum: &schema_j5pb.Enum{
				Name:   "TestEnum",
				Prefix: "TEST_ENUM_",
				Options: []*schema_j5pb.Enum_Option{{
					Name:   "FOO",
					Number: 1,
				}},
			},
		},
	}

	ec := errset.NewCollector()
	summary, err := SourceSummary(&sourcedef_j5pb.SourceFile{
		Path:     "test/v1/test.j5s",
		Package:  &sourcedef_j5pb.Package{Name: "test.v1"},
		Elements: []*sourcedef_j5pb.RootElement{enumSchema},
	}, ec)
	if err != nil {
		t.Fatalf("ConvertJ5File failed: %v", err)
	}

	testEnum, ok := summary.Exports["TestEnum"]
	if !ok {
		t.Fatalf("Expected TestEnum in summary, got %v", summary)
	}

	if testEnum.Name != "TestEnum" {
		t.Fatalf("Expected TestEnum name to be 'TestEnum', got %s", testEnum.Name)
	}
}
