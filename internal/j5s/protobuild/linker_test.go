package protobuild

import (
	"testing"

	"github.com/bufbuild/protocompile/linker"
	"github.com/pentops/j5/internal/j5s/protobuild/psrc"
	"google.golang.org/protobuf/proto"
)

func TestLinkerDeps(t *testing.T) {
	ctx := t.Context()

	tf := psrc.DescriptorFiles{
		"test/v1/foo.proto": {
			Name:    proto.String("test/v1/foo.proto"),
			Syntax:  proto.String("proto3"),
			Package: proto.String("test.v1"),
			Dependency: []string{
				"google/protobuf/timestamp.proto",
			},
		},
	}

	deps, err := psrc.ChainResolver(tf)
	if err != nil {
		t.Fatalf("FATAL: Unexpected error: %s", err.Error())
	}

	filenames := []string{"test/v1/foo.proto"}

	cc := newLinker(deps, &linker.Symbols{})
	files, err := cc.resolveAll(ctx, filenames)
	if err != nil {
		t.Fatalf("FATAL: Unexpected error: %s", err.Error())
	}
	if len(cc.errs.Errors) > 0 {
		t.Fatalf("FATAL: Unexpected errors: %v", cc.errs.Errors)
	}
	if len(files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(files))
	}

}
