package psrc

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

type testDeps struct {
	externalDeps map[string]*descriptorpb.FileDescriptorProto
}

func (tf *testDeps) GetDependencyFile(filename string) (*descriptorpb.FileDescriptorProto, error) {
	if desc, ok := tf.externalDeps[filename]; ok {
		return desc, nil
	}
	return nil, fmt.Errorf("file not found: %s", filename)
}

func (tf *testDeps) ListDependencyFiles(root string) []string {
	var files []string
	for k := range tf.externalDeps {
		if strings.HasPrefix(k, root) {
			files = append(files, k)
		}
	}
	sort.Strings(files) // makes testing easier
	return files
}

func TestFileLoad(t *testing.T) {
	tf := &testDeps{
		externalDeps: map[string]*descriptorpb.FileDescriptorProto{
			"external/v1/foo.proto": {
				Name:    proto.String("external/v1/foo.proto"),
				Syntax:  proto.String("proto3"),
				Package: proto.String("external.v1"),
			},
		},
	}

	rr, err := ChainResolver(tf.externalDeps)
	if err != nil {
		t.Fatalf("FATAL: Unexpected error: %s", err.Error())
	}

	t.Run("Builtin", func(t *testing.T) {
		path := "j5/list/v1/query.proto"
		result, err := rr.FindFileByPath(path)
		if err != nil {
			t.Fatalf("FATAL: Unexpected error: %s", err.Error())
		}
		if result.Refl == nil {
			t.Fatal("FATAL: result.Refl is nil")
		}
		assert.Equal(t, path, result.Refl.Path())
	})

	t.Run("External", func(t *testing.T) {
		result, err := rr.FindFileByPath("external/v1/foo.proto")
		if err != nil {
			t.Fatalf("FATAL: Unexpected error: %s", err.Error())
		}
		if result.Desc == nil {
			t.Fatal("FATAL: result.Desc is nil")
		}
		assert.Equal(t, "external/v1/foo.proto", *result.Desc.Name)
	})
}
