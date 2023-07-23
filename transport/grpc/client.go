package grpc

import (
	"context"
	"errors"
	"io"
	"reflect"

	"github.com/go-kit/kit/transport"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/hashicorp/go-multierror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	"github.com/einouqo/ext-kit/endpoint"
)

type ClientUnary[OUT, IN any] struct{ client[OUT, IN] }

func NewClientUnary[REPLY proto.Message, OUT, IN any](
	cc *grpc.ClientConn,
	fullMethod string,
	enc EncodeFunc[OUT],
	dec DecodeFunc[IN],
	opts ...ClientOption,
) *ClientUnary[OUT, IN] {
	var reply REPLY
	c := &ClientUnary[OUT, IN]{
		client[OUT, IN]{
			conn:   cc,
			method: fullMethod,
			enc:    enc,
			dec:    dec,
			reflectReply: reflect.New(
				reflect.TypeOf(reply).Elem(),
			),
		},
	}
	for _, opt := range opts {
		opt.apply(&c.opts)
	}
	return c
}

func (c *ClientUnary[OUT, IN]) Endpoint() endpoint.Unary[OUT, IN] {
	return func(ctx context.Context, out OUT) (in IN, err error) {
		if len(c.opts.errHandlers) > 0 {
			defer func() {
				if err != nil {
					for _, h := range c.opts.errHandlers {
						h.Handle(ctx, err)
					}
				}
			}()
		}
		if len(c.opts.finalizer) > 0 {
			defer func() {
				for _, f := range c.opts.finalizer {
					f(ctx, err)
				}
			}()
		}

		md := &metadata.MD{}
		for _, f := range c.opts.before {
			ctx = f(ctx, md)
		}
		ctx = metadata.NewOutgoingContext(ctx, *md)

		req, err := c.enc(ctx, out)
		if err != nil {
			return in, err
		}

		var header, trailer metadata.MD
		opts := append(
			c.opts.callOpts,
			grpc.Header(&header),
			grpc.Trailer(&trailer),
		)
		reply := c.reflectReply.Interface()
		if err := c.conn.Invoke(ctx, c.method, req, reply, opts...); err != nil {
			return in, err
		}

		for _, f := range c.opts.after {
			ctx = f(ctx, header, trailer)
		}

		in, err = c.dec(ctx, reply.(proto.Message))
		if err != nil {
			return in, err
		}

		return in, nil
	}
}

type ClientInnerStream[OUT, IN any] struct{ client[OUT, IN] }

func NewClientInnerStream[REPLY proto.Message, OUT, IN any](
	cc *grpc.ClientConn,
	fullMethod string,
	enc EncodeFunc[OUT],
	dec DecodeFunc[IN],
	opts ...ClientOption,
) *ClientInnerStream[OUT, IN] {
	var reply REPLY
	c := &ClientInnerStream[OUT, IN]{
		client: client[OUT, IN]{
			conn:   cc,
			desc:   &grpc.StreamDesc{ServerStreams: true},
			method: fullMethod,
			enc:    enc,
			dec:    dec,
			reflectReply: reflect.New(
				reflect.TypeOf(reply).Elem(),
			),
		},
	}
	for _, opt := range opts {
		opt.apply(&c.opts)
	}
	return c
}

func (c *ClientInnerStream[OUT, IN]) Endpoint() endpoint.InnerStream[OUT, IN] {
	return func(ctx context.Context, out OUT) (rcv endpoint.Receive[IN], stop endpoint.Stop, err error) {
		ctx, cancel := context.WithCancel(ctx)
		defer func() {
			if err != nil {
				cancel()
			}
		}()

		md := &metadata.MD{}
		for _, f := range c.opts.before {
			ctx = f(ctx, md)
		}
		ctx = metadata.NewOutgoingContext(ctx, *md)

		var header, trailer metadata.MD
		opts := append(
			c.opts.callOpts,
			grpc.Header(&header),
			grpc.Trailer(&trailer),
		)
		stream, err := c.conn.NewStream(ctx, c.desc, c.method, opts...)
		if err != nil {
			return nil, nil, err
		}

		for _, f := range c.opts.after {
			ctx = f(ctx, header, trailer)
		}

		msg, err := c.enc(ctx, out)
		if err != nil {
			return nil, nil, err
		}

		if err := stream.SendMsg(msg); err != nil {
			return nil, nil, err
		}
		if err := stream.CloseSend(); err != nil {
			return nil, nil, err
		}

		inC := make(chan IN)
		group := multierror.Group{}
		group.Go(func() (err error) {
			if c.opts.finalizer != nil {
				defer func() {
					for _, f := range c.opts.finalizer {
						f(ctx, err)
					}
				}()
			}
			defer close(inC)
			defer cancel()
			for {
				msg := c.reflectReply.Interface().(proto.Message)
				err = stream.RecvMsg(msg)
				switch {
				case errors.Is(err, io.EOF):
					return nil
				case err != nil:
					return err
				}
				in, err := c.dec(ctx, msg)
				if err != nil {
					return err
				}
				inC <- in
			}
		})

		errC := make(chan error)
		go func() {
			defer close(errC)
			err := group.Wait().ErrorOrNil()
			if err != nil {
				for _, h := range c.opts.errHandlers {
					h.Handle(ctx, err)
				}
				errC <- err
			}
		}()

		return func() (in IN, err error) {
			if in, ok := <-inC; ok {
				return in, nil
			}
			if err, ok := <-errC; ok {
				return in, err
			}
			return in, endpoint.StreamDone
		}, cancel, nil
	}
}

type ClientOuterStream[OUT, IN any] struct{ client[OUT, IN] }

func NewClientOuterStream[REPLY proto.Message, OUT, IN any](
	cc *grpc.ClientConn,
	fullMethod string,
	enc EncodeFunc[OUT],
	dec DecodeFunc[IN],
	opts ...ClientOption,
) *ClientOuterStream[OUT, IN] {
	var reply REPLY
	c := &ClientOuterStream[OUT, IN]{
		client: client[OUT, IN]{
			conn:   cc,
			desc:   &grpc.StreamDesc{ClientStreams: true},
			method: fullMethod,
			enc:    enc,
			dec:    dec,
			reflectReply: reflect.New(
				reflect.TypeOf(reply).Elem(),
			),
		},
	}
	for _, opt := range opts {
		opt.apply(&c.opts)
	}
	return c
}

func (c *ClientOuterStream[OUT, IN]) Endpoint() endpoint.OuterStream[OUT, IN] {
	return func(ctx context.Context, receiver <-chan OUT) (in IN, err error) {
		if len(c.opts.errHandlers) > 0 {
			defer func() {
				if err != nil {
					for _, h := range c.opts.errHandlers {
						h.Handle(ctx, err)
					}
				}
			}()
		}
		if len(c.opts.finalizer) > 0 {
			defer func() {
				for _, f := range c.opts.finalizer {
					f(ctx, err)
				}
			}()
		}

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		md := &metadata.MD{}
		for _, f := range c.opts.before {
			ctx = f(ctx, md)
		}
		ctx = metadata.NewOutgoingContext(ctx, *md)

		var header, trailer metadata.MD
		opts := append(
			c.opts.callOpts,
			grpc.Header(&header),
			grpc.Trailer(&trailer),
		)
		stream, err := c.conn.NewStream(ctx, c.desc, c.method, opts...)
		if err != nil {
			return in, err
		}

		group := multierror.Group{}
		group.Go(func() (err error) {
			defer func() {
				err = multierror.Append(
					err,
					stream.CloseSend(),
				).ErrorOrNil()
			}()
			for out := range receiver {
				msg, err := c.enc(ctx, out)
				if err != nil {
					return err
				}
				err = stream.SendMsg(msg)
				switch {
				case errors.Is(err, io.EOF):
					return nil
				case err != nil:
					return err
				}
			}
			return nil
		})
		inC := make(chan IN, 1)
		group.Go(func() error {
			defer close(inC)
			defer cancel()
			msg := c.reflectReply.Interface().(proto.Message)
			err := stream.RecvMsg(msg)
			switch {
			case errors.Is(err, io.EOF):
				return io.ErrUnexpectedEOF
			case err != nil:
				return err
			}
			in, err := c.dec(ctx, msg)
			if err != nil {
				return err
			}
			inC <- in
			return nil
		})
		if err := group.Wait().ErrorOrNil(); err != nil {
			return in, err
		}

		for _, f := range c.opts.after {
			ctx = f(ctx, header, trailer)
		}

		return <-inC, nil
	}
}

type ClientBiStream[OUT, IN any] struct{ client[OUT, IN] }

func NewClientBiStream[REPLY proto.Message, OUT, IN any](
	cc *grpc.ClientConn,
	fullMethod string,
	enc EncodeFunc[OUT],
	dec DecodeFunc[IN],
	opts ...ClientOption,
) *ClientBiStream[OUT, IN] {
	var reply REPLY
	c := &ClientBiStream[OUT, IN]{
		client: client[OUT, IN]{
			conn:   cc,
			desc:   &grpc.StreamDesc{ClientStreams: true, ServerStreams: true},
			method: fullMethod,
			enc:    enc,
			dec:    dec,
			reflectReply: reflect.New(
				reflect.TypeOf(reply).Elem(),
			),
		},
	}
	for _, opt := range opts {
		opt.apply(&c.opts)
	}
	return c
}

func (c *ClientBiStream[OUT, IN]) Endpoint() endpoint.BiStream[OUT, IN] {
	return func(ctx context.Context, receiver <-chan OUT) (rcv endpoint.Receive[IN], stop endpoint.Stop, err error) {
		ctx, cancel := context.WithCancel(ctx)
		defer func() {
			if err != nil {
				cancel()
			}
		}()

		md := &metadata.MD{}
		for _, f := range c.opts.before {
			ctx = f(ctx, md)
		}
		ctx = metadata.NewOutgoingContext(ctx, *md)

		var header, trailer metadata.MD
		opts := append(
			c.opts.callOpts,
			grpc.Header(&header),
			grpc.Trailer(&trailer),
		)
		stream, err := c.conn.NewStream(ctx, c.desc, c.method, opts...)
		if err != nil {
			return nil, nil, err
		}

		for _, f := range c.opts.after {
			ctx = f(ctx, header, trailer)
		}

		group := multierror.Group{}
		group.Go(func() (err error) {
			defer func() {
				err = multierror.Append(
					err,
					stream.CloseSend(),
				).ErrorOrNil()
			}()
			for out := range receiver {
				msg, err := c.enc(ctx, out)
				if err != nil {
					return err
				}
				err = stream.SendMsg(msg)
				switch {
				case errors.Is(err, io.EOF):
					return nil
				case err != nil:
					return err
				}
			}
			return nil
		})
		inC := make(chan IN)
		group.Go(func() error {
			if c.opts.finalizer != nil {
				defer func() {
					for _, f := range c.opts.finalizer {
						f(ctx, err)
					}
				}()
			}
			defer close(inC)
			defer cancel()
			for {
				msg := c.reflectReply.Interface().(proto.Message)
				err := stream.RecvMsg(msg)
				switch {
				case errors.Is(err, io.EOF):
					return nil
				case err != nil:
					return err
				}
				in, err := c.dec(ctx, msg)
				if err != nil {
					return err // TODO: should we do something with the a stream if an error here?
				}
				inC <- in
			}
		})

		errC := make(chan error)
		go func() {
			defer close(errC)
			err := group.Wait().ErrorOrNil()
			if err != nil {
				for _, h := range c.opts.errHandlers {
					h.Handle(ctx, err)
				}
				errC <- err
			}
		}()

		return func() (in IN, err error) {
			if in, ok := <-inC; ok {
				return in, nil
			}
			if err, ok := <-errC; ok {
				return in, err
			}
			return in, endpoint.StreamDone
		}, cancel, nil
	}
}

type clientOptions struct {
	callOpts []grpc.CallOption

	before      []kitgrpc.ClientRequestFunc
	after       []kitgrpc.ClientResponseFunc
	finalizer   []kitgrpc.ClientFinalizerFunc
	errHandlers []transport.ErrorHandler
}

type client[OUT, IN any] struct {
	conn   *grpc.ClientConn
	desc   *grpc.StreamDesc
	method string

	enc EncodeFunc[OUT]
	dec DecodeFunc[IN]

	opts clientOptions

	reflectReply reflect.Value
}
