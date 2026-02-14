package grpc

import (
	"github.com/go-kit/kit/transport"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
)

type ClientOption interface {
	apply(*clientOptions)
}

func WithClientCallOpt(opt grpc.CallOption, opts ...grpc.CallOption) ClientOption {
	return funcClientOption{f: func(o *clientOptions) {
		o.callOpts = append([]grpc.CallOption{opt}, opts...)
	}}
}

func WithClientBefore(before kitgrpc.ClientRequestFunc) ClientOption {
	return funcClientOption{f: func(o *clientOptions) {
		o.before = append(o.before, before)
	}}
}

func WithClientAfter(after kitgrpc.ClientResponseFunc) ClientOption {
	return funcClientOption{f: func(o *clientOptions) {
		o.after = append(o.after, after)
	}}
}

func WithClientFinalizer(finalize kitgrpc.ClientFinalizerFunc) ClientOption {
	return funcClientOption{f: func(o *clientOptions) {
		o.finalizer = append(o.finalizer, finalize)
	}}
}

func WithClientErrorHandler(handler transport.ErrorHandler) ClientOption {
	return funcClientOption{f: func(o *clientOptions) {
		o.errorHandlers = append(o.errorHandlers, handler)
	}}
}

type funcClientOption struct {
	f func(o *clientOptions)
}

func (fo funcClientOption) apply(o *clientOptions) {
	fo.f(o)
}

type ServerOption interface {
	apply(*serverOptions)
}

func WithServerBefore(before kitgrpc.ServerRequestFunc) ServerOption {
	return funcServerOption{f: func(o *serverOptions) {
		o.before = append(o.before, before)
	}}
}

func WithServerAfter(after kitgrpc.ServerResponseFunc) ServerOption {
	return funcServerOption{f: func(o *serverOptions) {
		o.after = append(o.after, after)
	}}
}

func WithServerFinalizer(finalize kitgrpc.ServerFinalizerFunc) ServerOption {
	return funcServerOption{f: func(o *serverOptions) {
		o.finalizer = append(o.finalizer, finalize)
	}}
}

func WithServerErrorHandler(handler transport.ErrorHandler) ServerOption {
	return funcServerOption{f: func(o *serverOptions) {
		o.errorHandlers = append(o.errorHandlers, handler)
	}}
}

type funcServerOption struct {
	f func(o *serverOptions)
}

func (fo funcServerOption) apply(o *serverOptions) {
	fo.f(o)
}
