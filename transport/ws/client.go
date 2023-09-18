package ws

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"sync"
	"syscall"

	"github.com/fasthttp/websocket"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	"golang.org/x/sync/errgroup"

	"github.com/einouqo/ext-kit/endpoint"
)

type Client[OUT, IN any] struct {
	url url.URL

	enc     EncodeFunc[OUT]
	dec     DecodeFunc[IN]
	closure Closure

	opts clientOptions
}

func NewClient[OUT, IN any](
	url url.URL,
	enc EncodeFunc[OUT],
	dec DecodeFunc[IN],
	closure Closure,
	opts ...ClientOption,
) *Client[OUT, IN] {
	c := &Client[OUT, IN]{
		url:     url,
		enc:     enc,
		dec:     dec,
		closure: closure,
	}
	for _, opt := range opts {
		opt.apply(&c.opts)
	}
	return c
}

func (c *Client[OUT, IN]) Endpoint() endpoint.BiStream[OUT, IN] {
	return func(ctx context.Context, receiver <-chan OUT) (rcv endpoint.Receive[IN], err error) {
		headers := make(http.Header)
		dialer := dialler{websocket.DefaultDialer}
		for _, f := range c.opts.before {
			ctx = f(ctx, dialer, headers)
		}

		wsc, resp, err := dialer.DialContext(ctx, c.url.String(), headers)
		if err != nil {
			return nil, err
		}
		defer func() {
			if err != nil {
				_ = wsc.Close()
			}
		}()
		defer resp.Body.Close()

		for _, f := range c.opts.after {
			ctx = f(ctx, resp, wsc)
		}

		conn := enhConn{wsc, c.opts.enhancement.config}

		once := sync.Once{}
		doneCh := make(chan struct{})
		group := errgroup.Group{}
		group.Go(func() (err error) {
			defer close(doneCh)
			defer once.Do(func() {
				code, msg, deadline := c.closure(ctx, err)
				data := websocket.FormatCloseMessage(code.fastsocket(), msg)
				_ = conn.WriteControl(websocket.CloseMessage, data, deadline)
			})
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case out, ok := <-receiver:
					if !ok {
						return nil
					}
					msg, mt, err := c.enc(ctx, out)
					if err != nil {
						return err
					}
					err = conn.WriteMessage(mt.fastsocket(), msg)
					switch {
					case errors.Is(err, net.ErrClosed):
						return nil
					case errors.Is(err, syscall.EPIPE): // broken pipe can appear on closed underlying tcp connection by peer
						return nil
					case errors.Is(err, websocket.ErrCloseSent):
						return nil
					case err != nil:
						return err
					}
				}
			}
		})
		inCh := make(chan IN)
		group.Go(func() (err error) {
			if c.opts.finalizer != nil {
				defer func() {
					for _, f := range c.opts.finalizer {
						f(ctx, err)
					}
				}()
			}
			defer close(inCh)
			defer conn.Close()
			defer once.Do(func() {
				code, msg, deadline := c.closure(ctx, err)
				data := websocket.FormatCloseMessage(code.fastsocket(), msg)
				_ = conn.WriteControl(websocket.CloseMessage, data, deadline)
			})
			for {
				messageType, msg, err := conn.ReadMessage()
				switch {
				case websocket.IsCloseError(err, websocket.CloseNormalClosure):
					return nil
				case err != nil:
					return err
				}
				mt := MessageType(messageType)
				in, err := c.dec(ctx, mt, msg)
				if err != nil {
					return err
				}
				inCh <- in
			}
		})

		if c.opts.heartbeat.enable {
			group.Go(func() error {
				defer conn.Close()
				return heartbeat(ctx, c.opts.heartbeat.config, conn, doneCh)
			})
		}

		errCh := make(chan error)
		go func() {
			defer close(errCh)
			err := group.Wait()
			if err != nil {
				for _, h := range c.opts.errHandlers {
					h.Handle(ctx, err)
				}
				errCh <- err
			}
		}()

		return func() (in IN, err error) {
			if in, ok := <-inCh; ok {
				return in, nil
			}
			if err, ok := <-errCh; ok {
				return in, err
			}
			return in, endpoint.StreamDone
		}, nil
	}
}

type clientOptions struct {
	before      []DiallerFunc
	after       []ClientTunerFunc
	finalizer   []kithttp.ClientFinalizerFunc
	errHandlers []transport.ErrorHandler

	enhancement struct {
		config enhConfig
	}

	heartbeat struct {
		enable bool
		config hbConfig
	}
}
