package protobuild

import (
	"errors"
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

func TestLinkerCircularDeps(t *testing.T) {
	ctx := t.Context()

	tf := psrc.DescriptorFiles{
		"test/v1/foo.proto": {
			Name:    proto.String("test/v1/foo.proto"),
			Syntax:  proto.String("proto3"),
			Package: proto.String("test.v1"),
			Dependency: []string{
				"test/v1/bar.proto",
			},
		},
		"test/v1/bar.proto": {
			Name:    proto.String("test/v1/bar.proto"),
			Syntax:  proto.String("proto3"),
			Package: proto.String("test.v1"),
			Dependency: []string{
				"test/v1/foo.proto",
			},
		},
	}

	deps, err := psrc.ChainResolver(tf)
	if err != nil {
		t.Fatalf("FATAL: Unexpected error: %s", err.Error())
	}

	filenames := []string{"test/v1/foo.proto"}

	cc := newLinker(deps, &linker.Symbols{})
	_, err = cc.resolveAll(ctx, filenames)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
	cde := &CircularDependencyError{}

	if !errors.As(err, &cde) {
		t.Fatalf("Expected CircularDependencyError, got %T (%s)", err, err.Error())
	}
}
