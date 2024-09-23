package grpcreflect

import (
	"context"
	"testing"

	"github.com/pentops/flowtest"
	"github.com/pentops/j5/gen/test/foo/v1/foo_testspb"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Service struct {
	foo_testspb.UnimplementedFooCommandServiceServer
	foo_testspb.UnimplementedFooQueryServiceServer
	foo_testspb.UnimplementedFooDownloadServiceServer
}

func TestProtoReadHappy(t *testing.T) {
	grpcPair := flowtest.NewGRPCPair(t)

	service := &Service{}
	foo_testspb.RegisterFooQueryServiceServer(grpcPair.Server, service)
	foo_testspb.RegisterFooCommandServiceServer(grpcPair.Server, service)
	foo_testspb.RegisterFooDownloadServiceServer(grpcPair.Server, service)

	reflection.Register(grpcPair.Server)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	grpcPair.ServeUntilDone(t, ctx)

	cl := NewClient(grpcPair.Client)

	desc, err := cl.FetchServices(ctx)
	if err != nil {
		t.Fatal(err)
	}

	byName := make(map[protoreflect.FullName]protoreflect.ServiceDescriptor)
	for _, d := range desc {
		byName[d.FullName()] = d
	}

	if len(byName) != 3 {
		t.Fatalf("expected 3 services, got %d", len(byName))
	}

	_, ok := byName["test.foo.v1.service.FooQueryService"]
	if !ok {
		t.Fatal("missing FooQueryService")
	}

}
