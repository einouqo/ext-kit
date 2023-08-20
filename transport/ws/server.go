package ws

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/fasthttp/websocket"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/einouqo/ext-kit/endpoint"
)

type Server[IN, OUT any] struct {
	e endpoint.BiStream[IN, OUT]

	dec     DecodeFunc[IN]
	enc     EncodeFunc[OUT]
	closure Closure

	opts serverOptions
}

func NewServer[IN, OUT any](
	e endpoint.BiStream[IN, OUT],
	dec DecodeFunc[IN],
	enc EncodeFunc[OUT],
	closure Closure,
	opts ...ServerOption,
) *Server[IN, OUT] {
	s := &Server[IN, OUT]{
		e:       e,
		dec:     dec,
		enc:     enc,
		closure: closure,
	}
	for _, opt := range opts {
		opt.apply(&s.opts)
	}
	return s
}

func (s *Server[IN, OUT]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if len(s.opts.finalizer) > 0 {
		iw := &interceptingWriter{w, http.StatusOK, 0}
		defer func() {
			ctx = context.WithValue(ctx, kithttp.ContextKeyResponseHeaders, iw.Header())
			ctx = context.WithValue(ctx, kithttp.ContextKeyResponseSize, iw.written)
			for _, f := range s.opts.finalizer {
				f(ctx, iw.code, r)
			}
		}()
		w = iw.reimplementInterfaces()
	}

	err := s.serve(ctx, w, r)
	if err != nil {
		for _, h := range s.opts.errHandlers {
			h.Handle(ctx, err)
		}
	}
}

func (s *Server[IN, OUT]) serve(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	headers := &http.Header{}
	for _, f := range s.opts.before {
		ctx = f(ctx, r.Header, headers)
	}

	conn, err := s.upgrader().Upgrade(w, r, *headers)
	if err != nil {
		return err
	}
	defer conn.Close()

	for _, f := range s.opts.after {
		ctx = f(ctx, conn)
	}

	inC := make(chan IN)
	receive, stop, err := s.e(ctx, inC)
	if err != nil {
		return err
	}

	pongC := make(chan struct{})
	defer close(pongC)
	handler := conn.PongHandler()
	conn.SetPongHandler(func(msg string) error {
		if s.opts.heartbeat.enable {
			pongC <- struct{}{}
		}
		return handler(msg)
	})

	doneC := make(chan struct{})
	group := errgroup.Group{}
	group.Go(func() (err error) {
		defer close(inC)
		defer func() {
			if err != nil {
				cancel()
			}
		}()
		for {
			if err := s.updateReadDeadline(conn); err != nil {
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
			in, err := s.dec(ctx, mt, msg)
			if err != nil {
				return err
			}
			select {
			case <-doneC:
				return nil
			case inC <- in:
			}
		}
	})
	group.Go(func() (err error) {
		defer close(doneC)
		defer func() {
			code, msg, deadline := s.closure(ctx, err)
			conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(code.fastsocket(), msg),
				deadline,
			)
		}()
		for {
			out, err := receive()
			switch {
			case errors.Is(err, endpoint.StreamDone):
				return nil
			case err != nil:
				return err
			}
			msg, mt, err := s.enc(ctx, out)
			if err != nil {
				return err
			}
			if err := s.updateWriteDeadline(conn); err != nil {
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
	})

	timeoutC := make(chan struct{}, 1)
	defer close(timeoutC)
	closedC := make(chan struct{})
	group.Go(func() error {
		defer stop()
		defer close(closedC)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeoutC:
			return nil
		case <-doneC:
			return nil
		}
	})
	if s.opts.heartbeat.enable {
		group.Go(func() error {
			defer func() { timeoutC <- struct{}{} }()
			ticker := time.NewTicker(s.opts.heartbeat.period)
			defer ticker.Stop()
			for {
				select {
				case <-closedC:
					return nil
				case <-ticker.C:
					msg, deadline := s.opts.heartbeat.pinging(ctx)
					if err := conn.WriteControl(websocket.PingMessage, msg, deadline); err != nil {
						return err
					}
				}
				select {
				case <-closedC:
					return nil
				case <-time.After(s.opts.heartbeat.await):
					return context.DeadlineExceeded
				case <-pongC:
					// nothing
				}
				ticker.Reset(s.opts.heartbeat.period)
			}
		})
	}

	if err = group.Wait(); err != nil {
		return err
	}

	return nil
}

func (s *Server[IN, OUT]) upgrader() *websocket.Upgrader {
	if s.opts.upgrader != nil {
		return s.opts.upgrader
	}
	return &websocket.Upgrader{}
}

func (s *Server[IN, OUT]) updateWriteDeadline(conn *websocket.Conn) error {
	if s.opts.timeout.write > 0 {
		deadline := time.Now().Add(s.opts.timeout.write)
		return conn.SetWriteDeadline(deadline)
	}
	return nil
}

func (s *Server[IN, OUT]) updateReadDeadline(conn *websocket.Conn) error {
	if s.opts.timeout.read > 0 {
		deadline := time.Now().Add(s.opts.timeout.read)
		return conn.SetReadDeadline(deadline)
	}
	return nil
}

type serverOptions struct {
	upgrader *websocket.Upgrader

	before      []ServerRequestFunc
	after       []ServerConnectionFunc
	finalizer   []kithttp.ServerFinalizerFunc
	errHandlers []transport.ErrorHandler

	timeout struct {
		read, write time.Duration
	}

	heartbeat struct {
		enable        bool
		period, await time.Duration
		pinging       Pinging
	}
}
