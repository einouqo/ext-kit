package ws

import (
	"context"
	"net/http"
	"time"

	"github.com/einouqo/ext-kit/endpoint"
	"github.com/einouqo/ext-kit/test/transport/_service"
	"github.com/einouqo/ext-kit/transport/ws"
)

type Service interface {
	Bi(ctx context.Context, receiver <-chan service.EchoRequest) (endpoint.Receive[service.EchoResponse], error)
}

type ServerBinding struct {
	Stream http.Handler
}

func NewServerBinding(svc Service, opts ...ws.ServerOption) *ServerBinding {
	return &ServerBinding{
		Stream: ws.NewServer(
			svc.Bi,
			decodeRequest,
			encodeResponse,
			sCloser,
			opts...,
		),
	}
}

func sCloser(_ context.Context, err error) (code ws.CloseCode, msg string, deadline time.Time) {
	if err != nil {
		return ws.InternalServerErrCloseCode, err.Error(), time.Now().Add(time.Second)
	}
	return ws.NormalClosureCloseCode, "", time.Now().Add(time.Second)
}
