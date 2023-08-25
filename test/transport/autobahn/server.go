package main

import (
	"context"
	"net/http"
	"time"

	"github.com/einouqo/ext-kit/endpoint"
	"github.com/einouqo/ext-kit/transport/ws"
)

type Service interface {
	Echo(ctx context.Context, receiver <-chan Message) (endpoint.Receive[Message], error)
}

type ServerBindings struct {
	M, P http.Handler
}

func NewServerBindings(srv Service) *ServerBindings {
	return &ServerBindings{
		M: ws.NewServer(srv.Echo, decode, encode, closer),
		P: ws.NewServer(srv.Echo, decode, encode, closer, ws.WithServerPreparedWrites()),
	}
}

func decode(ctx context.Context, messageType ws.MessageType, bytes []byte) (Message, error) {
	return Message{Type: messageType, Payload: bytes}, nil
}

func encode(ctx context.Context, message Message) ([]byte, ws.MessageType, error) {
	return message.Payload, message.Type, nil
}

func closer(context.Context, error) (code ws.CloseCode, msg string, deadline time.Time) {
	return ws.NormalClosureCloseCode, "", time.Now().Add(time.Second)
}
