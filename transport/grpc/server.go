package grpc

import (
	"context"
	"errors"
	"io"
	"reflect"

	"github.com/go-kit/kit/transport"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	"github.com/einouqo/ext-kit/endpoint"
)

type HandlerUnary interface {
	ServeUnary(context.Context, proto.Message) (context.Context, proto.Message, error)
}

type HandlerInnerStream interface {
	ServeInnerStream(proto.Message, grpc.ServerStream) (context.Context, error)
}

type HandlerOuterStream interface {
	ServeOuterStream(grpc.ServerStream) (context.Context, error)
}

type HandlerBiStream interface {
	ServeBiStream(grpc.ServerStream) (context.Context, error)
}

type ServerUnary[IN, OUT any] struct {
	server[IN, OUT, endpoint.Unary[IN, OUT]]
}

func NewServerUnary[IN, OUT any](
	e endpoint.Unary[IN, OUT],
	dec DecodeFunc[IN],
	enc EncodeFunc[OUT],
	opts ...ServerOption,
) *ServerUnary[IN, OUT] {
	s := &ServerUnary[IN, OUT]{
		server: server[IN, OUT, endpoint.Unary[IN, OUT]]{
			e:   e,
			dec: dec,
			enc: enc,
		},
	}
	for _, opt := range opts {
		opt.apply(&s.opts)
	}
	return s
}

func (srv ServerUnary[IN, OUT]) ServeUnary(ctx context.Context, req proto.Message) (ctxt context.Context, resp proto.Message, err error) {
	if len(srv.opts.finalizer) > 0 {
		defer func() {
			for _, f := range srv.opts.finalizer {
				f(ctx, err)
			}
		}()
	}
	if len(srv.opts.errHandlers) > 0 {
		defer func() {
			if err != nil {
				for _, h := range srv.opts.errHandlers {
					h.Handle(ctx, err)
				}
			}
		}()
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	for _, f := range srv.opts.before {
		ctx = f(ctx, md)
	}

	in, err := srv.dec(ctx, req)
	if err != nil {
		return ctx, nil, err
	}

	out, err := srv.e(ctx, in)
	if err != nil {
		return ctx, nil, err
	}

	mdHeader, mdTrailer := make(metadata.MD), make(metadata.MD)
	for _, f := range srv.opts.after {
		ctx = f(ctx, &mdHeader, &mdTrailer)
	}
	if len(mdHeader) > 0 {
		if err = grpc.SendHeader(ctx, mdHeader); err != nil {
			return ctx, nil, err
		}
	}

	resp, err = srv.enc(ctx, out)
	if err != nil {
		return ctx, nil, err
	}

	if len(mdTrailer) > 0 {
		if err = grpc.SetTrailer(ctx, mdTrailer); err != nil {
			return ctx, nil, err
		}
	}

	return ctx, resp, nil
}

type ServerInnerStream[IN, OUT any] struct {
	server[IN, OUT, endpoint.InnerStream[IN, OUT]]
}

func NewServerInnerStream[IN, OUT any](
	e endpoint.InnerStream[IN, OUT],
	dec DecodeFunc[IN],
	enc EncodeFunc[OUT],
	opts ...ServerOption,
) *ServerInnerStream[IN, OUT] {
	s := &ServerInnerStream[IN, OUT]{
		server: server[IN, OUT, endpoint.InnerStream[IN, OUT]]{
			e:   e,
			dec: dec,
			enc: enc,
		},
	}
	for _, opt := range opts {
		opt.apply(&s.opts)
	}
	return s
}

func (srv ServerInnerStream[IN, OUT]) ServeInnerStream(req proto.Message, s grpc.ServerStream) (ctx context.Context, err error) {
	ctx = s.Context()

	if len(srv.opts.finalizer) > 0 {
		defer func() {
			for _, f := range srv.opts.finalizer {
				f(ctx, err)
			}
		}()
	}
	if len(srv.opts.errHandlers) > 0 {
		defer func() {
			if err != nil {
				for _, h := range srv.opts.errHandlers {
					h.Handle(ctx, err)
				}
			}
		}()
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	for _, f := range srv.opts.before {
		ctx = f(ctx, md)
	}

	in, err := srv.dec(ctx, req)
	if err != nil {
		return ctx, err
	}

	receive, err := srv.e(ctx, in)
	if err != nil {
		return ctx, err
	}

	mdHeader, mdTrailer := make(metadata.MD), make(metadata.MD)
	for _, f := range srv.opts.after {
		ctx = f(ctx, &mdHeader, &mdTrailer)
	}

	group := errgroup.Group{}
	group.Go(func() error {
		if len(mdHeader) > 0 {
			if err = grpc.SendHeader(ctx, mdHeader); err != nil {
				return err
			}
		}
		for {
			out, err := receive()
			switch {
			case errors.Is(err, endpoint.StreamDone):
				return nil
			case err != nil:
				return err
			}
			msg, err := srv.enc(ctx, out)
			if err != nil {
				return err
			}
			err = s.SendMsg(msg)
			switch {
			case errors.Is(err, io.EOF):
				return io.ErrUnexpectedEOF
			case err != nil:
				return err
			}
		}
	})
	if err := group.Wait(); err != nil {
		return ctx, err
	}

	if len(mdTrailer) > 0 {
		if err = grpc.SetTrailer(ctx, mdTrailer); err != nil {
			return ctx, err
		}
	}

	return ctx, nil
}

type ServerOuterStream[IN, OUT any] struct {
	server[IN, OUT, endpoint.OuterStream[IN, OUT]]
}

func NewServerOuterStream[RECEIVE proto.Message, IN, OUT any](
	e endpoint.OuterStream[IN, OUT],
	dec DecodeFunc[IN],
	enc EncodeFunc[OUT],
	opts ...ServerOption,
) *ServerOuterStream[IN, OUT] {
	var receive RECEIVE
	s := &ServerOuterStream[IN, OUT]{
		server: server[IN, OUT, endpoint.OuterStream[IN, OUT]]{
			e:   e,
			dec: dec,
			enc: enc,
			reflectReceive: reflect.New(
				reflect.TypeOf(receive).Elem(),
			),
		},
	}
	for _, opt := range opts {
		opt.apply(&s.opts)
	}
	return s
}

func (srv ServerOuterStream[IN, OUT]) ServeOuterStream(s grpc.ServerStream) (ctx context.Context, err error) {
	ctx = s.Context()

	if len(srv.opts.finalizer) > 0 {
		defer func() {
			for _, f := range srv.opts.finalizer {
				f(ctx, err)
			}
		}()
	}
	if len(srv.opts.errHandlers) > 0 {
		defer func() {
			if err != nil {
				for _, h := range srv.opts.errHandlers {
					h.Handle(ctx, err)
				}
			}
		}()
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	for _, f := range srv.opts.before {
		ctx = f(ctx, md)
	}

	inCh := make(chan IN)
	doneCh := make(chan struct{})
	group := errgroup.Group{}
	group.Go(func() error {
		defer close(inCh)
		for {
			msg := srv.reflectReceive.Interface().(proto.Message)
			err := s.RecvMsg(msg)
			switch {
			case errors.Is(err, io.EOF):
				return nil
			case err != nil:
				return err
			}
			in, err := srv.dec(ctx, msg)
			if err != nil {
				return err
			}
			select {
			case <-doneCh:
				return nil
			case inCh <- in:
			}
		}
	})
	group.Go(func() error {
		defer close(doneCh)
		out, err := srv.e(ctx, inCh)
		if err != nil {
			return err
		}
		mdHeader, mdTrailer := make(metadata.MD), make(metadata.MD)
		for _, f := range srv.opts.after {
			ctx = f(ctx, &mdHeader, &mdTrailer)
		}
		if len(mdHeader) > 0 {
			if err = grpc.SendHeader(ctx, mdHeader); err != nil {
				return err
			}
		}

		msg, err := srv.enc(ctx, out)
		if err != nil {
			return err
		}
		err = s.SendMsg(msg)
		switch {
		case errors.Is(err, io.EOF):
			return nil
		case err != nil:
			return err
		}

		if len(mdTrailer) > 0 {
			if err = grpc.SetTrailer(ctx, mdTrailer); err != nil {
				return err
			}
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		return ctx, err
	}

	return ctx, nil
}

type ServerBiStream[IN, OUT any] struct {
	server[IN, OUT, endpoint.BiStream[IN, OUT]]
}

func NewServerBiStream[RECEIVE proto.Message, IN, OUT any](
	e endpoint.BiStream[IN, OUT],
	dec DecodeFunc[IN],
	enc EncodeFunc[OUT],
	opts ...ServerOption,
) *ServerBiStream[IN, OUT] {
	var receive RECEIVE
	s := &ServerBiStream[IN, OUT]{
		server: server[IN, OUT, endpoint.BiStream[IN, OUT]]{
			e:   e,
			dec: dec,
			enc: enc,
			reflectReceive: reflect.New(
				reflect.TypeOf(receive).Elem(),
			),
		},
	}
	for _, opt := range opts {
		opt.apply(&s.opts)
	}
	return s
}

func (srv ServerBiStream[IN, OUT]) ServeBiStream(s grpc.ServerStream) (ctx context.Context, err error) {
	ctx = s.Context()

	if len(srv.opts.finalizer) > 0 {
		defer func() {
			for _, f := range srv.opts.finalizer {
				f(ctx, err)
			}
		}()
	}
	if len(srv.opts.errHandlers) > 0 {
		defer func() {
			if err != nil {
				for _, h := range srv.opts.errHandlers {
					h.Handle(ctx, err)
				}
			}
		}()
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	for _, f := range srv.opts.before {
		ctx = f(ctx, md)
	}

	inCh := make(chan IN)
	receive, err := srv.e(ctx, inCh)
	if err != nil {
		return ctx, err
	}

	mdHeader, mdTrailer := make(metadata.MD), make(metadata.MD)
	for _, f := range srv.opts.after {
		ctx = f(ctx, &mdHeader, &mdTrailer)
	}

	doneCh := make(chan struct{})
	group := errgroup.Group{}
	group.Go(func() error {
		defer close(inCh)
		for {
			msg := srv.reflectReceive.Interface().(proto.Message)
			err := s.RecvMsg(msg)
			switch {
			case errors.Is(err, io.EOF):
				return nil
			case err != nil:
				return err
			}
			in, err := srv.dec(ctx, msg)
			if err != nil {
				return err
			}
			select {
			case <-doneCh:
				return nil
			case inCh <- in:
			}
		}
	})
	group.Go(func() error {
		defer close(doneCh)
		if len(mdHeader) > 0 {
			if err = grpc.SendHeader(ctx, mdHeader); err != nil {
				return err
			}
		}
		for {
			out, err := receive()
			switch {
			case errors.Is(err, endpoint.StreamDone):
				return nil
			case err != nil:
				return err
			}
			msg, err := srv.enc(ctx, out)
			if err != nil {
				return err
			}
			err = s.SendMsg(msg)
			switch {
			case errors.Is(err, io.EOF):
				return nil
			case err != nil:
				return err
			}
		}
	})
	if err := group.Wait(); err != nil {
		return ctx, err
	}

	if len(mdTrailer) > 0 {
		if err = grpc.SetTrailer(ctx, mdTrailer); err != nil {
			return ctx, err
		}
	}

	return ctx, nil
}

type serverOptions struct {
	before      []kitgrpc.ServerRequestFunc
	after       []kitgrpc.ServerResponseFunc
	finalizer   []kitgrpc.ServerFinalizerFunc
	errHandlers []transport.ErrorHandler
}

type server[IN, OUT any, E endpoint.Endpoint[IN, OUT]] struct {
	e   E
	dec DecodeFunc[IN]
	enc EncodeFunc[OUT]

	opts serverOptions

	reflectReceive reflect.Value
}
