package grpc

import (
	"context"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/einouqo/ext-kit/test/transport/_service"
	"github.com/einouqo/ext-kit/test/transport/grpc/pb"
)

func encodeRequest(_ context.Context, r service.EchoRequest) (proto.Message, error) {
	req := pb.EchoRequest{
		Payload: &pb.EchoPayload{Msg: r.Message},
	}
	if r.Latency > 0 {
		req.Latency = durationpb.New(r.Latency)
	}
	if r.IsError {
		req.Options = &pb.EchoRequest_IsError{IsError: true}
	}
	if r.Repeat > 0 {
		req.Options = &pb.EchoRequest_Repeat{Repeat: uint32(r.Repeat)}
	}
	return &req, nil
}

func decodeRequest(_ context.Context, request proto.Message) (service.EchoRequest, error) {
	req := request.(*pb.EchoRequest)
	r := service.EchoRequest{
		Message: req.Payload.Msg,
		IsError: req.GetIsError(),
		Repeat:  uint(req.GetRepeat()),
	}
	if req.Latency != nil {
		r.Latency = req.Latency.AsDuration()
	}
	return r, nil
}

func encodeResponse(_ context.Context, r service.EchoResponse) (proto.Message, error) {
	resp := pb.EchoResponse{
		Payloads: make([]*pb.EchoPayload, 0, len(r.Messages)),
	}
	for _, msg := range r.Messages {
		resp.Payloads = append(resp.Payloads, &pb.EchoPayload{Msg: msg})
	}
	return &resp, nil
}

func decodeResponse(_ context.Context, response proto.Message) (service.EchoResponse, error) {
	resp := response.(*pb.EchoResponse)
	r := service.EchoResponse{
		Messages: make([]string, 0, len(resp.Payloads)),
	}
	for _, payload := range resp.Payloads {
		r.Messages = append(r.Messages, payload.Msg)
	}
	return r, nil
}
