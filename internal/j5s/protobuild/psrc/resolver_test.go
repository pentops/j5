package psrc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestFileLoad(t *testing.T) {
	externalDeps := DescriptorFiles{
		"external/v1/foo.proto": {
			Name:    proto.String("external/v1/foo.proto"),
			Syntax:  proto.String("proto3"),
			Package: proto.String("external.v1"),
		},
	}

	rr, err := ChainResolver(externalDeps)
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
