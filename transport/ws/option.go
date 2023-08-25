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

func WithClientBefore(before ClientHeaderFunc) ClientOption {
	return funcClientOption{f: func(o *clientOptions) {
		o.before = append(o.before, before)
	}}
}

func WithClientAfter(after ClientResponseFunc) ClientOption {
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
		o.enhancement.config.write.timeout = timeout
	}}
}

func WithClientWriteCompression(level int) ClientOption {
	return funcClientOption{f: func(o *clientOptions) {
		o.enhancement.preset.write.compression.enable = true
		o.enhancement.preset.write.compression.level = level
	}}
}

func WithClientReadTimeout(timeout time.Duration) ClientOption {
	return funcClientOption{f: func(o *clientOptions) {
		o.enhancement.config.read.timeout = timeout
	}}
}

func WithClientReadLimit(limit int64) ClientOption {
	return funcClientOption{f: func(o *clientOptions) {
		o.enhancement.preset.read.limit = limit
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

func WithServerBefore(before ServerHeaderFunc) ServerOption {
	return funcServerOption{f: func(o *serverOptions) {
		o.before = append(o.before, before)
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
		o.enhancement.config.read.timeout = timeout
	}}
}

func WithServerReadLimit(limit int64) ServerOption {
	return funcServerOption{f: func(o *serverOptions) {
		o.enhancement.preset.read.limit = limit
	}}
}

func WithServerWriteTimeout(timeout time.Duration) ServerOption {
	return funcServerOption{f: func(o *serverOptions) {
		o.enhancement.config.write.timeout = timeout
	}}
}

func WithServerWriteCompression(level int) ServerOption {
	return funcServerOption{f: func(o *serverOptions) {
		o.enhancement.preset.write.compression.enable = true
		o.enhancement.preset.write.compression.level = level
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
