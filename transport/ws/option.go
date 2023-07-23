package ws

import (
	"time"

	"github.com/fasthttp/websocket"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
)

type ClientOption interface {
	apply(*clientOptions)
}

func WithClientDialler(dialer *websocket.Dialer) ClientOption {
	return funcClientOption{f: func(o *clientOptions) {
		o.dialer = dialer
	}}
}

func WithClientBefore(before ClientRequestFunc) ClientOption {
	return funcClientOption{f: func(o *clientOptions) {
		o.before = append(o.before, before)
	}}
}

func WithClientAfter(after ClientConnectionFunc) ClientOption {
	return funcClientOption{f: func(o *clientOptions) {
		o.after = append(o.after, after)
	}}
}

func WithClientFinalizer(finalizer kithttp.ClientFinalizerFunc) ClientOption {
	return funcClientOption{f: func(o *clientOptions) {
		o.finalizer = append(o.finalizer, finalizer)
	}}
}

func WithClientErrorHandler(handler transport.ErrorHandler) ClientOption {
	return funcClientOption{f: func(o *clientOptions) {
		o.errHandlers = append(o.errHandlers, handler)
	}}
}

func WithClientWriteTimeout(timeout time.Duration) ClientOption {
	return funcClientOption{f: func(o *clientOptions) {
		o.timeout.write = timeout
	}}
}

func WithClientReadTimeout(timeout time.Duration) ClientOption {
	return funcClientOption{f: func(o *clientOptions) {
		o.timeout.read = timeout
	}}
}

func WithClientPing(period, await time.Duration, pinging Pinging) ClientOption {
	return funcClientOption{f: func(o *clientOptions) {
		o.heartbeat.enable = true
		o.heartbeat.period = period
		o.heartbeat.await = await
		o.heartbeat.pinging = pinging
	}}
}

var withClientNothing = funcClientOption{f: func(*clientOptions) {}}

type funcClientOption struct {
	f func(o *clientOptions)
}

func (fo funcClientOption) apply(o *clientOptions) {
	fo.f(o)
}

type ServerOption interface {
	apply(*serverOptions)
}

func WithServerUpgrader(upgrader *websocket.Upgrader) ServerOption {
	return funcServerOption{f: func(o *serverOptions) {
		o.upgrader = upgrader
	}}
}

func WithServerBefore(before ServerRequestFunc) ServerOption {
	return funcServerOption{f: func(o *serverOptions) {
		o.before = append(o.before, before)
	}}
}

func WithServerAfter(after ServerConnectionFunc) ServerOption {
	return funcServerOption{f: func(o *serverOptions) {
		o.after = append(o.after, after)
	}}
}

func WithServerFinalizer(finalizer kithttp.ServerFinalizerFunc) ServerOption {
	return funcServerOption{f: func(o *serverOptions) {
		o.finalizer = append(o.finalizer, finalizer)
	}}
}

func WithServerErrorHandler(handler transport.ErrorHandler) ServerOption {
	return funcServerOption{f: func(o *serverOptions) {
		o.errHandlers = append(o.errHandlers, handler)
	}}
}

func WithServerReadTimeout(timeout time.Duration) ServerOption {
	return funcServerOption{f: func(o *serverOptions) {
		o.timeout.read = timeout
	}}
}

func WithServerWriteTimeout(timeout time.Duration) ServerOption {
	return funcServerOption{f: func(o *serverOptions) {
		o.timeout.write = timeout
	}}
}

func WithServerPing(period, await time.Duration, pinging Pinging) ServerOption {
	return funcServerOption{f: func(o *serverOptions) {
		o.heartbeat.enable = true
		o.heartbeat.period = period
		o.heartbeat.await = await
		o.heartbeat.pinging = pinging
	}}
}

type funcServerOption struct {
	f func(o *serverOptions)
}

func (fo funcServerOption) apply(o *serverOptions) {
	fo.f(o)
}
