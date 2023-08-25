package ws

import (
	"context"
	"net"
	"net/http"
	"time"
)

type Closure func(context.Context, error) (code CloseCode, msg string, deadline time.Time)

type Pinging func(context.Context) (msg []byte, deadline time.Time)

type ClientHeaderFunc func(context.Context, *http.Header) context.Context

type ClientResponseFunc func(context.Context, *http.Response) context.Context

type ServerHeaderFunc func(ctx context.Context, req http.Header, upg *http.Header) context.Context

type ServerAddressFunc func(ctx context.Context, local net.Addr, remote net.Addr) context.Context
