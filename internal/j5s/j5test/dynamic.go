package j5test

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/internal/j5s/protobuild"
	"github.com/pentops/j5/internal/j5s/protobuild/psrc"
	"github.com/pentops/j5/lib/j5reflect"
	"github.com/pentops/j5/lib/j5schema"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

type testFiles struct {
	localFiles    map[string][]byte
	localPackages []string
}

func newTestFiles() *testFiles {
	return &testFiles{
		localFiles:    map[string][]byte{},
		localPackages: []string{},
	}
}

func (tf *testFiles) ListPackages() []string {
	return tf.localPackages
}

func (tf *testFiles) ListSourceFiles(ctx context.Context, prefix string) ([]string, error) {
	var files []string
	for k := range tf.localFiles {
		if strings.HasPrefix(k, prefix) {
			files = append(files, k)
		}
	}
	sort.Strings(files) // makes testing easier
	return files, nil
}

func (tf *testFiles) GetLocalFile(ctx context.Context, filename string) ([]byte, error) {
	if desc, ok := tf.localFiles[filename]; ok {
		return desc, nil
	}
	return nil, fmt.Errorf("file not found: %s", filename)
}

func (tf *testFiles) ProseFiles(pkgName string) ([]*source_j5pb.ProseFile, error) {
	return []*source_j5pb.ProseFile{}, nil
}

func (tf *testFiles) tIncludePackage(pkg string) {
	if slices.Contains(tf.localPackages, pkg) {
		return
	}
	tf.localPackages = append(tf.localPackages, pkg)
}

func tFileToPackage(filename string) string {
	parts := strings.Split(filename, "/")
	if len(parts) < 2 {
		return "default"
	}
	pkg := strings.Join(parts[:len(parts)-1], ".")
	return pkg
}

func (tf *testFiles) tAddJ5SFile(filename string, body ...string) {
	pkg := tFileToPackage(filename)
	body = append([]string{
		fmt.Sprintf("package %s", pkg),
	}, body...)
	tf.localFiles[filename] = []byte(strings.Join(body, "\n"))
	tf.tIncludePackage(pkg)
}

func ObjectReflect(t testing.TB, file string) protoreflect.MessageDescriptor {

	pkgName := "pkg" + strings.ToLower(strings.ReplaceAll(uuid.New().String(), "-", ""))

	locals := newTestFiles()
	locals.tAddJ5SFile(pkgName+"/v1/file.j5s", file)

	deps := psrc.NewBuiltinResolver()
	ps, err := protobuild.NewPackageSet(deps, locals)
	if err != nil {
		t.Fatalf("FATAL: Unexpected error: %s", err.Error())
	}

	pkg, err := ps.CompilePackage(t.Context(), pkgName+".v1")
	if err != nil {
		t.Fatalf("FATAL: Unexpected error: %s", err.Error())
	}

	if len(pkg.Proto) != 1 {
		t.Fatalf("FATAL: Expected exactly one proto file, got %d", len(pkg.Proto))
	}
	fd, err := protodesc.NewFile(pkg.Proto[0].Desc, protoregistry.GlobalFiles)
	if err != nil {
		t.Fatal(fmt.Errorf("FATAL: Failed to create file descriptor: %w", err))
	}

	messages := fd.Messages()
	if messages.Len() == 0 {
		t.Fatal("FATAL: No messages found in file descriptor")
	} else if messages.Len() > 1 {
		t.Fatalf("FATAL: Expected exactly one message, got %d", messages.Len())
	}
	desc := messages.Get(0)

	return desc
}

func DynamicObject(t testing.TB, file string) j5reflect.Object {
	desc := ObjectReflect(t, file)

	schemaCache := j5schema.NewSchemaCache()
	msg := dynamicpb.NewMessage(desc)

	reflector := j5reflect.NewWithCache(schemaCache)
	root, err := reflector.NewRoot(msg)
	if err != nil {
		t.Fatalf("FATAL: Failed to create root: %s", err.Error())
	}

	rootObj, ok := root.(j5reflect.Object)
	if !ok {
		t.Fatalf("FATAL: Root is not an object")
	}

	return rootObj

}
