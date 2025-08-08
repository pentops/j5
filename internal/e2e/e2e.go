package e2e

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pentops/j5/gen/j5/client/v1/client_j5pb"
	"github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/gen/j5/sourcedef/v1/sourcedef_j5pb"
	"github.com/pentops/j5/internal/j5client"
	"github.com/pentops/j5/internal/j5s/j5convert"
	"github.com/pentops/j5/internal/j5s/protobuild"
	"github.com/pentops/j5/internal/j5s/protobuild/errset"
	"github.com/pentops/j5/internal/j5s/protobuild/psrc"
	"github.com/pentops/j5/internal/source/image"
	"github.com/pentops/j5/internal/structure"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type FileBuilder struct {
	*sourcedef_j5pb.SourceFile
}

func NewFileBuilder(name string) *FileBuilder {
	packageName := strings.ReplaceAll(path.Dir(name), "/", ".")
	return &FileBuilder{
		SourceFile: &sourcedef_j5pb.SourceFile{
			Path: name,
			Package: &sourcedef_j5pb.Package{
				Name: packageName,
			},
		},
	}
}

func (fb *FileBuilder) BuildPackage(t testing.TB) *protobuild.BuiltPackage {
	deps := psrc.NewBuiltinResolver()
	ps, err := protobuild.NewPackageSet(deps, fb)
	if err != nil {
		t.Fatalf("FATAL: Unexpected error: %s", err.Error())
	}

	pkg, err := ps.CompilePackage(t.Context(), fb.Package.Name)
	if err != nil {
		t.Fatalf("FATAL: Unexpected error: %s", err.Error())
	}

	return pkg
}

func (fb *FileBuilder) ToDescriptors(t testing.TB) protoreflect.FileDescriptor {
	pkg := fb.BuildPackage(t)
	if len(pkg.Proto) != 1 {
		t.Fatalf("FATAL: Expected exactly one proto file, got %d", len(pkg.Proto))
	}
	fd, err := protodesc.NewFile(pkg.Proto[0].Desc, protoregistry.GlobalFiles)
	if err != nil {
		t.Fatal(fmt.Errorf("FATAL: Failed to create file descriptor: %w", err))
	}

	return fd
}

func (fb *FileBuilder) BuildImage(t testing.TB) *source_j5pb.SourceImage {
	fd := fb.BuildPackage(t)

	builder := image.NewBuilder()
	builder.AddPackage(&source_j5pb.Package{
		Name: "test.v1",
	})
	if err := builder.AddBuilt(fd, &ext_j5pb.J5Source{}); err != nil {
		t.Fatalf("FATAL: Unexpected error: %s", err.Error())
	}

	return builder.Image()
}

func (fb *FileBuilder) BuildStructure(t testing.TB) *schema_j5pb.API {
	img := fb.BuildImage(t)
	reflectionAPI, err := structure.APIFromImage(img)
	if err != nil {
		t.Fatal(err.Error())
	}
	return reflectionAPI
}

func (fb *FileBuilder) BuildClientAPI(t testing.TB) *client_j5pb.API {
	api := fb.BuildStructure(t)
	clientAPI, err := j5client.APIFromSource(api)
	if err != nil {
		t.Fatalf("FATAL: Unexpected error: %s", err.Error())
	}
	return clientAPI
}

func (fb *FileBuilder) ListPackages() []string {
	return []string{fb.Package.Name}
}

func (fb *FileBuilder) PackageForFile(filename string) (string, bool, error) {
	dir := filepath.Dir(filename)
	pkgName := strings.ReplaceAll(dir, "/", ".")
	if strings.HasPrefix(pkgName, fb.Package.Name) {
		return fb.Package.Name, true, nil
	}
	return "", false, nil
}

func (fb *FileBuilder) IsLocalPackage(name string) bool {
	return strings.HasPrefix(name, fb.Package.Name)
}

func (fb *FileBuilder) PackageSourceFiles(ctx context.Context, pkgName string) ([]*protobuild.SourceFile, error) {
	if pkgName != fb.Package.Name {
		return nil, fmt.Errorf("package %s not found", pkgName)
	}

	ec := errset.NewCollector()
	summary, err := j5convert.SourceSummary(fb.SourceFile, ec)
	if err != nil {
		return nil, err
	}
	if ec.HasAny() {
		for _, err := range ec.Errors {
			return nil, fmt.Errorf("error in source file %s: %w", fb.Path, err)
		}
	}
	return []*protobuild.SourceFile{{
		J5File: fb.SourceFile,

		Summary: summary,
	}}, nil
}

func (fb *FileBuilder) PackageProseFiles(pkgName string) ([]*source_j5pb.ProseFile, error) {
	return []*source_j5pb.ProseFile{}, nil
}

func (fb *FileBuilder) ListObjectMethod(service, method string, object *schema_j5pb.Ref) *MethodBuilder {
	testService := fb.Service(service)
	listFoos := testService.Method(method)

	listFoos.Paged = true
	listFoos.Query = true
	listFoos.Request = &sourcedef_j5pb.AnonymousObject{}
	listFoos.Response = &sourcedef_j5pb.AnonymousObject{
		Properties: []*schema_j5pb.ObjectProperty{{
			Name: "items",
			Schema: &schema_j5pb.Field{
				Type: &schema_j5pb.Field_Array{
					Array: &schema_j5pb.ArrayField{
						Items: &schema_j5pb.Field{
							Type: &schema_j5pb.Field_Object{
								Object: &schema_j5pb.ObjectField{
									Schema: &schema_j5pb.ObjectField_Ref{
										Ref: object,
									},
								},
							},
						},
					},
				},
			},
		}},
	}

	return listFoos
}

type ObjectBuilder struct {
	*sourcedef_j5pb.Object
	file *FileBuilder
}

func (fb *FileBuilder) Object(name string) *ObjectBuilder {
	obj := &sourcedef_j5pb.Object{
		Def: &schema_j5pb.Object{
			Name: name,
		},
	}

	fb.Elements = append(fb.Elements, &sourcedef_j5pb.RootElement{
		Type: &sourcedef_j5pb.RootElement_Object{
			Object: obj,
		},
	})

	return &ObjectBuilder{
		Object: obj,
		file:   fb,
	}
}

type OneofBuilder struct {
	*sourcedef_j5pb.Oneof
	file *FileBuilder
}

func (fb *FileBuilder) Oneof(name string) *OneofBuilder {
	oneof := &sourcedef_j5pb.Oneof{
		Def: &schema_j5pb.Oneof{
			Name: name,
		},
	}
	fb.Elements = append(fb.Elements, &sourcedef_j5pb.RootElement{
		Type: &sourcedef_j5pb.RootElement_Oneof{
			Oneof: oneof,
		},
	})
	return &OneofBuilder{
		Oneof: oneof,
		file:  fb,
	}
}

type SerivceBuilder struct {
	*sourcedef_j5pb.Service
}

func (fb *FileBuilder) Service(name string) *SerivceBuilder {
	sb := &SerivceBuilder{
		Service: &sourcedef_j5pb.Service{
			Name: &name,
		},
	}

	fb.Elements = append(fb.Elements, &sourcedef_j5pb.RootElement{
		Type: &sourcedef_j5pb.RootElement_Service{
			Service: sb.Service,
		},
	})

	return sb
}

func (sb *SerivceBuilder) Method(name string) *MethodBuilder {
	method := &sourcedef_j5pb.APIMethod{
		HttpPath:   "path",
		HttpMethod: schema_j5pb.HTTPMethod_GET,
		Name:       name,
	}

	sb.Methods = append(sb.Methods, method)

	return &MethodBuilder{
		APIMethod: method,
	}
}

type MethodBuilder struct {
	*sourcedef_j5pb.APIMethod
}
