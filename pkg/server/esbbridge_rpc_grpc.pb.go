// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package server

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

// EsbBridgeClient is the client API for EsbBridge service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EsbBridgeClient interface {
	// Transfers an ESB message to a peripheral device and returns the anwser
	Transfer(ctx context.Context, in *EsbMessage, opts ...grpc.CallOption) (*EsbMessage, error)
	// Starts listening for specific packages. Server will send matching messages async to the client
	Listen(ctx context.Context, in *Listener, opts ...grpc.CallOption) (EsbBridge_ListenClient, error)
}

type esbBridgeClient struct {
	cc grpc.ClientConnInterface
}

func NewEsbBridgeClient(cc grpc.ClientConnInterface) EsbBridgeClient {
	return &esbBridgeClient{cc}
}

func (c *esbBridgeClient) Transfer(ctx context.Context, in *EsbMessage, opts ...grpc.CallOption) (*EsbMessage, error) {
	out := new(EsbMessage)
	err := c.cc.Invoke(ctx, "/server.EsbBridge/Transfer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *esbBridgeClient) Listen(ctx context.Context, in *Listener, opts ...grpc.CallOption) (EsbBridge_ListenClient, error) {
	stream, err := c.cc.NewStream(ctx, &EsbBridge_ServiceDesc.Streams[0], "/server.EsbBridge/Listen", opts...)
	if err != nil {
		return nil, err
	}
	x := &esbBridgeListenClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type EsbBridge_ListenClient interface {
	Recv() (*EsbMessage, error)
	grpc.ClientStream
}

type esbBridgeListenClient struct {
	grpc.ClientStream
}

func (x *esbBridgeListenClient) Recv() (*EsbMessage, error) {
	m := new(EsbMessage)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// EsbBridgeServer is the server API for EsbBridge service.
// All implementations must embed UnimplementedEsbBridgeServer
// for forward compatibility
type EsbBridgeServer interface {
	// Transfers an ESB message to a peripheral device and returns the anwser
	Transfer(context.Context, *EsbMessage) (*EsbMessage, error)
	// Starts listening for specific packages. Server will send matching messages async to the client
	Listen(*Listener, EsbBridge_ListenServer) error
	mustEmbedUnimplementedEsbBridgeServer()
}

// UnimplementedEsbBridgeServer must be embedded to have forward compatible implementations.
type UnimplementedEsbBridgeServer struct {
}

func (UnimplementedEsbBridgeServer) Transfer(context.Context, *EsbMessage) (*EsbMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Transfer not implemented")
}
func (UnimplementedEsbBridgeServer) Listen(*Listener, EsbBridge_ListenServer) error {
	return status.Errorf(codes.Unimplemented, "method Listen not implemented")
}
func (UnimplementedEsbBridgeServer) mustEmbedUnimplementedEsbBridgeServer() {}

// UnsafeEsbBridgeServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EsbBridgeServer will
// result in compilation errors.
type UnsafeEsbBridgeServer interface {
	mustEmbedUnimplementedEsbBridgeServer()
}

func RegisterEsbBridgeServer(s grpc.ServiceRegistrar, srv EsbBridgeServer) {
	s.RegisterService(&EsbBridge_ServiceDesc, srv)
}

func _EsbBridge_Transfer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EsbMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EsbBridgeServer).Transfer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/server.EsbBridge/Transfer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EsbBridgeServer).Transfer(ctx, req.(*EsbMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _EsbBridge_Listen_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Listener)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(EsbBridgeServer).Listen(m, &esbBridgeListenServer{stream})
}

type EsbBridge_ListenServer interface {
	Send(*EsbMessage) error
	grpc.ServerStream
}

type esbBridgeListenServer struct {
	grpc.ServerStream
}

func (x *esbBridgeListenServer) Send(m *EsbMessage) error {
	return x.ServerStream.SendMsg(m)
}

// EsbBridge_ServiceDesc is the grpc.ServiceDesc for EsbBridge service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var EsbBridge_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "server.EsbBridge",
	HandlerType: (*EsbBridgeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Transfer",
			Handler:    _EsbBridge_Transfer_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Listen",
			Handler:       _EsbBridge_Listen_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "pkg/server/esbbridge_rpc.proto",
}