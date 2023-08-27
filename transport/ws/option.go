package ws

import (
	"time"

	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
)

type ClientOption interface {
	apply(*clientOptions)
}

func WithClientBefore(before DiallerFunc) ClientOption {
	return funcClientOption{f: func(o *clientOptions) {
		o.before = append(o.before, before)
	}}
}

func WithClientAfter(after ClientTunerFunc) ClientOption {
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

func WithClientPreparedWrites() ClientOption {
	return funcClientOption{f: func(o *clientOptions) {
		o.enhancement.config.write.prepared = true
	}}
}

func WithClientReadTimeout(timeout time.Duration) ClientOption {
	return funcClientOption{f: func(o *clientOptions) {
		o.enhancement.config.read.timeout = timeout
	}}
}

func WithClientPing(period, await time.Duration, pinging Pinging) ClientOption {
	return funcClientOption{f: func(o *clientOptions) {
		o.heartbeat.enable = true
		o.heartbeat.config.period = period
		o.heartbeat.config.await = await
		o.heartbeat.config.pinging = pinging
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

func WithServerBefore(before UpgradeFunc) ServerOption {
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

func WithServerWriteTimeout(timeout time.Duration) ServerOption {
	return funcServerOption{f: func(o *serverOptions) {
		o.enhancement.config.write.timeout = timeout
	}}
}

func WithServerPreparedWrites() ServerOption {
	return funcServerOption{f: func(o *serverOptions) {
		o.enhancement.config.write.prepared = true
	}}
}

func WithServerPing(period, await time.Duration, pinging Pinging) ServerOption {
	return funcServerOption{f: func(o *serverOptions) {
		o.heartbeat.enable = true
		o.heartbeat.config.period = period
		o.heartbeat.config.await = await
		o.heartbeat.config.pinging = pinging
	}}
}

type funcServerOption struct {
	f func(o *serverOptions)
}

func (fo funcServerOption) apply(o *serverOptions) {
	fo.f(o)
}
