package grpc

import (
	"context"

	"github.com/einouqo/ext-kit/endpoint"
	"github.com/einouqo/ext-kit/test/transport/_service"
	"github.com/einouqo/ext-kit/test/transport/grpc/pb"
	kitgrpc "github.com/einouqo/ext-kit/transport/grpc"
)

type Service interface {
	Once(ctx context.Context, req service.EchoRequest) (service.EchoResponse, error)
	Inner(ctx context.Context, req service.EchoRequest) (endpoint.Receive[service.EchoResponse], error)
	Outer(ctx context.Context, receiver <-chan service.EchoRequest) (service.EchoResponse, error)
	Bi(ctx context.Context, receiver <-chan service.EchoRequest) (endpoint.Receive[service.EchoResponse], error)
}

type ServerBinding struct {
	pb.UnimplementedEchoServer

	unary       kitgrpc.HandlerUnary
	innerStream kitgrpc.HandlerInnerStream
	outerStream kitgrpc.HandlerOuterStream
	biStream    kitgrpc.HandlerBiStream
}

func NewServerBinding(svc Service, opts ...kitgrpc.ServerOption) *ServerBinding {
	return &ServerBinding{
		unary: kitgrpc.NewServerUnary(
			svc.Once,
			decodeRequest,
			encodeResponse,
			opts...,
		),
		innerStream: kitgrpc.NewServerInnerStream(
			svc.Inner,
			decodeRequest,
			encodeResponse,
			opts...,
		),
		outerStream: kitgrpc.NewServerOuterStream[*pb.EchoRequest](
			svc.Outer,
			decodeRequest,
			encodeResponse,
			opts...,
		),
		biStream: kitgrpc.NewServerBiStream[*pb.EchoRequest](
			svc.Bi,
			decodeRequest,
			encodeResponse,
			opts...,
		),
	}
}

func (sb ServerBinding) Unary(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	_, response, err := sb.unary.ServeUnary(ctx, req)
	if err != nil {
		return nil, err
	}
	return response.(*pb.EchoResponse), nil
}
func (sb ServerBinding) InnerStream(req *pb.EchoRequest, s pb.Echo_InnerStreamServer) error {
	_, err := sb.innerStream.ServeInnerStream(req, s)
	return err
}
func (sb ServerBinding) OuterStream(s pb.Echo_OuterStreamServer) error {
	_, err := sb.outerStream.ServeOuterStream(s)
	return err
}
func (sb ServerBinding) BiStream(s pb.Echo_BiStreamServer) error {
	_, err := sb.biStream.ServeBiStream(s)
	return err
}
