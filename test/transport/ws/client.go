package ws

import (
	"context"
	"net/url"
	"time"

	"github.com/einouqo/ext-kit/endpoint"
	"github.com/einouqo/ext-kit/test/transport/_service"
	"github.com/einouqo/ext-kit/transport/ws"
)

type ClientBinding struct {
	Stream endpoint.BiStream[service.EchoRequest, service.EchoResponse]
}

func NewClientBinding(url url.URL, opts ...ws.ClientOption) *ClientBinding {
	return &ClientBinding{
		Stream: ws.NewClient(
			url,
			encodeRequest,
			decodeResponse,
			cCloser,
			opts...,
		).Endpoint(),
	}
}

func cCloser(context.Context, error) (code ws.CloseCode, msg string, deadline time.Time) {
	return ws.NormalClosureCloseCode, "", time.Now().Add(time.Second)
}
