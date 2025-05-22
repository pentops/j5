package j5convert

import (
	"errors"
	"testing"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/pentops/flowtest/prototest"
	"github.com/pentops/golib/gl"
	"github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/gen/j5/sourcedef/v1/sourcedef_j5pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestPackageParse(t *testing.T) {

	for _, tc := range []struct {
		input   string
		wantPkg string
		wantSub string
		wantErr bool
	}{{
		input:   "test/v1/foo.j5s",
		wantPkg: "test.v1",
		wantSub: "",
	}, {
		input:   "test/v1/sub/foo.j5s",
		wantPkg: "test.v1",
		wantSub: "sub",
	}, {
		input:   "test/v1/sub/subsub/foo.j5s",
		wantErr: true,
	}, {
		input:   "test",
		wantErr: true,
	}, {
		input:   "foo/bar/v1/foo.j5s",
		wantPkg: "foo.bar.v1",
		wantSub: "",
	}, {
		input:   "foo/bar/v1/sub/foo.j5s",
		wantPkg: "foo.bar.v1",
		wantSub: "sub",
	}} {
		t.Run(tc.input, func(t *testing.T) {
			gotPkg, gotSub, err := SplitPackageFromFilename(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("parsePackage(%q) = %q, %q, nil, want error", tc.input, gotPkg, gotSub)
				}
				return
			}
			if err != nil {
				t.Fatalf("returned error: %s", err)
			}
			if gotPkg != tc.wantPkg || gotSub != tc.wantSub {
				t.Fatalf("%q -> %q, %q want %q, %q", tc.input, gotPkg, gotSub, tc.wantPkg, tc.wantSub)

			}
		})
	}

}
func withOption[T protoreflect.ProtoMessage](opt T, extType protoreflect.ExtensionType, extVal proto.Message) T {
	proto.SetExtension(opt, extType, extVal)
	return opt
}

type testDeps struct {
	pkg   string
	types map[string]*TypeRef
}

func (d *testDeps) PackageName() string {
	return d.pkg
}

func (d *testDeps) ResolveType(pkg string, name string) (*TypeRef, error) {
	if pkg == "" {
		pkg = d.pkg
	}

	if tr, ok := d.types[pkg+"."+name]; ok {
		return tr, nil
	}

	return nil, &TypeNotFoundError{
		Package: pkg,
		Name:    name,
	}
}
func assertIsTypeNotFound(t *testing.T, err error, want *TypeNotFoundError) {
	gotNotFound := &TypeNotFoundError{}
	if !errors.As(err, &gotNotFound) {
		t.Fatalf("got error %v, want TypeNotFoundError", err)
	}
	if gotNotFound.Package != want.Package || gotNotFound.Name != want.Name {
		t.Fatalf("got error %v, want %v", gotNotFound, want)
	}
}

func assertIsPackageNotFound(t *testing.T, err error, want *PackageNotFoundError) {
	gotErr := &PackageNotFoundError{}
	if !errors.As(err, &gotErr) {
		t.Fatalf("got error %v, want TypeNotFoundError", err)
	}
	if gotErr.Package != want.Package {
		t.Fatalf("got error %v, want %v", gotErr, want)
	}
}

var emptyObjectOption = withOption(&descriptorpb.MessageOptions{}, ext_j5pb.E_Message, &ext_j5pb.MessageOptions{
	Type: &ext_j5pb.MessageOptions_Object{
		Object: &ext_j5pb.ObjectMessageOptions{},
	},
})

func TestSchemaToProto(t *testing.T) {

	deps := &testDeps{
		pkg: "test.v1",
		types: map[string]*TypeRef{
			"test.v1.TestEnum": {
				Package: "test.v1",
				Name:    "TestEnum",
				File:    "test/v1/test.j5s.proto",
				Enum: &EnumRef{
					Prefix: "TEST_ENUM_",
					ValMap: map[string]int32{
						"TEST_ENUM_FOO": 1,
					},
				},
			},
		},
	}

	objectSchema := &sourcedef_j5pb.RootElement{
		Type: &sourcedef_j5pb.RootElement_Object{
			Object: &sourcedef_j5pb.Object{
				Def: &schema_j5pb.Object{
					Name:        "Referenced",
					Description: "Message Comment",
					Properties: []*schema_j5pb.ObjectProperty{
						{
							Name:        "field1",
							Description: "Field Comment",
							Schema: &schema_j5pb.Field{
								Type: &schema_j5pb.Field_String_{
									String_: &schema_j5pb.StringField{},
								},
							},
						},
						{
							Name: "enum",
							Schema: &schema_j5pb.Field{
								Type: &schema_j5pb.Field_Enum{
									Enum: &schema_j5pb.EnumField{
										Schema: &schema_j5pb.EnumField_Ref{
											Ref: &schema_j5pb.Ref{
												Package: "",
												Schema:  "TestEnum",
											},
										},
									},
								},
							},
						},
						{
							Name:        "array",
							Description: "Field Comment",
							Schema: &schema_j5pb.Field{
								Type: &schema_j5pb.Field_Array{
									Array: &schema_j5pb.ArrayField{
										Items: &schema_j5pb.Field{
											Type: &schema_j5pb.Field_String_{
												String_: &schema_j5pb.StringField{
													Rules: &schema_j5pb.StringField_Rules{
														MinLength: proto.Uint64(1),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	enumSchema := &sourcedef_j5pb.RootElement{
		Type: &sourcedef_j5pb.RootElement_Enum{
			Enum: &schema_j5pb.Enum{
				Name:   "TestEnum",
				Prefix: "TEST_ENUM_",
				Options: []*schema_j5pb.Enum_Option{
					{
						Name:   "UNSPECIFIED",
						Number: 0,
					},
					{
						Name:   "FOO",
						Number: 1,
					},
				},
			},
		},
	}
	sourceFile := &sourcedef_j5pb.SourceFile{
		Package:  &sourcedef_j5pb.Package{Name: "test.v1"},
		Path:     "test/v1/test.j5s",
		Elements: []*sourcedef_j5pb.RootElement{objectSchema, enumSchema},
	}

	gotFile, err := ConvertJ5File(deps, sourceFile)
	if err != nil {
		t.Fatalf("ConvertJ5File failed: %v", err)
	}

	wantFile := &descriptorpb.FileDescriptorProto{
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{
			//GoPackage: proto.String("github.com/pentops/j5/test/v1/test_pb"),
		},
		Dependency: []string{
			"buf/validate/validate.proto",
			"j5/ext/v1/annotations.proto",
		},
		Name:    proto.String("test/v1/test.j5s.proto"),
		Package: proto.String("test.v1"),
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String("Referenced"),
			Field: []*descriptorpb.FieldDescriptorProto{
				{
					Name:     proto.String("field_1"),
					Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					Number:   proto.Int32(1),
					Options:  tEmptyTypeExt(t, "string"),
					JsonName: proto.String("field1"),
				},
				{
					Name:     proto.String("enum"),
					Type:     descriptorpb.FieldDescriptorProto_TYPE_ENUM.Enum(),
					Number:   proto.Int32(2),
					TypeName: proto.String(".test.v1.TestEnum"),
					Options: withOption(tEmptyTypeExt(t, "enum"), validate.E_Field, &validate.FieldConstraints{
						Type: &validate.FieldConstraints_Enum{
							Enum: &validate.EnumRules{
								DefinedOnly: gl.Ptr(true),
							},
						},
					}),
					JsonName: proto.String("enum"),
				},
				{
					Name:   proto.String("array"),
					Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					Label:  descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
					Number: proto.Int32(3),
					Options: withOption(tEmptyTypeExt(t, "array"), validate.E_Field, &validate.FieldConstraints{
						Type: &validate.FieldConstraints_Repeated{
							Repeated: &validate.RepeatedRules{
								Items: &validate.FieldConstraints{
									Type: &validate.FieldConstraints_String_{
										String_: &validate.StringRules{
											MinLen: proto.Uint64(1),
										},
									},
								},
							},
						},
					}),
					JsonName: proto.String("array"),
				},
			},
			Options: emptyObjectOption,
		}},
		EnumType: []*descriptorpb.EnumDescriptorProto{{
			Name: proto.String("TestEnum"),
			Value: []*descriptorpb.EnumValueDescriptorProto{{
				Name:   proto.String("TEST_ENUM_UNSPECIFIED"),
				Number: proto.Int32(0),
			}, {
				Name:   proto.String("TEST_ENUM_FOO"),
				Number: proto.Int32(1),
			}},
		}},
	}

	gotFile[0].SourceCodeInfo = nil
	prototest.AssertEqualProto(t, wantFile, gotFile[0])
}

func TestEnumConvert(t *testing.T) {
	schema := &sourcedef_j5pb.RootElement{
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
	}

	wantFile := &descriptorpb.FileDescriptorProto{
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{},
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

	gotFile := testConvert(t, schema)
	prototest.AssertEqualProto(t, wantFile, gotFile)
}

func TestPolymorphConvert(t *testing.T) {

	polymorphSchema := &sourcedef_j5pb.RootElement{
		Type: &sourcedef_j5pb.RootElement_Polymorph{
			Polymorph: &sourcedef_j5pb.Polymorph{
				Def: &schema_j5pb.Polymorph{
					Name:  "TestPolymorph",
					Types: []string{"foo.v1.Foo", "bar.v1.Bar"},
				},
				Includes: []string{"baz.v1.BazPoly"},
			},
		},
	}

	wantFile := &descriptorpb.FileDescriptorProto{
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{},
		Name:    proto.String("test/v1/test.j5s.proto"),
		Package: proto.String("test.v1"),
		Dependency: []string{
			"j5/ext/v1/annotations.proto",
		},
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String("TestPolymorph"),
			Options: withOption(&descriptorpb.MessageOptions{}, ext_j5pb.E_Message, &ext_j5pb.MessageOptions{
				Type: &ext_j5pb.MessageOptions_Polymorph{
					Polymorph: &ext_j5pb.PolymorphMessageOptions{
						Types: []string{
							"foo.v1.Foo",
							"bar.v1.Bar",
						},
					},
				},
			}),
		}},
	}

	gotFile := testConvert(t, polymorphSchema)
	prototest.AssertEqualProto(t, wantFile, gotFile)

}

func testConvert(t testing.TB, schemas ...*sourcedef_j5pb.RootElement) *descriptorpb.FileDescriptorProto {
	t.Helper()
	deps := &testDeps{
		pkg: "test.v1",
	}
	sourceFile := &sourcedef_j5pb.SourceFile{
		Path:     "test/v1/test.j5s",
		Package:  &sourcedef_j5pb.Package{Name: "test.v1"},
		Elements: schemas,
	}
	gotFiles, err := ConvertJ5File(deps, sourceFile)
	if err != nil {
		t.Fatalf("ConvertJ5File failed: %v", err)
	}
	if len(gotFiles) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(gotFiles))
	}
	gotFile := gotFiles[0]
	gotFile.SourceCodeInfo = nil
	return gotFile
}

// Copies the J5 extension object to the equivalent protoreflect extension type
// by field names.
func tEmptyTypeExt(t testing.TB, fieldType protoreflect.Name) *descriptorpb.FieldOptions {

	// Options in the *proto* representation.
	extOptions := &ext_j5pb.FieldOptions{}
	extOptionsRefl := extOptions.ProtoReflect()

	// The proto extension is a oneof to each field type, which should match the
	// specified type.

	typeField := extOptionsRefl.Descriptor().Fields().ByName(fieldType)
	if typeField == nil {
		t.Fatalf("Field %s does not have a type field", fieldType)
	}

	extTypedRefl := extOptionsRefl.Mutable(typeField).Message()
	if extTypedRefl == nil {
		t.Fatalf("Field %s type field is not a message", fieldType)
	}

	fieldOptions := &descriptorpb.FieldOptions{}

	proto.SetExtension(fieldOptions, ext_j5pb.E_Field, extOptions)
	return fieldOptions
}
