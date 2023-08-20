package ws

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/fasthttp/websocket"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"

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
	return func(ctx context.Context, receiver <-chan OUT) (rcv endpoint.Receive[IN], stop endpoint.Stop, err error) {
		ctx, cancel := context.WithCancel(ctx)
		defer func() {
			if err != nil {
				cancel()
			}
		}()

		headers := &http.Header{}
		for _, f := range c.opts.before {
			ctx = f(ctx, headers)
		}

		conn, _, err := c.dialer().DialContext(ctx, c.url.String(), *headers)
		if err != nil {
			return nil, nil, err
		}

		for _, f := range c.opts.after {
			f(ctx, conn)
		}

		pongC := make(chan struct{})
		handler := conn.PongHandler()
		conn.SetPongHandler(func(msg string) error {
			if c.opts.heartbeat.enable {
				pongC <- struct{}{}
			}
			return handler(msg)
		})

		doneC := make(chan struct{})
		group := errgroup.Group{}
		group.Go(func() (err error) {
			defer close(doneC)
			defer func() {
				code, msg, deadline := c.closure(ctx, err)
				conn.WriteControl(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(code.fastsocket(), msg),
					deadline,
				)
			}()
			for out := range receiver {
				msg, mt, err := c.enc(ctx, out)
				if err != nil {
					return err
				}
				if err := c.updateWriteDeadline(conn); err != nil {
					return err
				}
				err = conn.WriteMessage(mt.fastsocket(), msg)
				switch {
				case errors.Is(err, net.ErrClosed):
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
			for {
				if err := c.updateReadDeadline(conn); err != nil {
					return err
				}
				messageType, msg, err := conn.ReadMessage()
				switch {
				case errors.Is(err, net.ErrClosed):
					return nil
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
				inC <- in
			}
		})

		timeoutC := make(chan struct{}, 1)
		closedC := make(chan struct{})
		group.Go(func() error {
			defer close(closedC)
			defer conn.Close()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timeoutC:
				return nil
			case <-doneC:
				return nil
			}
		})
		if c.opts.heartbeat.enable {
			group.Go(func() error {
				defer func() { timeoutC <- struct{}{} }()
				ticker := time.NewTicker(c.opts.heartbeat.period)
				defer ticker.Stop()
				for {
					select {
					case <-closedC:
						return nil
					case <-ticker.C:
						msg, deadline := c.opts.heartbeat.pinging(ctx)
						if err := conn.WriteControl(websocket.PingMessage, msg, deadline); err != nil {
							return err
						}
					}
					select {
					case <-closedC:
						return nil
					case <-time.After(c.opts.heartbeat.await):
						return context.DeadlineExceeded
					case <-pongC:
						// nothing
					}
					ticker.Reset(c.opts.heartbeat.period)
				}
			})
		}

		errC := make(chan error)
		go func() {
			defer close(errC)
			defer close(pongC)
			defer close(timeoutC)
			if err := group.Wait(); err != nil {
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

func (c *Client[OUT, IN]) dialer() *websocket.Dialer {
	if c.opts.dialer != nil {
		return c.opts.dialer
	}
	return websocket.DefaultDialer
}

func (c *Client[OUT, IN]) updateWriteDeadline(conn *websocket.Conn) error {
	if c.opts.timeout.write > 0 {
		deadline := time.Now().Add(c.opts.timeout.write)
		return conn.SetWriteDeadline(deadline)
	}
	return nil
}

func (c *Client[OUT, IN]) updateReadDeadline(conn *websocket.Conn) error {
	if c.opts.timeout.read > 0 {
		deadline := time.Now().Add(c.opts.timeout.read)
		return conn.SetReadDeadline(deadline)
	}
	return nil
}

type clientOptions struct {
	dialer *websocket.Dialer

	before      []ClientRequestFunc
	after       []ClientConnectionFunc
	finalizer   []kithttp.ClientFinalizerFunc
	errHandlers []transport.ErrorHandler

	timeout struct {
		write, read time.Duration
	}

	heartbeat struct {
		enable        bool
		period, await time.Duration
		pinging       Pinging
	}
}
