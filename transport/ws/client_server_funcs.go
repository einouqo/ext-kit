package ws

import (
	"context"
	"net/http"
	"time"
)

type Closure func(context.Context, error) (code CloseCode, msg string, deadline time.Time)

type Pinging func(context.Context) (msg []byte, deadline time.Time)

type ClientRequestFunc func(context.Context, *http.Header) context.Context

type ClientConnectionFunc func(context.Context, Connection) context.Context

type ServerRequestFunc func(ctx context.Context, req http.Header, upg *http.Header) context.Context

type ServerConnectionFunc func(context.Context, Connection) context.Context
