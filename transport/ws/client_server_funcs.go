package ws

import (
	"context"
	"net/http"
	"time"
)

type Closure func(context.Context, error) (code CloseCode, msg string, deadline time.Time)

type Pinging func(context.Context) (msg []byte, deadline time.Time)

type DiallerFunc func(context.Context, Dialler, *http.Header) context.Context

type UpgradeFunc func(context.Context, Upgrader, *http.Request, *http.Header) context.Context

type ClientTunerFunc func(context.Context, *http.Response, Tuner) context.Context

type ServerTunerFunc func(context.Context, Tuner) context.Context
