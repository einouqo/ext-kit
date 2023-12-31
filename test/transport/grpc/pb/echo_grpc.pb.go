// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.23.4
// source: echo.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Echo_Unary_FullMethodName       = "/ext_kit.test.pb.Echo/Unary"
	Echo_InnerStream_FullMethodName = "/ext_kit.test.pb.Echo/InnerStream"
	Echo_OuterStream_FullMethodName = "/ext_kit.test.pb.Echo/OuterStream"
	Echo_BiStream_FullMethodName    = "/ext_kit.test.pb.Echo/BiStream"
)

// EchoClient is the client API for Echo service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EchoClient interface {
	Unary(ctx context.Context, in *EchoRequest, opts ...grpc.CallOption) (*EchoResponse, error)
	InnerStream(ctx context.Context, in *EchoRequest, opts ...grpc.CallOption) (Echo_InnerStreamClient, error)
	OuterStream(ctx context.Context, opts ...grpc.CallOption) (Echo_OuterStreamClient, error)
	BiStream(ctx context.Context, opts ...grpc.CallOption) (Echo_BiStreamClient, error)
}

type echoClient struct {
	cc grpc.ClientConnInterface
}

func NewEchoClient(cc grpc.ClientConnInterface) EchoClient {
	return &echoClient{cc}
}

func (c *echoClient) Unary(ctx context.Context, in *EchoRequest, opts ...grpc.CallOption) (*EchoResponse, error) {
	out := new(EchoResponse)
	err := c.cc.Invoke(ctx, Echo_Unary_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *echoClient) InnerStream(ctx context.Context, in *EchoRequest, opts ...grpc.CallOption) (Echo_InnerStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Echo_ServiceDesc.Streams[0], Echo_InnerStream_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &echoInnerStreamClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Echo_InnerStreamClient interface {
	Recv() (*EchoResponse, error)
	grpc.ClientStream
}

type echoInnerStreamClient struct {
	grpc.ClientStream
}

func (x *echoInnerStreamClient) Recv() (*EchoResponse, error) {
	m := new(EchoResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *echoClient) OuterStream(ctx context.Context, opts ...grpc.CallOption) (Echo_OuterStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Echo_ServiceDesc.Streams[1], Echo_OuterStream_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &echoOuterStreamClient{stream}
	return x, nil
}

type Echo_OuterStreamClient interface {
	Send(*EchoRequest) error
	CloseAndRecv() (*EchoResponse, error)
	grpc.ClientStream
}

type echoOuterStreamClient struct {
	grpc.ClientStream
}

func (x *echoOuterStreamClient) Send(m *EchoRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *echoOuterStreamClient) CloseAndRecv() (*EchoResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(EchoResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *echoClient) BiStream(ctx context.Context, opts ...grpc.CallOption) (Echo_BiStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Echo_ServiceDesc.Streams[2], Echo_BiStream_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &echoBiStreamClient{stream}
	return x, nil
}

type Echo_BiStreamClient interface {
	Send(*EchoRequest) error
	Recv() (*EchoResponse, error)
	grpc.ClientStream
}

type echoBiStreamClient struct {
	grpc.ClientStream
}

func (x *echoBiStreamClient) Send(m *EchoRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *echoBiStreamClient) Recv() (*EchoResponse, error) {
	m := new(EchoResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// EchoServer is the server API for Echo service.
// All implementations must embed UnimplementedEchoServer
// for forward compatibility
type EchoServer interface {
	Unary(context.Context, *EchoRequest) (*EchoResponse, error)
	InnerStream(*EchoRequest, Echo_InnerStreamServer) error
	OuterStream(Echo_OuterStreamServer) error
	BiStream(Echo_BiStreamServer) error
	mustEmbedUnimplementedEchoServer()
}

// UnimplementedEchoServer must be embedded to have forward compatible implementations.
type UnimplementedEchoServer struct {
}

func (UnimplementedEchoServer) Unary(context.Context, *EchoRequest) (*EchoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Unary not implemented")
}
func (UnimplementedEchoServer) InnerStream(*EchoRequest, Echo_InnerStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method InnerStream not implemented")
}
func (UnimplementedEchoServer) OuterStream(Echo_OuterStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method OuterStream not implemented")
}
func (UnimplementedEchoServer) BiStream(Echo_BiStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method BiStream not implemented")
}
func (UnimplementedEchoServer) mustEmbedUnimplementedEchoServer() {}

// UnsafeEchoServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EchoServer will
// result in compilation errors.
type UnsafeEchoServer interface {
	mustEmbedUnimplementedEchoServer()
}

func RegisterEchoServer(s grpc.ServiceRegistrar, srv EchoServer) {
	s.RegisterService(&Echo_ServiceDesc, srv)
}

func _Echo_Unary_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EchoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EchoServer).Unary(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Echo_Unary_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EchoServer).Unary(ctx, req.(*EchoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Echo_InnerStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(EchoRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(EchoServer).InnerStream(m, &echoInnerStreamServer{stream})
}

type Echo_InnerStreamServer interface {
	Send(*EchoResponse) error
	grpc.ServerStream
}

type echoInnerStreamServer struct {
	grpc.ServerStream
}

func (x *echoInnerStreamServer) Send(m *EchoResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Echo_OuterStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(EchoServer).OuterStream(&echoOuterStreamServer{stream})
}

type Echo_OuterStreamServer interface {
	SendAndClose(*EchoResponse) error
	Recv() (*EchoRequest, error)
	grpc.ServerStream
}

type echoOuterStreamServer struct {
	grpc.ServerStream
}

func (x *echoOuterStreamServer) SendAndClose(m *EchoResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *echoOuterStreamServer) Recv() (*EchoRequest, error) {
	m := new(EchoRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Echo_BiStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(EchoServer).BiStream(&echoBiStreamServer{stream})
}

type Echo_BiStreamServer interface {
	Send(*EchoResponse) error
	Recv() (*EchoRequest, error)
	grpc.ServerStream
}

type echoBiStreamServer struct {
	grpc.ServerStream
}

func (x *echoBiStreamServer) Send(m *EchoResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *echoBiStreamServer) Recv() (*EchoRequest, error) {
	m := new(EchoRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Echo_ServiceDesc is the grpc.ServiceDesc for Echo service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Echo_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ext_kit.test.pb.Echo",
	HandlerType: (*EchoServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Unary",
			Handler:    _Echo_Unary_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "InnerStream",
			Handler:       _Echo_InnerStream_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "OuterStream",
			Handler:       _Echo_OuterStream_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "BiStream",
			Handler:       _Echo_BiStream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "echo.proto",
}
