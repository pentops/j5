package protobuild

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
)

type testFiles struct {
	localFiles    map[string][]byte
	localPackages []string
	proseFiles    []*source_j5pb.ProseFile
}

func newTestFiles() *testFiles {
	return &testFiles{
		localFiles:    map[string][]byte{},
		localPackages: []string{},
	}
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

func (tf *testFiles) tAddProtoFile(filename string, body ...string) {
	pkg := tFileToPackage(filename)
	body = append([]string{
		`syntax = "proto3";`,
		fmt.Sprintf("package %s;", pkg),
	}, body...)
	tf.localFiles[filename] = []byte(strings.Join(body, "\n"))

	tf.tIncludePackage(pkg)
}

func (tf *testFiles) tAddJ5SFile(filename string, body ...string) {
	pkg := tFileToPackage(filename)
	body = append([]string{
		fmt.Sprintf("package %s", pkg),
	}, body...)
	tf.localFiles[filename] = []byte(strings.Join(body, "\n"))
	tf.tIncludePackage(pkg)
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
	return tf.proseFiles, nil
}

func TestLocalResolver(t *testing.T) {
	tf := newTestFiles()
	tf.tAddProtoFile("local/v1/foo.proto", "local.v1")
	tf.tAddJ5SFile("local/v1/bar.j5s", "local.v1")

	sourceResolver, err := NewSourceResolver(tf)
	if err != nil {
		t.Fatalf("FATAL: Unexpected error: %s", err.Error())
	}
	pkgName, isLocal, err := sourceResolver.PackageForFile("local/v1/foo.proto")
	if err != nil {
		t.Fatalf("FATAL: Unexpected error: %s", err.Error())
	}
	if !isLocal {
		t.Fatalf("Expected local package, got external")
	}
	if pkgName != "local.v1" {
		t.Fatalf("Expected package name to be local.v1, got %s", pkgName)
	}

	pkgIsLocal := sourceResolver.IsLocalPackage("local.v1")
	if !pkgIsLocal {
		t.Fatalf("Expected local package, got external")
	}
}
