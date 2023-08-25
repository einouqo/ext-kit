package ws

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/einouqo/ext-kit/test/transport/_service"
	"github.com/einouqo/ext-kit/transport/ws"
)

func prepareTest(address string) (client *ClientBinding, tidy func(), err error) {
	sTydy, err := prepareServer(address)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to prepare server: %+v", err)
	}

	client = prepareClient(address)

	return client, sTydy, nil
}

func prepareClient(address string, opts ...ws.ClientOption) *ClientBinding {
	return NewClientBinding(
		url.URL{
			Scheme: "ws",
			Host:   address,
			Path:   "/stream",
		},
		opts...,
	)
}

func prepareServer(address string, opts ...ws.ServerOption) (tidy func(), err error) {
	sc, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("unable to listen: %+v", err)
	}

	echo := service.NewEcho()
	server := NewServerBinding(echo, opts...)
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/stream", server.Stream)
		_ = http.Serve(sc, mux)
	}()

	return func() {
		_ = sc.Close()
	}, nil
}
