// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: remote-build/remote-build.proto

package remote_build

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	MicService_StartBuild_FullMethodName = "/remote_build.MicService/StartBuild"
)

// MicServiceClient is the client API for MicService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// Client requests a build from the server
type MicServiceClient interface {
	StartBuild(ctx context.Context, in *BuildRequest, opts ...grpc.CallOption) (*BuildResponse, error)
}

type micServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMicServiceClient(cc grpc.ClientConnInterface) MicServiceClient {
	return &micServiceClient{cc}
}

func (c *micServiceClient) StartBuild(ctx context.Context, in *BuildRequest, opts ...grpc.CallOption) (*BuildResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(BuildResponse)
	err := c.cc.Invoke(ctx, MicService_StartBuild_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MicServiceServer is the server API for MicService service.
// All implementations must embed UnimplementedMicServiceServer
// for forward compatibility.
//
// Client requests a build from the server
type MicServiceServer interface {
	StartBuild(context.Context, *BuildRequest) (*BuildResponse, error)
	mustEmbedUnimplementedMicServiceServer()
}

// UnimplementedMicServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedMicServiceServer struct{}

func (UnimplementedMicServiceServer) StartBuild(context.Context, *BuildRequest) (*BuildResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartBuild not implemented")
}
func (UnimplementedMicServiceServer) mustEmbedUnimplementedMicServiceServer() {}
func (UnimplementedMicServiceServer) testEmbeddedByValue()                    {}

// UnsafeMicServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MicServiceServer will
// result in compilation errors.
type UnsafeMicServiceServer interface {
	mustEmbedUnimplementedMicServiceServer()
}

func RegisterMicServiceServer(s grpc.ServiceRegistrar, srv MicServiceServer) {
	// If the following call pancis, it indicates UnimplementedMicServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&MicService_ServiceDesc, srv)
}

func _MicService_StartBuild_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BuildRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MicServiceServer).StartBuild(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MicService_StartBuild_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MicServiceServer).StartBuild(ctx, req.(*BuildRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// MicService_ServiceDesc is the grpc.ServiceDesc for MicService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MicService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "remote_build.MicService",
	HandlerType: (*MicServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "StartBuild",
			Handler:    _MicService_StartBuild_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "remote-build/remote-build.proto",
}

const (
	WorkerService_ProcessWork_FullMethodName = "/remote_build.WorkerService/ProcessWork"
)

// WorkerServiceClient is the client API for WorkerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// Server sends a work request to the worker
type WorkerServiceClient interface {
	ProcessWork(ctx context.Context, in *WorkRequest, opts ...grpc.CallOption) (*WorkResponse, error)
}

type workerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewWorkerServiceClient(cc grpc.ClientConnInterface) WorkerServiceClient {
	return &workerServiceClient{cc}
}

func (c *workerServiceClient) ProcessWork(ctx context.Context, in *WorkRequest, opts ...grpc.CallOption) (*WorkResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(WorkResponse)
	err := c.cc.Invoke(ctx, WorkerService_ProcessWork_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WorkerServiceServer is the server API for WorkerService service.
// All implementations must embed UnimplementedWorkerServiceServer
// for forward compatibility.
//
// Server sends a work request to the worker
type WorkerServiceServer interface {
	ProcessWork(context.Context, *WorkRequest) (*WorkResponse, error)
	mustEmbedUnimplementedWorkerServiceServer()
}

// UnimplementedWorkerServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedWorkerServiceServer struct{}

func (UnimplementedWorkerServiceServer) ProcessWork(context.Context, *WorkRequest) (*WorkResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProcessWork not implemented")
}
func (UnimplementedWorkerServiceServer) mustEmbedUnimplementedWorkerServiceServer() {}
func (UnimplementedWorkerServiceServer) testEmbeddedByValue()                       {}

// UnsafeWorkerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WorkerServiceServer will
// result in compilation errors.
type UnsafeWorkerServiceServer interface {
	mustEmbedUnimplementedWorkerServiceServer()
}

func RegisterWorkerServiceServer(s grpc.ServiceRegistrar, srv WorkerServiceServer) {
	// If the following call pancis, it indicates UnimplementedWorkerServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&WorkerService_ServiceDesc, srv)
}

func _WorkerService_ProcessWork_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WorkRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerServiceServer).ProcessWork(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorkerService_ProcessWork_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerServiceServer).ProcessWork(ctx, req.(*WorkRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// WorkerService_ServiceDesc is the grpc.ServiceDesc for WorkerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var WorkerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "remote_build.WorkerService",
	HandlerType: (*WorkerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ProcessWork",
			Handler:    _WorkerService_ProcessWork_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "remote-build/remote-build.proto",
}
