package j5convert

import (
	"testing"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/pentops/golib/gl"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/gen/j5/sourcedef/v1/sourcedef_j5pb"
	"github.com/pentops/j5/internal/j5s/protobuild/errset"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestEnum(t *testing.T) {
	wantFile := &descriptorpb.FileDescriptorProto{
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{
			//GoPackage: proto.String("github.com/pentops/j5/test/v1/test_pb"),
		},
		Name:    proto.String("test/v1/test.j5s.proto"),
		Package: proto.String("test.v1"),
		EnumType: []*descriptorpb.EnumDescriptorProto{{
			Name: proto.String("TestEnum"),
			Value: []*descriptorpb.EnumValueDescriptorProto{{
				Name:   proto.String("TEST_ENUM_UNSPECIFIED"),
				Number: proto.Int32(0),
			}, {
				Name:   proto.String("TEST_ENUM_FOO"),
				Number: proto.Int32(1),
			}, {
				Name:   proto.String("TEST_ENUM_BAR"),
				Number: proto.Int32(2),
			}},
		}},
	}

	run := func(name string, schema *sourcedef_j5pb.RootElement) {
		t.Run(name, func(t *testing.T) {
			deps := &testDeps{
				pkg: "test.v1",
			}
			sourceFile := &sourcedef_j5pb.SourceFile{
				Path:     "test/v1/test.j5s",
				Package:  &sourcedef_j5pb.Package{Name: "test.v1"},
				Elements: []*sourcedef_j5pb.RootElement{schema},
			}
			gotFiles, err := ConvertJ5File(deps, sourceFile)
			if err != nil {
				t.Fatalf("ConvertJ5File failed: %v", err)
			}

			gotFile := gotFiles[0]
			gotFile.SourceCodeInfo = nil
			equal(t, wantFile, gotFile)

			ec := errset.NewCollector()
			summary, err := SourceSummary(sourceFile, ec)
			if err != nil {
				t.Fatalf("ConvertJ5File failed: %v", err)
			}

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
		})
	}

	run("mixed", &sourcedef_j5pb.RootElement{
		Type: &sourcedef_j5pb.RootElement_Enum{
			Enum: &schema_j5pb.Enum{
				Name:   "TestEnum",
				Prefix: "TEST_ENUM_",
				Options: []*schema_j5pb.Enum_Option{{
					Name: "TEST_ENUM_UNSPECIFIED",
				}, {
					Name: "FOO",
				}, {
					Name: "TEST_ENUM_BAR",
				}},
			},
		},
	})
}

func TestImportEnum(t *testing.T) {

	deps := &testDeps{
		pkg: "test.v1",
		types: map[string]*TypeRef{
			"test.v1.TestEnum": {
				Package: "test.v1",
				Name:    "TestEnum",
				File:    "test/v1/test_enum.proto",
				Enum: &EnumRef{
					Prefix: "TEST_ENUM_",
					ValMap: map[string]int32{
						"TEST_ENUM_FOO": 1,
						"TEST_ENUM_BAR": 2,
					},
				},
			},
		},
	}

	enumField := &schema_j5pb.EnumField{
		Schema: &schema_j5pb.EnumField_Ref{
			Ref: &schema_j5pb.Ref{
				Package: "",
				Schema:  "TestEnum",
			},
		},
		Rules: &schema_j5pb.EnumField_Rules{
			In: []string{"FOO", "BAR"},
		},
	}

	wantField := &descriptorpb.FieldDescriptorProto{
		Name:     proto.String("enum"),
		Type:     descriptorpb.FieldDescriptorProto_TYPE_ENUM.Enum(),
		Number:   proto.Int32(1),
		TypeName: proto.String(".test.v1.TestEnum"),
		Options: withOption(tEmptyTypeExt(t, "enum"), validate.E_Field, &validate.FieldConstraints{

			Type: &validate.FieldConstraints_Enum{
				Enum: &validate.EnumRules{
					DefinedOnly: gl.Ptr(true),
					In:          []int32{1, 2},
				},
			},
		}),
		JsonName: proto.String("enum"),
	}

	inputProp := &schema_j5pb.ObjectProperty{
		Name: "enum",
		Schema: &schema_j5pb.Field{
			Type: &schema_j5pb.Field_Enum{
				Enum: enumField,
			},
		},
		ProtoField: []int32{3},
	}

	runField(t, deps, inputProp, wantField)

}
