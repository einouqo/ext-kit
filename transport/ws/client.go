package ws

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"sync"
	"syscall"
	"time"

	"github.com/einouqo/ext-kit/transport"
	"github.com/fasthttp/websocket"
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
		dialer := &dialler{
			ws: &websocket.Dialer{
				Proxy:            http.ProxyFromEnvironment,
				HandshakeTimeout: 45 * time.Second,
			},
		}
		for _, f := range c.opts.before {
			ctx = f(ctx, dialer, headers)
		}

		wsc, resp, err := dialer.ws.DialContext(ctx, c.url.String(), headers)
		if err != nil {
			return nil, err
		}
		defer func() { _ = resp.Body.Close() }()

		conn := &conn{
			ws:     wsc,
			config: c.opts.connection.config,
		}

		for _, f := range c.opts.after {
			ctx = f(ctx, resp, conn)
		}

		ctx, cancel := context.WithCancel(ctx)

		once := sync.Once{}
		leave := func(err error) {
			once.Do(func() {
				code, msg, deadline := c.closure(ctx, err)
				data := websocket.FormatCloseMessage(code.fastsocket(), msg)
				_ = conn.WriteControl(websocket.CloseMessage, data, deadline)
			})
		}

		group := errgroup.Group{}
		group.Go(func() (err error) {
			defer func() { leave(err) }()
			defer cancel()
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
					case
						errors.Is(err, net.ErrClosed),
						errors.Is(err, syscall.EPIPE), // broken pipe can appear on closed underlying TCP connection by peer
						errors.Is(err, websocket.ErrCloseSent):
						return nil
					case err != nil:
						return err
					}
				}
			}
		})
		ins := make(chan IN)
		group.Go(func() (err error) {
			if c.opts.finalizer != nil {
				defer func() {
					for _, f := range c.opts.finalizer {
						f(ctx, err)
					}
				}()
			}
			defer func() { leave(err) }()
			defer close(ins)
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
				select {
				case <-ctx.Done():
					return nil
				default:
					// context is not done, continue
				}
				select {
				case <-ctx.Done():
					return nil
				case ins <- in:
				}
			}
		})

		if c.opts.heartbeat.enable {
			group.Go(func() (err error) {
				defer func() { leave(err) }()
				return heartbeat(ctx, c.opts.heartbeat.config, conn)
			})
		}

		errs := make(chan error)
		go func() {
			defer func() { _ = conn.Close() }()
			defer close(errs)
			err := group.Wait()
			if err != nil {
				c.opts.errorHandlers.Handle(ctx, err)
				errs <- err
			}
		}()

		return func() (in IN, err error) {
			if in, ok := <-ins; ok {
				return in, nil
			}
			if err, ok := <-errs; ok {
				return in, err
			}
			return in, endpoint.StreamDone
		}, nil
	}
}

type clientOptions struct {
	before        []DiallerFunc
	after         []ClientTunerFunc
	finalizer     []kithttp.ClientFinalizerFunc
	errorHandlers transport.ErrorHandlers

	connection struct {
		config connConfig
	}

	heartbeat struct {
		enable bool
		config heartbeatConfig
	}
}
