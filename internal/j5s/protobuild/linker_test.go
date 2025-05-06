package protobuild

import (
	"testing"

	"github.com/bufbuild/protocompile/linker"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestLinkerDeps(t *testing.T) {
	ctx := t.Context()

	tf := &testDeps{
		externalDeps: map[string]*descriptorpb.FileDescriptorProto{
			"test/v1/foo.proto": {
				Name:    proto.String("test/v1/foo.proto"),
				Syntax:  proto.String("proto3"),
				Package: proto.String("test.v1"),
				Dependency: []string{
					"google/protobuf/timestamp.proto",
				},
			},
		},
	}

	deps, err := newDependencyResolver(tf)
	if err != nil {
		t.Fatalf("FATAL: Unexpected error: %s", err.Error())
	}

	resolver := newResolverCache(newBuiltinResolver(), deps)

	filenames := []string{"test/v1/foo.proto"}

	cc := newLinker(resolver, &linker.Symbols{})
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
