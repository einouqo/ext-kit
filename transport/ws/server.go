package ws

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"syscall"

	"github.com/fasthttp/websocket"
	"github.com/go-kit/kit/transport"
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
		for _, h := range s.opts.errHandlers {
			h.Handle(ctx, err)
		}
	}
}

func (s *Server[IN, OUT]) serve(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {
	headers := make(http.Header)
	upg := upgrader{new(websocket.Upgrader)}
	for _, f := range s.opts.before {
		ctx = f(ctx, upg, r, headers)
	}

	wsc, err := upg.Upgrade(w, r, headers)
	if err != nil {
		return err
	}
	defer wsc.Close()

	for _, f := range s.opts.after {
		ctx = f(ctx, wsc)
	}

	conn := enhConn{wsc, s.opts.enhancement.config}

	inCh := make(chan IN)
	receive, err := s.e(ctx, inCh)
	if err != nil {
		return err
	}

	once := sync.Once{}
	doneCh := make(chan struct{})
	group := errgroup.Group{}
	group.Go(func() (err error) {
		defer close(inCh)
		defer conn.Close()
		defer once.Do(func() {
			code, msg, deadline := s.closure(ctx, err)
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
			in, err := s.dec(ctx, mt, msg)
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
	group.Go(func() (err error) {
		defer close(doneCh)
		defer once.Do(func() {
			code, msg, deadline := s.closure(ctx, err)
			data := websocket.FormatCloseMessage(code.fastsocket(), msg)
			_ = conn.WriteControl(websocket.CloseMessage, data, deadline)
		})
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
	})

	if s.opts.heartbeat.enable {
		group.Go(func() error {
			defer conn.Close()
			return heartbeat(ctx, s.opts.heartbeat.config, conn, doneCh)
		})
	}

	if err := group.Wait(); err != nil {
		return err
	}

	return nil
}

type serverOptions struct {
	before      []UpgradeFunc
	after       []ServerTunerFunc
	finalizer   []kithttp.ServerFinalizerFunc
	errHandlers []transport.ErrorHandler

	enhancement struct {
		config enhConfig
	}

	heartbeat struct {
		enable bool
		config hbConfig
	}
}
