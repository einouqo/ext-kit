package ws

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"syscall"
	"time"

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
		headers := &http.Header{}
		for _, f := range c.opts.before {
			ctx = f(ctx, headers)
		}

		wsc, resp, err := c.dialer(
			c.opts.enhancement.preset.write.compression.enable,
		).DialContext(ctx, c.url.String(), *headers)
		if err != nil {
			return nil, err
		}
		defer func() {
			if err != nil {
				_ = wsc.Close()
			}
		}()

		for _, f := range c.opts.after {
			ctx = f(ctx, resp)
		}

		conn, err := enhance(c.opts.enhancement.preset, c.opts.enhancement.config, wsc)
		if err != nil {
			return nil, err
		}

		doneCh := make(chan struct{})
		group := errgroup.Group{}
		group.Go(func() (err error) {
			defer close(doneCh)
			defer func() {
				code, msg, deadline := c.closure(ctx, err)
				data := websocket.FormatCloseMessage(code.fastsocket(), msg)
				_ = conn.WriteControl(websocket.CloseMessage, data, deadline)
			}()
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

				pongCh := make(chan struct{})
				handler := conn.PongHandler()
				conn.SetPongHandler(func(msg string) error {
					select {
					case pongCh <- struct{}{}:
					case <-doneCh:
					}
					return handler(msg)
				})

				ticker := time.NewTicker(c.opts.heartbeat.period)
				defer ticker.Stop()
				for {
					select {
					case <-doneCh:
						return nil
					case <-ticker.C:
						msg, deadline := c.opts.heartbeat.pinging(ctx)
						err := conn.WriteControl(websocket.PingMessage, msg, deadline)
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
					select {
					case <-doneCh:
						return nil
					case <-time.After(c.opts.heartbeat.await):
						return context.DeadlineExceeded
					case <-pongCh:
						ticker.Reset(c.opts.heartbeat.period)
					}
				}
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

func (c *Client[OUT, IN]) dialer(compression bool) *websocket.Dialer {
	if c.opts.dialer != nil {
		return c.opts.dialer
	}
	dd := websocket.DefaultDialer
	dd.EnableCompression = compression
	return dd
}

type clientOptions struct {
	dialer *websocket.Dialer

	before      []ClientHeaderFunc
	after       []ClientResponseFunc
	finalizer   []kithttp.ClientFinalizerFunc
	errHandlers []transport.ErrorHandler

	enhancement struct {
		preset enhPreset
		config enhConfig
	}

	heartbeat struct {
		enable        bool
		period, await time.Duration
		pinging       Pinging
	}
}
