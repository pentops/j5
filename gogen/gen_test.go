package gogen

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"

	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/gen/test/foo/v1/foo_testpb"
	"github.com/pentops/j5/schema/structure"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type TestOutput struct {
	Files map[string]string
}

func (o TestOutput) WriteFile(name string, data []byte) error {
	if _, ok := o.Files[name]; ok {
		return fmt.Errorf("file %q already exists", name)
	}
	fmt.Printf("writing file %q\n", name)
	o.Files[name] = string(data)
	return nil
}

func TestTestProtoGen(t *testing.T) {

	ctx := context.Background()
	ss := structure.NewSchemaSet(&config_j5pb.CodecOptions{})

	mustBuildSchema := func(desc protoreflect.MessageDescriptor) *schema_j5pb.Schema {
		schemaItem, err := ss.BuildSchemaObject(ctx, desc)
		if err != nil {
			t.Fatal(err.Error())
		}
		return schemaItem

	}

	jdef := &schema_j5pb.API{
		Metadata: &schema_j5pb.Metadata{
			BuiltAt: timestamppb.Now(),
		},
		Packages: []*schema_j5pb.Package{{
			Label:        "package label",
			Name:         "test.v1",
			Hidden:       false,
			Introduction: "FOOBAR",
			Methods: []*schema_j5pb.Method{{
				GrpcServiceName: "TestService",
				FullGrpcName:    "/test.v1.TestService/Test",
				GrpcMethodName:  "PostFoo",
				HttpMethod:      "get",
				HttpPath:        "/test/v1/foo",
				ResponseBody:    mustBuildSchema((&foo_testpb.PostFooRequest{}).ProtoReflect().Descriptor()),
				RequestBody:     mustBuildSchema((&foo_testpb.PostFooRequest{}).ProtoReflect().Descriptor()),
			}},
		}},
		Schemas: ss.Schemas,
	}

	output := TestOutput{
		Files: map[string]string{},
	}

	options := Options{
		TrimPackagePrefix: "",
		AddGoPrefix:       "github.com/pentops/j5/testproto/clientgen",
	}

	if err := WriteGoCode(jdef, output, options); err != nil {
		t.Fatal(err)
	}

	outFile, ok := output.Files["/test/v1/test/generated.go"]
	if !ok {
		t.Fatal("file test/v1/generated.go not found")
	}

	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, "", outFile, 0)
	if err != nil {
		t.Fatal(err)
	}

	structTypes := map[string]*ast.StructType{}

	for _, decl := range parsed.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			t.Logf("func: %#v", decl.Name.Name)
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch spec := spec.(type) {
				case *ast.TypeSpec:
					t.Logf("type: %#v", spec.Name.Name)
					switch specType := spec.Type.(type) {
					case *ast.StructType:
						structTypes[spec.Name.Name] = specType
					}
				}
			}
		}
	}

	posString := func(thing interface {
		Pos() token.Pos
		End() token.Pos
	}) string {
		return outFile[fset.Position(thing.Pos()).Offset:fset.Position(thing.End()).Offset]
	}

	assertField := func(typeName string, name string, wantTypeName, wantTag string) {
		structType, ok := structTypes[typeName]
		if !ok {
			t.Fatalf("type %q not found", typeName)
		}

		for _, field := range structType.Fields.List {
			for _, fieldName := range field.Names {
				if fieldName.Name == name {
					gotTypeName := posString(field.Type)
					assert.Equal(t, wantTypeName, gotTypeName, "field %q", name)

					gotTag := field.Tag.Value

					assert.Equal(t, "`"+wantTag+"`", gotTag, "field %q tag:", name)
					return
				}
			}
		}
	}

	assertField("PostFooRequest", "SString", "string", `json:"sString,omitempty"`)
	assertField("PostFooRequest", "OString", "*string", `json:"oString,omitempty"`)
	assertField("PostFooRequest", "RString", "[]string", `json:"rString,omitempty"`)
	assertField("PostFooRequest", "MapStringString", "map[string]string", `json:"mapStringString,omitempty"`)

}
