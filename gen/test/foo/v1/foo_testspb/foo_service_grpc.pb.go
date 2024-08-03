// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.4.0
// - protoc             (unknown)
// source: test/foo/v1/service/foo_service.proto

package foo_testspb

import (
	context "context"
	httpbody "google.golang.org/genproto/googleapis/api/httpbody"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.62.0 or later.
const _ = grpc.SupportPackageIsVersion8

const (
	FooQueryService_GetFoo_FullMethodName        = "/test.foo.v1.service.FooQueryService/GetFoo"
	FooQueryService_ListFoos_FullMethodName      = "/test.foo.v1.service.FooQueryService/ListFoos"
	FooQueryService_ListFooEvents_FullMethodName = "/test.foo.v1.service.FooQueryService/ListFooEvents"
)

// FooQueryServiceClient is the client API for FooQueryService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FooQueryServiceClient interface {
	GetFoo(ctx context.Context, in *GetFooRequest, opts ...grpc.CallOption) (*GetFooResponse, error)
	ListFoos(ctx context.Context, in *ListFoosRequest, opts ...grpc.CallOption) (*ListFoosResponse, error)
	ListFooEvents(ctx context.Context, in *ListFooEventsRequest, opts ...grpc.CallOption) (*ListFooEventsResponse, error)
}

type fooQueryServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewFooQueryServiceClient(cc grpc.ClientConnInterface) FooQueryServiceClient {
	return &fooQueryServiceClient{cc}
}

func (c *fooQueryServiceClient) GetFoo(ctx context.Context, in *GetFooRequest, opts ...grpc.CallOption) (*GetFooResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetFooResponse)
	err := c.cc.Invoke(ctx, FooQueryService_GetFoo_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fooQueryServiceClient) ListFoos(ctx context.Context, in *ListFoosRequest, opts ...grpc.CallOption) (*ListFoosResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListFoosResponse)
	err := c.cc.Invoke(ctx, FooQueryService_ListFoos_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fooQueryServiceClient) ListFooEvents(ctx context.Context, in *ListFooEventsRequest, opts ...grpc.CallOption) (*ListFooEventsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListFooEventsResponse)
	err := c.cc.Invoke(ctx, FooQueryService_ListFooEvents_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FooQueryServiceServer is the server API for FooQueryService service.
// All implementations must embed UnimplementedFooQueryServiceServer
// for forward compatibility
type FooQueryServiceServer interface {
	GetFoo(context.Context, *GetFooRequest) (*GetFooResponse, error)
	ListFoos(context.Context, *ListFoosRequest) (*ListFoosResponse, error)
	ListFooEvents(context.Context, *ListFooEventsRequest) (*ListFooEventsResponse, error)
	mustEmbedUnimplementedFooQueryServiceServer()
}

// UnimplementedFooQueryServiceServer must be embedded to have forward compatible implementations.
type UnimplementedFooQueryServiceServer struct {
}

func (UnimplementedFooQueryServiceServer) GetFoo(context.Context, *GetFooRequest) (*GetFooResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFoo not implemented")
}
func (UnimplementedFooQueryServiceServer) ListFoos(context.Context, *ListFoosRequest) (*ListFoosResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListFoos not implemented")
}
func (UnimplementedFooQueryServiceServer) ListFooEvents(context.Context, *ListFooEventsRequest) (*ListFooEventsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListFooEvents not implemented")
}
func (UnimplementedFooQueryServiceServer) mustEmbedUnimplementedFooQueryServiceServer() {}

// UnsafeFooQueryServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FooQueryServiceServer will
// result in compilation errors.
type UnsafeFooQueryServiceServer interface {
	mustEmbedUnimplementedFooQueryServiceServer()
}

func RegisterFooQueryServiceServer(s grpc.ServiceRegistrar, srv FooQueryServiceServer) {
	s.RegisterService(&FooQueryService_ServiceDesc, srv)
}

func _FooQueryService_GetFoo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetFooRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FooQueryServiceServer).GetFoo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FooQueryService_GetFoo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FooQueryServiceServer).GetFoo(ctx, req.(*GetFooRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FooQueryService_ListFoos_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListFoosRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FooQueryServiceServer).ListFoos(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FooQueryService_ListFoos_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FooQueryServiceServer).ListFoos(ctx, req.(*ListFoosRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FooQueryService_ListFooEvents_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListFooEventsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FooQueryServiceServer).ListFooEvents(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FooQueryService_ListFooEvents_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FooQueryServiceServer).ListFooEvents(ctx, req.(*ListFooEventsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// FooQueryService_ServiceDesc is the grpc.ServiceDesc for FooQueryService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var FooQueryService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "test.foo.v1.service.FooQueryService",
	HandlerType: (*FooQueryServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetFoo",
			Handler:    _FooQueryService_GetFoo_Handler,
		},
		{
			MethodName: "ListFoos",
			Handler:    _FooQueryService_ListFoos_Handler,
		},
		{
			MethodName: "ListFooEvents",
			Handler:    _FooQueryService_ListFooEvents_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "test/foo/v1/service/foo_service.proto",
}

const (
	FooCommandService_PostFoo_FullMethodName = "/test.foo.v1.service.FooCommandService/PostFoo"
)

// FooCommandServiceClient is the client API for FooCommandService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FooCommandServiceClient interface {
	PostFoo(ctx context.Context, in *PostFooRequest, opts ...grpc.CallOption) (*PostFooResponse, error)
}

type fooCommandServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewFooCommandServiceClient(cc grpc.ClientConnInterface) FooCommandServiceClient {
	return &fooCommandServiceClient{cc}
}

func (c *fooCommandServiceClient) PostFoo(ctx context.Context, in *PostFooRequest, opts ...grpc.CallOption) (*PostFooResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PostFooResponse)
	err := c.cc.Invoke(ctx, FooCommandService_PostFoo_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FooCommandServiceServer is the server API for FooCommandService service.
// All implementations must embed UnimplementedFooCommandServiceServer
// for forward compatibility
type FooCommandServiceServer interface {
	PostFoo(context.Context, *PostFooRequest) (*PostFooResponse, error)
	mustEmbedUnimplementedFooCommandServiceServer()
}

// UnimplementedFooCommandServiceServer must be embedded to have forward compatible implementations.
type UnimplementedFooCommandServiceServer struct {
}

func (UnimplementedFooCommandServiceServer) PostFoo(context.Context, *PostFooRequest) (*PostFooResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostFoo not implemented")
}
func (UnimplementedFooCommandServiceServer) mustEmbedUnimplementedFooCommandServiceServer() {}

// UnsafeFooCommandServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FooCommandServiceServer will
// result in compilation errors.
type UnsafeFooCommandServiceServer interface {
	mustEmbedUnimplementedFooCommandServiceServer()
}

func RegisterFooCommandServiceServer(s grpc.ServiceRegistrar, srv FooCommandServiceServer) {
	s.RegisterService(&FooCommandService_ServiceDesc, srv)
}

func _FooCommandService_PostFoo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PostFooRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FooCommandServiceServer).PostFoo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FooCommandService_PostFoo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FooCommandServiceServer).PostFoo(ctx, req.(*PostFooRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// FooCommandService_ServiceDesc is the grpc.ServiceDesc for FooCommandService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var FooCommandService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "test.foo.v1.service.FooCommandService",
	HandlerType: (*FooCommandServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PostFoo",
			Handler:    _FooCommandService_PostFoo_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "test/foo/v1/service/foo_service.proto",
}

const (
	FooDownloadService_DownloadRaw_FullMethodName = "/test.foo.v1.service.FooDownloadService/DownloadRaw"
)

// FooDownloadServiceClient is the client API for FooDownloadService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FooDownloadServiceClient interface {
	DownloadRaw(ctx context.Context, in *DownloadRawRequest, opts ...grpc.CallOption) (*httpbody.HttpBody, error)
}

type fooDownloadServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewFooDownloadServiceClient(cc grpc.ClientConnInterface) FooDownloadServiceClient {
	return &fooDownloadServiceClient{cc}
}

func (c *fooDownloadServiceClient) DownloadRaw(ctx context.Context, in *DownloadRawRequest, opts ...grpc.CallOption) (*httpbody.HttpBody, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(httpbody.HttpBody)
	err := c.cc.Invoke(ctx, FooDownloadService_DownloadRaw_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FooDownloadServiceServer is the server API for FooDownloadService service.
// All implementations must embed UnimplementedFooDownloadServiceServer
// for forward compatibility
type FooDownloadServiceServer interface {
	DownloadRaw(context.Context, *DownloadRawRequest) (*httpbody.HttpBody, error)
	mustEmbedUnimplementedFooDownloadServiceServer()
}

// UnimplementedFooDownloadServiceServer must be embedded to have forward compatible implementations.
type UnimplementedFooDownloadServiceServer struct {
}

func (UnimplementedFooDownloadServiceServer) DownloadRaw(context.Context, *DownloadRawRequest) (*httpbody.HttpBody, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DownloadRaw not implemented")
}
func (UnimplementedFooDownloadServiceServer) mustEmbedUnimplementedFooDownloadServiceServer() {}

// UnsafeFooDownloadServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FooDownloadServiceServer will
// result in compilation errors.
type UnsafeFooDownloadServiceServer interface {
	mustEmbedUnimplementedFooDownloadServiceServer()
}

func RegisterFooDownloadServiceServer(s grpc.ServiceRegistrar, srv FooDownloadServiceServer) {
	s.RegisterService(&FooDownloadService_ServiceDesc, srv)
}

func _FooDownloadService_DownloadRaw_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DownloadRawRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FooDownloadServiceServer).DownloadRaw(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FooDownloadService_DownloadRaw_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FooDownloadServiceServer).DownloadRaw(ctx, req.(*DownloadRawRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// FooDownloadService_ServiceDesc is the grpc.ServiceDesc for FooDownloadService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var FooDownloadService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "test.foo.v1.service.FooDownloadService",
	HandlerType: (*FooDownloadServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "DownloadRaw",
			Handler:    _FooDownloadService_DownloadRaw_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "test/foo/v1/service/foo_service.proto",
}
