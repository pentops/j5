package j5convert

import (
	"testing"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/gen/j5/sourcedef/v1/sourcedef_j5pb"
	"github.com/pentops/j5/internal/j5s/protobuild/errset"
	"github.com/stretchr/testify/assert"
)

func testSummary(t *testing.T, elements ...*sourcedef_j5pb.RootElement) *FileSummary {
	ec := errset.NewCollector()
	summary, err := SourceSummary(&sourcedef_j5pb.SourceFile{
		Path:     "test/v1/test.j5s",
		Package:  &sourcedef_j5pb.Package{Name: "test.v1"},
		Elements: elements,
	}, ec)
	if err != nil {
		t.Fatalf("ConvertJ5File failed: %v", err)
	}
	if len(ec.Errors) > 0 {
		t.Fatalf("ConvertJ5File failed: %v", ec.Errors)
	}
	if len(ec.Warnings) > 0 {
		t.Errorf("ConvertJ5File warnings: %v", ec.Warnings)
	}
	return summary
}

func TestEnumSummary(t *testing.T) {
	enumSchema := &sourcedef_j5pb.RootElement{
		Type: &sourcedef_j5pb.RootElement_Enum{
			Enum: &schema_j5pb.Enum{
				Name:   "TestEnum",
				Prefix: "TEST_ENUM_",
				Options: []*schema_j5pb.Enum_Option{{
					Name: "UNSPECIFIED",
				}, {
					Name: "FOO",
				}, {
					Name: "BAR",
				}},
			},
		},
	}

	summary := testSummary(t, enumSchema)

	testEnum, ok := summary.Exports["TestEnum"]
	if !ok {
		t.Fatalf("Expected TestEnum in summary, got %v", summary)
	}

	assert.Equal(t, "TestEnum", testEnum.Name, "TypeRef.Name")
	assert.Equal(t, "test.v1", testEnum.Package, "TypeRef.Package")

	if testEnum.Enum == nil {
		t.Fatalf("Expected EnumRef in summary, got %v", testEnum)
	}

	assert.Equal(t, "TEST_ENUM_", testEnum.Enum.Prefix, "TypeRef.Prefix")
	assert.Equal(t, 3, len(testEnum.Enum.ValMap), "TypeRef.ValMap")
	t.Logf("ValMap: %v", testEnum.Enum.ValMap)
	want := map[string]int32{
		"UNSPECIFIED": 0,
		"FOO":         1,
		"BAR":         2,
	}
	for k, v := range want {
		if got, ok := testEnum.Enum.ValMap["TEST_ENUM_"+k]; !ok {
			t.Fatalf("Expected TEST_ENUM_%s in ValMap, got %v", k, testEnum.Enum.ValMap)
		} else if got != v {
			t.Fatalf("Expected TEST_ENUM_%s to be %d, got %d", k, v, got)
		}
	}
}

func TestPolymorphSummary(t *testing.T) {

	polymorphSchema := &sourcedef_j5pb.RootElement{
		Type: &sourcedef_j5pb.RootElement_Polymorph{
			Polymorph: &sourcedef_j5pb.Polymorph{
				Def: &schema_j5pb.Polymorph{
					Name:  "TestPolymorph",
					Types: []string{"foo.v1.Foo", "bar.v1.Bar"},
				},
				Includes: []string{"baz.v1.Baz"},
			},
		},
	}

	summary := testSummary(t, polymorphSchema)

	testPolymorph, ok := summary.Exports["TestPolymorph"]
	if !ok {
		t.Fatalf("Expected TestPolymorph in summary, got %v", summary)
	}

	assert.Equal(t, "TestPolymorph", testPolymorph.Name, "TypeRef.Name")
	assert.Equal(t, "test.v1", testPolymorph.Package, "TypeRef.Package")
	if testPolymorph.Polymorph == nil {
		t.Fatalf("Expected PolymorphRef in summary, got %v", testPolymorph)
	}

	assert.Equal(t, 2, len(testPolymorph.Polymorph.Types), "TypeRef.Types")
	assert.Equal(t, "foo.v1.Foo", testPolymorph.Polymorph.Types[0], "TypeRef.Types[0]")
	assert.Equal(t, "bar.v1.Bar", testPolymorph.Polymorph.Types[1], "TypeRef.Types[1]")

	assert.Equal(t, 1, len(testPolymorph.Polymorph.Includes), "TypeRef.Includes")
	assert.Equal(t, "baz.v1.Baz", testPolymorph.Polymorph.Includes[0], "TypeRef.Includes[0]")

}
