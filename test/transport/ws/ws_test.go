//go:build integration
// +build integration

package ws

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"golang.org/x/exp/slices"

	"github.com/einouqo/ext-kit/endpoint"
	"github.com/einouqo/ext-kit/test/transport/_service"
	"github.com/einouqo/ext-kit/transport/ws"
)

const (
	address string = ":8811"
)

type errNoErrHandler struct {
	t *testing.T
}

func (h errNoErrHandler) Handle(_ context.Context, err error) {
	if err != nil {
		h.t.Fatalf("handle error: expect no error but got %q", err)
	}
}

func TestStreamWS_ok(t *testing.T) {
	srvDone := atomic.Value{}
	srvDone.Store(false)
	tidy, err := prepareServer(
		address,
		ws.WithServerErrorHandler(errNoErrHandler{t}),
		ws.WithServerFinalizer(func(ctx context.Context, code int, r *http.Request) {
			srvDone.Store(true)
		}),
	)
	if err != nil {
		t.Fatalf("unable to prepare server: %+v", err)
	}
	defer tidy()

	cliDone := atomic.Value{}
	cliDone.Store(false)
	client := prepareClient(
		address,
		ws.WithClientFinalizer(func(ctx context.Context, err error) {
			cliDone.Store(true)
		}),
	)

	ctx := context.Background()

	n := 10
	nn := n * (n - 1) / 2 // expect triangular like number
	mu := sync.Mutex{}
	msgs := make([]string, 0, nn)
	sendC := make(chan service.EchoRequest)
	go func() {
		defer close(sendC)
		for i := 0; i < n; i++ {
			req := service.EchoRequest{
				Message: fmt.Sprintf("bi %d", i),
				Repeat:  uint(i),
			}
			mu.Lock()
			sendC <- req
			for j := 0; j < int(req.Repeat); j++ {
				msgs = append(msgs, req.Message)
			}
			mu.Unlock()
		}
		time.Sleep(50 * time.Millisecond) // wait for server to process all messages
	}()
	receive, err := client.Stream(ctx, sendC)
	if err != nil {
		t.Fatalf("call error: %+v", err)
	}

	i := 0
	for {
		msg, err := receive()
		switch {
		case errors.Is(err, endpoint.StreamDone):
			if want, have := nn, i; want != have {
				t.Fatalf("iteration: want %d, have %d", want, have)
			}
			if len(msgs) != 0 {
				t.Fatalf("message: want empty, have %q", msgs)
			}
			if !cliDone.Load().(bool) {
				t.Fatal("client didn't finish")
			}
			dur := time.Second
			wait := time.After(dur)
			for {
				select {
				case <-wait:
					t.Fatalf("server didn't finish for %s", dur)
				default:
					if srvDone.Load().(bool) {
						return
					}
					time.Sleep(time.Millisecond)
				}
			}
			return
		case err != nil:
			t.Fatalf("request: %+v", err)
		}
		if i > nn {
			t.Fatalf("iteration: want less than %d, have %d", nn, i)
		}
		if len(msg.Messages) != 1 {
			t.Fatalf("message: want exactly 1 message, have %d", len(msg.Messages))
		}
		mu.Lock()
		if !slices.Contains(msgs, msg.Messages[0]) {
			t.Fatalf("message: want %q to contain in %q", msg.Messages[0], msgs)
		}
		once := true
		msgs = slices.DeleteFunc(msgs, func(s string) bool {
			del := s == msg.Messages[0]
			if del && once {
				defer func() { once = false }()
			}
			return del && once
		})
		mu.Unlock()
		i++
	}
}

func TestStreamWS_error(t *testing.T) {
	srvDone := atomic.Value{}
	srvDone.Store(false)
	tidy, err := prepareServer(
		address,
		ws.WithServerFinalizer(func(ctx context.Context, code int, r *http.Request) {
			srvDone.Store(true)
		}),
	)
	if err != nil {
		t.Fatalf("unable to prepare server: %+v", err)
	}
	defer tidy()

	cliDone := atomic.Value{}
	cliDone.Store(false)
	client := prepareClient(
		address,
		ws.WithClientFinalizer(func(ctx context.Context, err error) {
			cliDone.Store(true)
		}),
	)

	ctx := context.Background()

	n := 10
	nn := n * (n - 1) / 2 // expect triangular like number
	mu := sync.Mutex{}
	msgs := make([]string, 0, nn)
	errIter := n / 2
	errMsg := ""
	sendC := make(chan service.EchoRequest)
	go func() {
		defer close(sendC)
		for i := 0; i < n; i++ {
			req := service.EchoRequest{
				Message: fmt.Sprintf("bi %d", i),
			}
			if i == errIter {
				req.IsError = true
				errMsg = req.Message
			} else {
				req.Repeat = uint(i)
				mu.Lock()
				for j := 0; j < int(req.Repeat); j++ {
					msgs = append(msgs, req.Message)
				}
				mu.Unlock()
			}
			sendC <- req
		}
		time.Sleep(50 * time.Millisecond) // wait for server to process all messages
	}()

	receive, err := client.Stream(ctx, sendC)
	if err != nil {
		t.Fatalf("call error: %+v", err)
	}

	i := 0
	for {
		msg, err := receive()
		switch {
		case errors.Is(err, endpoint.StreamDone):
			t.Fatal("want error, have stream done")
		case err != nil:
			wsErr, ok := err.(*websocket.CloseError)
			if !ok {
				t.Fatalf("error type: want %T, have %T (with %q)", &websocket.CloseError{}, err, err)
			}
			code, msg, _ := sCloser(ctx, errors.New(errMsg))
			target := &websocket.CloseError{
				Code: int(code),
				Text: msg,
			}
			if wsErr.Code != target.Code {
				t.Fatalf("error code: want %d, have %d", target.Code, wsErr.Code)
			}
			if wsErr.Text != target.Text {
				t.Fatalf("error message: want %q, have %q", target.Text, wsErr.Text)
			}
			if !cliDone.Load().(bool) {
				t.Fatal("client didn't finish")
			}
			dur := time.Second
			wait := time.After(dur)
			for {
				select {
				case <-wait:
					t.Fatalf("server didn't finish for %s", dur)
				default:
					if srvDone.Load().(bool) {
						return
					}
					time.Sleep(time.Millisecond)
				}
			}
			return
		}
		if i > nn {
			t.Fatalf("iterations: expect less than %d, have %d", nn, i)
		}
		if len(msg.Messages) != 1 {
			t.Fatalf("message: want exactly 1 message, have %d", len(msg.Messages))
		}
		mu.Lock()
		if !slices.Contains(msgs, msg.Messages[0]) {
			t.Fatalf("message: want %q to contain in %q", msg.Messages[0], msgs)
		}
		once := true
		msgs = slices.DeleteFunc(msgs, func(s string) bool {
			del := s == msg.Messages[0]
			if del && once {
				defer func() { once = false }()
			}
			return del && once
		})
		mu.Unlock()
		i++
	}
}

func TestStreamWS_cancel(t *testing.T) {
	srvDone := atomic.Value{}
	srvDone.Store(false)
	tidy, err := prepareServer(
		address,
		ws.WithServerFinalizer(func(ctx context.Context, code int, r *http.Request) {
			srvDone.Store(true)
		}),
	)
	if err != nil {
		t.Fatalf("unable to prepare server: %+v", err)
	}
	defer tidy()

	cliDone := atomic.Value{}
	cliDone.Store(false)
	client := prepareClient(
		address,
		ws.WithClientFinalizer(func(ctx context.Context, err error) {
			cliDone.Store(true)
		}),
	)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	n := 10
	nn := n * (n - 1) / 2 // expect triangular like number
	cancelIter := n / 2
	sendC := make(chan service.EchoRequest)
	go func() {
		defer close(sendC)
		for i := 0; i < n; i++ {
			req := service.EchoRequest{
				Message: fmt.Sprintf("bi %d", i),
				Repeat:  uint(i),
				Latency: 10 * time.Millisecond,
			}
			if i == cancelIter {
				cancel()
				time.Sleep(50 * time.Millisecond)
				return
			}
			sendC <- req
		}
	}()

	receive, err := client.Stream(ctx, sendC)
	if err != nil {
		t.Fatalf("call error: %+v", err)
	}

	i := 0
	for {
		msg, err := receive()
		switch {
		case errors.Is(err, endpoint.StreamDone):
			t.Fatal("want error, have stream done")
		case err != nil:
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("want %s to be %s", err, context.Canceled)
			}
			if !cliDone.Load().(bool) {
				t.Fatal("client didn't finish")
			}
			dur := time.Second
			wait := time.After(dur)
			for {
				select {
				case <-wait:
					t.Fatalf("server didn't finish for %s", dur)
				default:
					if srvDone.Load().(bool) {
						return
					}
					time.Sleep(time.Millisecond)
				}
			}
			return
		}
		if i > nn {
			t.Fatalf("iterations: expect less than %d, have %d", nn, i)
		}
		if len(msg.Messages) != 1 {
			t.Fatalf("message: want exactly 1 message, have %d", len(msg.Messages))
		}
		i++
	}
}

func TestStreamWS_heartbeat_client(t *testing.T) {
	sTidy, err := prepareServer(address)
	if err != nil {
		t.Fatalf("unable to prepare server: %+v", err)
	}
	defer sTidy()

	roundMu := sync.Mutex{}
	clientPingRounds := 0
	client := prepareClient(
		address,
		ws.WithClientPing(time.Nanosecond, time.Second, func(context.Context) (msg []byte, deadline time.Time) {
			roundMu.Lock()
			clientPingRounds++
			roundMu.Unlock()
			return []byte("ping"), time.Now().Add(time.Second)
		}),
	)

	ctx := context.Background()
	sendC := make(chan service.EchoRequest)
	defer close(sendC)
	_, err = client.Stream(ctx, sendC)
	if err != nil {
		t.Fatalf("call error: %+v", err)
	}

	dur := 5 * time.Second
	wait := time.After(5 * time.Second)
	for {
		roundMu.Lock()
		rounds := clientPingRounds
		roundMu.Unlock()
		select {
		case <-wait:
			t.Fatalf("not enougth pings for %s (have: %d rounds)", dur, rounds)
		default:
			if min, have := 2, rounds; min < have {
				return
			}
			time.Sleep(time.Millisecond)
		}
	}
}

func TestStreamWS_heartbeat_server(t *testing.T) {
	roundMu := sync.Mutex{}
	serverPingRounds := 0
	sTidy, err := prepareServer(
		address,
		ws.WithServerPing(time.Nanosecond, time.Second, func(context.Context) (msg []byte, deadline time.Time) {
			roundMu.Lock()
			serverPingRounds++
			roundMu.Unlock()
			return []byte("ping"), time.Now().Add(time.Second)
		}),
	)
	if err != nil {
		t.Fatalf("unable to prepare server: %+v", err)
	}
	defer sTidy()

	client := prepareClient(address)

	ctx := context.Background()
	sendC := make(chan service.EchoRequest)
	defer close(sendC)
	_, err = client.Stream(ctx, sendC)
	if err != nil {
		t.Fatalf("call error: %+v", err)
	}

	dur := 5 * time.Second
	wait := time.After(dur)
	for {
		roundMu.Lock()
		rounds := serverPingRounds
		roundMu.Unlock()
		select {
		case <-wait:
			t.Fatalf("not enougth pings for %s (have: %d rounds)", dur, rounds)
		default:
			if min, have := 2, rounds; min < have {
				return
			}
			time.Sleep(time.Millisecond)
		}
	}
}

func TestStreamWS_heartbeat_both(t *testing.T) {
	serverRoundMu := sync.Mutex{}
	serverPingRounds := 0
	sTidy, err := prepareServer(
		address,
		ws.WithServerPing(time.Nanosecond, time.Second, func(context.Context) (msg []byte, deadline time.Time) {
			serverRoundMu.Lock()
			serverPingRounds++
			serverRoundMu.Unlock()
			return []byte("ping"), time.Now().Add(time.Second)
		}),
	)
	if err != nil {
		t.Fatalf("unable to prepare server: %+v", err)
	}
	defer sTidy()

	clientRoundMu := sync.Mutex{}
	clientPingRounds := 0
	client := prepareClient(
		address,
		ws.WithClientPing(time.Nanosecond, time.Second, func(context.Context) (msg []byte, deadline time.Time) {
			clientRoundMu.Lock()
			clientPingRounds++
			clientRoundMu.Unlock()
			return []byte("ping"), time.Now().Add(time.Second)
		}),
	)

	ctx := context.Background()
	sendC := make(chan service.EchoRequest)
	defer close(sendC)
	_, err = client.Stream(ctx, sendC)
	if err != nil {
		t.Fatalf("call error: %+v", err)
	}

	dur := 5 * time.Second
	wait := time.After(5 * time.Second)
	for {
		clientRoundMu.Lock()
		crounds := clientPingRounds
		clientRoundMu.Unlock()
		serverRoundMu.Lock()
		srounds := serverPingRounds
		serverRoundMu.Unlock()
		select {
		case <-wait:
			t.Fatalf("not enougth pings for %s (have: %d client rounds and %d server rounds)", dur, crounds, srounds)
		default:
			if min, chave, shave := 2, crounds, srounds; min < chave && min < shave {
				return
			}
			time.Sleep(time.Millisecond)
		}
	}
}
