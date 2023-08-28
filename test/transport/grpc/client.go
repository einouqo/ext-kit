package grpc

import (
	"google.golang.org/grpc"

	"github.com/einouqo/ext-kit/endpoint"
	"github.com/einouqo/ext-kit/test/transport/_service"
	"github.com/einouqo/ext-kit/test/transport/grpc/pb"
	kitgrpc "github.com/einouqo/ext-kit/transport/grpc"
)

type ClientBinding struct {
	Unary       endpoint.Unary[service.EchoRequest, service.EchoResponse]
	InnerStream endpoint.InnerStream[service.EchoRequest, service.EchoResponse]
	OuterStream endpoint.OuterStream[service.EchoRequest, service.EchoResponse]
	BiStream    endpoint.BiStream[service.EchoRequest, service.EchoResponse]
}

func NewClientBinding(cc *grpc.ClientConn, opts ...kitgrpc.ClientOption) *ClientBinding {
	return &ClientBinding{
		Unary: kitgrpc.NewClientUnary[*pb.EchoResponse](
			cc,
			pb.Echo_Unary_FullMethodName,
			encodeRequest,
			decodeResponse,
			opts...,
		).Endpoint(),
		InnerStream: kitgrpc.NewClientInnerStream[*pb.EchoResponse](
			cc,
			pb.Echo_InnerStream_FullMethodName,
			encodeRequest,
			decodeResponse,
			opts...,
		).Endpoint(),
		OuterStream: kitgrpc.NewClientOuterStream[*pb.EchoResponse](
			cc,
			pb.Echo_OuterStream_FullMethodName,
			encodeRequest,
			decodeResponse,
			opts...,
		).Endpoint(),
		BiStream: kitgrpc.NewClientBiStream[*pb.EchoResponse](
			cc,
			pb.Echo_BiStream_FullMethodName,
			encodeRequest,
			decodeResponse,
			opts...,
		).Endpoint(),
	}
}
