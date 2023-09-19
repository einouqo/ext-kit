package main

import (
	"context"
	"errors"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/einouqo/ext-kit/endpoint"
	"github.com/einouqo/ext-kit/transport/ws"
)

var (
	ErrInvalidUTF8 = errors.New("invalid utf8")
)

type Service interface {
	Echo(ctx context.Context, receiver <-chan Message) (endpoint.Receive[Message], error)
}

type ServerBindings struct {
	R, M, P http.Handler
}

func NewServerBindings(srv Service) *ServerBindings {
	return &ServerBindings{
		R: ws.NewServer(
			srv.Echo,
			decode,
			encode,
			closer,
			ws.WithServerBefore(upgrade),
			ws.WithServerWriteMod(ws.WriteModPlain),
		),
		M: ws.NewServer(
			srv.Echo,
			decode,
			encode,
			closer,
			ws.WithServerBefore(upgrade),
		),
		P: ws.NewServer(
			srv.Echo,
			decode,
			encode,
			closer,
			ws.WithServerBefore(upgrade),
			ws.WithServerWriteMod(ws.WriteModPrepared),
		),
	}
}

func decode(_ context.Context, messageType ws.MessageType, bytes []byte) (Message, error) {
	if messageType == ws.TextMessageType && !utf8.Valid(bytes) {
		return Message{}, ErrInvalidUTF8
	}
	return Message{Type: messageType, Payload: bytes}, nil
}

func encode(_ context.Context, message Message) ([]byte, ws.MessageType, error) {
	return message.Payload, message.Type, nil
}

func closer(_ context.Context, err error) (code ws.CloseCode, msg string, deadline time.Time) {
	if errors.Is(err, ErrInvalidUTF8) {
		return ws.InvalidFramePayloadDataCloseCode, "", time.Now().Add(time.Second)
	}
	return ws.NormalClosureCloseCode, "", time.Now().Add(time.Second)
}

func upgrade(ctx context.Context, upg ws.Upgrader, _ *http.Request, _ http.Header) context.Context {
	upg.SetReadBufferSize(1 << 12)
	upg.SetWriteBufferSize(1 << 12)
	upg.SetEnableCompression(true)
	upg.SetCheckOrigin(func(r *http.Request) bool { return true })
	return ctx
}
