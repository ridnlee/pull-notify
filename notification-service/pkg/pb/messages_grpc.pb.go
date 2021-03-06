// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

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

// NotifyerClient is the client API for Notifyer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NotifyerClient interface {
	GetLastMsgs(ctx context.Context, in *GetLastMsgsRequest, opts ...grpc.CallOption) (*GetLastMsgsResponse, error)
	GetMsgList(ctx context.Context, in *GetMsgListRequest, opts ...grpc.CallOption) (*GetMsgListResponse, error)
	MarkRead(ctx context.Context, in *MarkReadRequest, opts ...grpc.CallOption) (*MarkReadResponse, error)
}

type notifyerClient struct {
	cc grpc.ClientConnInterface
}

func NewNotifyerClient(cc grpc.ClientConnInterface) NotifyerClient {
	return &notifyerClient{cc}
}

func (c *notifyerClient) GetLastMsgs(ctx context.Context, in *GetLastMsgsRequest, opts ...grpc.CallOption) (*GetLastMsgsResponse, error) {
	out := new(GetLastMsgsResponse)
	err := c.cc.Invoke(ctx, "/proto.Notifyer/GetLastMsgs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notifyerClient) GetMsgList(ctx context.Context, in *GetMsgListRequest, opts ...grpc.CallOption) (*GetMsgListResponse, error) {
	out := new(GetMsgListResponse)
	err := c.cc.Invoke(ctx, "/proto.Notifyer/GetMsgList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notifyerClient) MarkRead(ctx context.Context, in *MarkReadRequest, opts ...grpc.CallOption) (*MarkReadResponse, error) {
	out := new(MarkReadResponse)
	err := c.cc.Invoke(ctx, "/proto.Notifyer/MarkRead", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NotifyerServer is the server API for Notifyer service.
// All implementations must embed UnimplementedNotifyerServer
// for forward compatibility
type NotifyerServer interface {
	GetLastMsgs(context.Context, *GetLastMsgsRequest) (*GetLastMsgsResponse, error)
	GetMsgList(context.Context, *GetMsgListRequest) (*GetMsgListResponse, error)
	MarkRead(context.Context, *MarkReadRequest) (*MarkReadResponse, error)
	mustEmbedUnimplementedNotifyerServer()
}

// UnimplementedNotifyerServer must be embedded to have forward compatible implementations.
type UnimplementedNotifyerServer struct {
}

func (UnimplementedNotifyerServer) GetLastMsgs(context.Context, *GetLastMsgsRequest) (*GetLastMsgsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLastMsgs not implemented")
}
func (UnimplementedNotifyerServer) GetMsgList(context.Context, *GetMsgListRequest) (*GetMsgListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMsgList not implemented")
}
func (UnimplementedNotifyerServer) MarkRead(context.Context, *MarkReadRequest) (*MarkReadResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MarkRead not implemented")
}
func (UnimplementedNotifyerServer) mustEmbedUnimplementedNotifyerServer() {}

// UnsafeNotifyerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NotifyerServer will
// result in compilation errors.
type UnsafeNotifyerServer interface {
	mustEmbedUnimplementedNotifyerServer()
}

func RegisterNotifyerServer(s grpc.ServiceRegistrar, srv NotifyerServer) {
	s.RegisterService(&Notifyer_ServiceDesc, srv)
}

func _Notifyer_GetLastMsgs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetLastMsgsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotifyerServer).GetLastMsgs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Notifyer/GetLastMsgs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotifyerServer).GetLastMsgs(ctx, req.(*GetLastMsgsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Notifyer_GetMsgList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMsgListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotifyerServer).GetMsgList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Notifyer/GetMsgList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotifyerServer).GetMsgList(ctx, req.(*GetMsgListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Notifyer_MarkRead_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MarkReadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotifyerServer).MarkRead(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Notifyer/MarkRead",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotifyerServer).MarkRead(ctx, req.(*MarkReadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Notifyer_ServiceDesc is the grpc.ServiceDesc for Notifyer service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Notifyer_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.Notifyer",
	HandlerType: (*NotifyerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetLastMsgs",
			Handler:    _Notifyer_GetLastMsgs_Handler,
		},
		{
			MethodName: "GetMsgList",
			Handler:    _Notifyer_GetMsgList_Handler,
		},
		{
			MethodName: "MarkRead",
			Handler:    _Notifyer_MarkRead_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/messages.proto",
}
