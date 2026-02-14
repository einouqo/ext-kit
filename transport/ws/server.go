package ws

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"syscall"

	"github.com/einouqo/ext-kit/transport"
	"github.com/fasthttp/websocket"
	kithttp "github.com/go-kit/kit/transport/http"
	"golang.org/x/sync/errgroup"

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
		s.opts.errorHandlers.Handle(ctx, err)
	}
}

func (s *Server[IN, OUT]) serve(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {
	headers := make(http.Header)
	upg := new(upgrader)
	for _, f := range s.opts.before {
		ctx = f(ctx, upg, r, headers)
	}

	wsc, err := upg.Upgrade(w, r, headers)
	if err != nil {
		return err
	}
	conn := &conn{
		ws:     wsc,
		config: s.opts.connection.config,
	}
	defer func() { _ = conn.Close() }()

	for _, f := range s.opts.after {
		ctx = f(ctx, conn)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ins := make(chan IN)
	receive, err := s.e(ctx, ins)
	if err != nil {
		return err
	}

	once := sync.Once{}
	leave := func(err error) {
		once.Do(func() {
			code, msg, deadline := s.closure(ctx, err)
			data := websocket.FormatCloseMessage(code.fastsocket(), msg)
			_ = conn.WriteControl(websocket.CloseMessage, data, deadline)
		})
	}

	group := errgroup.Group{}
	group.Go(func() (err error) {
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
			in, err := s.dec(ctx, mt, msg)
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
	group.Go(func() (err error) {
		defer func() { leave(err) }()
		defer cancel()
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
	})

	if s.opts.heartbeat.enable {
		group.Go(func() (err error) {
			defer func() { leave(err) }()
			return heartbeat(ctx, s.opts.heartbeat.config, conn)
		})
	}

	return group.Wait()
}

type serverOptions struct {
	before        []UpgradeFunc
	after         []ServerTunerFunc
	finalizer     []kithttp.ServerFinalizerFunc
	errorHandlers transport.ErrorHandlers

	connection struct {
		config connConfig
	}

	heartbeat struct {
		enable bool
		config heartbeatConfig
	}
}
