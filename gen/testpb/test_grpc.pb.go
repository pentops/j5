// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: test/v1/test.proto

package testpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// FooServiceClient is the client API for FooService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FooServiceClient interface {
	GetFoo(ctx context.Context, in *GetFooRequest, opts ...grpc.CallOption) (*GetFooResponse, error)
	PostFoo(ctx context.Context, in *PostFooRequest, opts ...grpc.CallOption) (*PostFooResponse, error)
}

type fooServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewFooServiceClient(cc grpc.ClientConnInterface) FooServiceClient {
	return &fooServiceClient{cc}
}

func (c *fooServiceClient) GetFoo(ctx context.Context, in *GetFooRequest, opts ...grpc.CallOption) (*GetFooResponse, error) {
	out := new(GetFooResponse)
	err := c.cc.Invoke(ctx, "/test.v1.FooService/GetFoo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fooServiceClient) PostFoo(ctx context.Context, in *PostFooRequest, opts ...grpc.CallOption) (*PostFooResponse, error) {
	out := new(PostFooResponse)
	err := c.cc.Invoke(ctx, "/test.v1.FooService/PostFoo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FooServiceServer is the server API for FooService service.
// All implementations must embed UnimplementedFooServiceServer
// for forward compatibility
type FooServiceServer interface {
	GetFoo(context.Context, *GetFooRequest) (*GetFooResponse, error)
	PostFoo(context.Context, *PostFooRequest) (*PostFooResponse, error)
	mustEmbedUnimplementedFooServiceServer()
}

// UnimplementedFooServiceServer must be embedded to have forward compatible implementations.
type UnimplementedFooServiceServer struct {
}

func (UnimplementedFooServiceServer) GetFoo(context.Context, *GetFooRequest) (*GetFooResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFoo not implemented")
}
func (UnimplementedFooServiceServer) PostFoo(context.Context, *PostFooRequest) (*PostFooResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostFoo not implemented")
}
func (UnimplementedFooServiceServer) mustEmbedUnimplementedFooServiceServer() {}

// UnsafeFooServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FooServiceServer will
// result in compilation errors.
type UnsafeFooServiceServer interface {
	mustEmbedUnimplementedFooServiceServer()
}

func RegisterFooServiceServer(s grpc.ServiceRegistrar, srv FooServiceServer) {
	s.RegisterService(&FooService_ServiceDesc, srv)
}

func _FooService_GetFoo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetFooRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FooServiceServer).GetFoo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/test.v1.FooService/GetFoo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FooServiceServer).GetFoo(ctx, req.(*GetFooRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FooService_PostFoo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PostFooRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FooServiceServer).PostFoo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/test.v1.FooService/PostFoo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FooServiceServer).PostFoo(ctx, req.(*PostFooRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// FooService_ServiceDesc is the grpc.ServiceDesc for FooService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var FooService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "test.v1.FooService",
	HandlerType: (*FooServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetFoo",
			Handler:    _FooService_GetFoo_Handler,
		},
		{
			MethodName: "PostFoo",
			Handler:    _FooService_PostFoo_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "test/v1/test.proto",
}

// FooTopicClient is the client API for FooTopic service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FooTopicClient interface {
	Foo(ctx context.Context, in *FooMessage, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type fooTopicClient struct {
	cc grpc.ClientConnInterface
}

func NewFooTopicClient(cc grpc.ClientConnInterface) FooTopicClient {
	return &fooTopicClient{cc}
}

func (c *fooTopicClient) Foo(ctx context.Context, in *FooMessage, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/test.v1.FooTopic/Foo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FooTopicServer is the server API for FooTopic service.
// All implementations must embed UnimplementedFooTopicServer
// for forward compatibility
type FooTopicServer interface {
	Foo(context.Context, *FooMessage) (*emptypb.Empty, error)
	mustEmbedUnimplementedFooTopicServer()
}

// UnimplementedFooTopicServer must be embedded to have forward compatible implementations.
type UnimplementedFooTopicServer struct {
}

func (UnimplementedFooTopicServer) Foo(context.Context, *FooMessage) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Foo not implemented")
}
func (UnimplementedFooTopicServer) mustEmbedUnimplementedFooTopicServer() {}

// UnsafeFooTopicServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FooTopicServer will
// result in compilation errors.
type UnsafeFooTopicServer interface {
	mustEmbedUnimplementedFooTopicServer()
}

func RegisterFooTopicServer(s grpc.ServiceRegistrar, srv FooTopicServer) {
	s.RegisterService(&FooTopic_ServiceDesc, srv)
}

func _FooTopic_Foo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FooMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FooTopicServer).Foo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/test.v1.FooTopic/Foo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FooTopicServer).Foo(ctx, req.(*FooMessage))
	}
	return interceptor(ctx, in, info, handler)
}

// FooTopic_ServiceDesc is the grpc.ServiceDesc for FooTopic service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var FooTopic_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "test.v1.FooTopic",
	HandlerType: (*FooTopicServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Foo",
			Handler:    _FooTopic_Foo_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "test/v1/test.proto",
}
