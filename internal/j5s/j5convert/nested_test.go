package j5convert

import (
	"testing"

	"github.com/pentops/flowtest/prototest"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/gen/j5/sourcedef/v1/sourcedef_j5pb"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestMessageNesting(t *testing.T) {

	wantFile := &descriptorpb.FileDescriptorProto{
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{
			//GoPackage: proto.String("github.com/pentops/j5/test/v1/test_pb"),
		},
		Dependency: []string{"j5/ext/v1/annotations.proto"},
		Name:       proto.String("test/v1/test.j5s.proto"),
		Package:    proto.String("test.v1"),
		MessageType: []*descriptorpb.DescriptorProto{{
			Name:    proto.String("Outer"),
			Options: emptyObjectOption,
			Field: []*descriptorpb.FieldDescriptorProto{{
				Name:     proto.String("field"),
				JsonName: proto.String("field"),
				Number:   proto.Int32(1),
				Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
				TypeName: proto.String("Outer.Inner"),
				Options:  tEmptyTypeExt(t, "object"),
			}},

			NestedType: []*descriptorpb.DescriptorProto{{
				Name:    proto.String("Inner"),
				Options: emptyObjectOption,
				Field: []*descriptorpb.FieldDescriptorProto{{
					JsonName: proto.String("innerField"),
					Name:     proto.String("inner_field"),
					Number:   proto.Int32(1),
					Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					Options:  tEmptyTypeExt(t, "string"),
				}},
			}},
		}},
	}

	deps := &testDeps{
		pkg: "test.v1",
	}

	innerObject := &schema_j5pb.Object{
		Name: "Inner",
		Properties: []*schema_j5pb.ObjectProperty{{
			Name: "innerField",
			Schema: &schema_j5pb.Field{
				Type: &schema_j5pb.Field_String_{
					String_: &schema_j5pb.StringField{},
				},
			},
		}},
	}

	outerObject := &schema_j5pb.Object{
		Name: "Outer",
		Properties: []*schema_j5pb.ObjectProperty{{
			Name: "field",
			Schema: &schema_j5pb.Field{
				Type: &schema_j5pb.Field_Object{
					Object: &schema_j5pb.ObjectField{
						Schema: &schema_j5pb.ObjectField_Object{
							Object: innerObject,
						},
					},
				},
			},
		}},
	}

	gotFiles, err := ConvertJ5File(deps, &sourcedef_j5pb.SourceFile{
		Path:    "test/v1/test.j5s",
		Package: &sourcedef_j5pb.Package{Name: "test.v1"},
		Elements: []*sourcedef_j5pb.RootElement{{
			Type: &sourcedef_j5pb.RootElement_Object{
				Object: &sourcedef_j5pb.Object{
					Def: outerObject,
				},
			},
		}},
	})
	if err != nil {
		t.Fatalf("ConvertJ5File failed: %v", err)
	}
	gotFile := gotFiles[0]

	gotFile.SourceCodeInfo = nil
	t.Log(prototext.Format(gotFile))
	prototest.AssertEqualProto(t, wantFile, gotFile)

}
