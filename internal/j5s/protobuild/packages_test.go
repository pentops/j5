package protobuild

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/bufbuild/protocompile/linker"
	"github.com/pentops/golib/gl"
	"github.com/pentops/j5/internal/j5s/protobuild/psrc"
	"github.com/pentops/j5/internal/j5s/protoprint"
	"github.com/pentops/log.go/log"
	"google.golang.org/protobuf/types/descriptorpb"
)

type fileSet map[string]linker.File

func (fs fileSet) expectFile(t *testing.T, filename string) linker.File {
	t.Helper()
	file, ok := fs[filename]
	if !ok {
		t.Fatalf("Expected file %s", filename)
	}
	return file
}

func testCompile(t *testing.T, tf *testFiles, td psrc.DescriptorFiles, pkg string) fileSet {
	if td == nil {
		td = psrc.DescriptorFiles{}
	}
	before := log.DefaultLogger

	log.DefaultLogger = log.NewTestLogger(t)
	defer func() {
		log.DefaultLogger = before
	}()

	tfResolver, err := NewSourceResolver(tf)
	if err != nil {
		t.Fatalf("FATAL: Unexpected error: %s", err.Error())
	}

	cc, err := NewPackageSet(td, tfResolver)
	if err != nil {
		t.Fatalf("FATAL: Unexpected error: %s", err.Error())
	}

	out, err := cc.CompilePackage(context.Background(), pkg)
	if err != nil {
		t.Fatalf("FATAL: Unexpected error: %s", err.Error())
	}

	files := make(fileSet)
	for _, file := range out.Proto {
		t.Logf("GOT FILE %s", file.Linked.Path())
		files[file.Linked.Path()] = file.Linked
	}

	return files
}

func TestImportJ5FromProto(t *testing.T) {
	tf := newTestFiles()

	tf.tAddJ5SFile("local/v1/foo.j5s",
		"object Foo {",
		"  field f1 string",
		"}",
	)

	tf.tAddProtoFile("local/v1/bar.proto",
		`import "local/v1/foo.j5s.proto";`,
		"message Bar {",
		"  Foo foo = 1;",
		"}",
	)

	{
		files := testCompile(t, tf, nil, "local.v1")
		files.expectFile(t, "local/v1/foo.j5s.proto")
		files.expectFile(t, "local/v1/bar.proto")
	}

	{ // Import the same file twice.
		tf.tAddProtoFile("local/v1/baz.proto",
			`import "local/v1/foo.j5s.proto";`,
			"message Baz {",
			"  Foo foo = 1;",
			"}",
		)

		files := testCompile(t, tf, nil, "local.v1")
		files.expectFile(t, "local/v1/foo.j5s.proto")
		files.expectFile(t, "local/v1/bar.proto")
		files.expectFile(t, "local/v1/baz.proto")
	}
}

func TestImportProtoToJ5Local(t *testing.T) {
	tf := newTestFiles()

	tf.tAddJ5SFile("local/v1/foo.j5s",
		"object Foo {",
		"  field bar object:Bar", // No import required for local package
		"}",
	)

	tf.tAddProtoFile("local/v1/bar.proto",
		"message Bar {",
		"  string f1 = 1;",
		"}",
	)

	files := testCompile(t, tf, nil, "local.v1")
	files.expectFile(t, "local/v1/foo.j5s.proto")
	files.expectFile(t, "local/v1/bar.proto")
}

func TestImportProtoToJ5Other(t *testing.T) {
	tf := newTestFiles()

	tf.tAddJ5SFile("foo/v1/foo.j5s",
		`import "bar/v1/bar.proto"`,
		"object Foo {",
		"  field bar object:bar.v1.Bar",
		"}",
	)

	tf.tAddProtoFile("bar/v1/bar.proto",
		"message Bar {",
		"  string f1 = 1;",
		"}",
	)

	files := testCompile(t, tf, nil, "foo.v1")
	files.expectFile(t, "foo/v1/foo.j5s.proto")
}

func TestImportDoubleExternal(t *testing.T) {
	// This test ensures that the imports from both Proto and J5s, and across
	// multiple files generally, reuse the same descriptor or otherwise
	// de-duplicate in a way which works.

	log.DefaultLogger = log.NewTestLogger(t)
	tf := newTestFiles()

	tf.tAddJ5SFile("local/v1/foo.j5s",
		`import "external/v1/ext.proto"`,
		"object Use {",
		"  field ext object:external.v1.Ext", // External
		"}")

	tf.tAddProtoFile("local/v1/bar.proto",
		`import "external/v1/ext.proto";`,
		"message Bar {",
		"  external.v1.Ext ext = 1;",
		"}",
	)

	td := psrc.DescriptorFiles{
		"external/v1/ext.proto": {
			Name:    gl.Ptr("external/v1/ext.proto"),
			Syntax:  gl.Ptr("proto3"),
			Package: gl.Ptr("external.v1"),
			MessageType: []*descriptorpb.DescriptorProto{{
				Name: gl.Ptr("Ext"),
			}},
		},
	}

	files := testCompile(t, tf, td, "local.v1")
	files.expectFile(t, "local/v1/foo.j5s.proto")
	files.expectFile(t, "local/v1/bar.proto")
}

func TestImportNestedJ5ToJ5Local(t *testing.T) {
	tf := newTestFiles()

	tf.tAddJ5SFile("local/v1/foo.j5s", `
		object Foo {
		  field bar object:"Bar.Baz"
		}

		object Bar {
		  object Baz {
		    field f1 string
		  }
		}
	`)

	files := testCompile(t, tf, nil, "local.v1")
	files.expectFile(t, "local/v1/foo.j5s.proto")
}

func TestImportNestedJ5ToJ5Other(t *testing.T) {
	tf := newTestFiles()

	tf.tAddJ5SFile("foo/v1/foo.j5s", `
		import bar.v1
		
		object Foo {
		  field barExplicit object:bar.v1."Bar.Baz"
		}
	`)

	tf.tAddJ5SFile("bar/v1/bar.j5s", `
		object Bar {
		  object Baz {
		    field f1 string
		  }
		}
	`)

	files := testCompile(t, tf, nil, "foo.v1")
	files.expectFile(t, "foo/v1/foo.j5s.proto")
}

func TestCircularPackageDependency(t *testing.T) {
	ctx := context.Background()

	tf := &testFiles{
		localFiles: map[string][]byte{
			"foo/v1/foo.proto": []byte(`
				syntax = "proto3";
				package foo.v1;
				import "bar/v1/bar.proto";
			`),
			"bar/v1/bar.proto": []byte(`
				syntax = "proto3";
				package bar.v1;
				import "baz/v1/baz.proto";
			`),
			"baz/v1/baz.proto": []byte(`
				syntax = "proto3";
				package baz.v1;
				import "foo/v1/foo.proto";
			`),
		},
		localPackages: []string{"foo.v1", "bar.v1", "baz.v1"},
	}

	tfResolver, err := NewSourceResolver(tf)
	if err != nil {
		t.Fatalf("FATAL: Unexpected error: %s", err.Error())
	}
	cc, err := NewPackageSet(psrc.DescriptorFiles{}, tfResolver)
	if err != nil {
		t.Fatalf("FATAL: Unexpected error: %s", err.Error())
	}

	_, err = cc.CompilePackage(ctx, "foo.v1")
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
	cde := &CircularDependencyError{}

	if !errors.As(err, &cde) {
		t.Fatalf("Expected CircularDependencyError, got %T (%s)", err, err.Error())
	}
}

func TestPreserveComments(t *testing.T) {
	tf := newTestFiles()

	tf.tAddJ5SFile("local/v1/foo.j5s",
		"object Foo {",
		"  | Description of Foo",
		"  field field string",
		"}",
	)

	files := testCompile(t, tf, nil, "local.v1")
	ff := files.expectFile(t, "local/v1/foo.j5s.proto")

	foo := ff.Messages().ByName("Foo")
	sourceLoc := ff.SourceLocations().ByDescriptor(foo)

	if sourceLoc.LeadingComments != " Description of Foo\n" {
		t.Fatalf("Expected leading comments, got %#v", sourceLoc)
	}

	out, err := protoprint.PrintFile(context.Background(), ff, "generate comment")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(out)

	lines := strings.Split(out, "\n")
	want := []string{
		`// generate comment`,
		``,
		`syntax = "proto3";`,
		``,
		`package local.v1;`,
		``,
		`import "j5/ext/v1/annotations.proto";`,
		``,
		`// Description of Foo`,
		`message Foo {`,
		`  option (j5.ext.v1.message).object = {};`,
		``,
		`  string field = 1 [(j5.ext.v1.field).string = {}];`,
		`}`,
		``,
	}

	assertEqualLines(t, want, lines)
}

func assertEqualLines(t testing.TB, wantLines, gotLines []string) {
	for idx, line := range gotLines {
		t.Logf("got %03d: '%s'", idx, line)
		if idx >= len(wantLines) {
			t.Fatalf("Extra line %d: %s", idx, line)
		}
		if line != wantLines[idx] {
			t.Fatalf("Line %d: want %q, got %q", idx, wantLines[idx], line)
		}
	}
}
