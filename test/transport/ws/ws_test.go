package ws

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"golang.org/x/exp/slices"

	"github.com/einouqo/ext-kit/endpoint"
	"github.com/einouqo/ext-kit/test/service"
	"github.com/einouqo/ext-kit/transport/ws"
)

const (
	address string = ":8811"
)

func TestStreamWS_ok(t *testing.T) {
	client, tidy, err := prepareTest(address)
	if err != nil {
		t.Fatalf("unable to prepare test: %+v", err)
	}
	defer tidy()

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
		time.Sleep(100 * time.Millisecond) // wait for server to process all messages
	}()
	receive, stop, err := client.Stream(ctx, sendC)
	if err != nil {
		t.Fatalf("call error: %+v", err)
	}
	defer stop()

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
	client, tidy, err := prepareTest(address)
	if err != nil {
		t.Fatalf("unable to prepare test: %+v", err)
	}
	defer tidy()

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
		time.Sleep(100 * time.Millisecond) // wait for server to process all messages
	}()

	receive, stop, err := client.Stream(ctx, sendC)
	if err != nil {
		t.Fatalf("call error: %+v", err)
	}
	defer stop()

	i := 0
	for {
		msg, err := receive()
		switch {
		case errors.Is(err, endpoint.StreamDone):
			t.Fatal("want error, have stream done")
		case err != nil:
			wsErr, ok := err.(*websocket.CloseError)
			if !ok {
				t.Fatalf("error type: want %T, have %T", &websocket.CloseError{}, err)
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

func TestStreamWS_stop(t *testing.T) {
	client, tidy, err := prepareTest(address)
	if err != nil {
		t.Fatalf("unable to prepare test: %+v", err)
	}
	defer tidy()

	ctx := context.Background()

	n := 10
	nn := n * (n - 1) / 2 // expect triangular like number
	cancelIter := n / 2
	sendC := make(chan service.EchoRequest)

	receive, stop, err := client.Stream(ctx, sendC)
	if err != nil {
		t.Fatalf("call error: %+v", err)
	}
	defer stop()

	go func() {
		defer close(sendC)
		for i := 0; i < n; i++ {
			req := service.EchoRequest{
				Message: fmt.Sprintf("bi %d", i),
				Repeat:  uint(i),
				Latency: 10 * time.Millisecond,
			}
			if i == cancelIter {
				stop()
				time.Sleep(100 * time.Millisecond)
				return
			}
			sendC <- req
		}
	}()

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

func TestStreamWS_cancel(t *testing.T) {
	client, tidy, err := prepareTest(address)
	if err != nil {
		t.Fatalf("unable to prepare test: %+v", err)
	}
	defer tidy()

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
				time.Sleep(100 * time.Millisecond)
				return
			}
			sendC <- req
		}
	}()

	receive, stop, err := client.Stream(ctx, sendC)
	if err != nil {
		t.Fatalf("call error: %+v", err)
	}
	defer stop()

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

	clientPingRounds := 0
	client := prepareClient(
		address,
		ws.WithClientPing(50*time.Millisecond, 500*time.Millisecond, func(context.Context) (msg []byte, deadline time.Time) {
			clientPingRounds++
			return []byte("ping"), time.Now().Add(time.Second)
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
		time.Sleep(500 * time.Millisecond) // wait for server to process all messages
	}()
	receive, stop, err := client.Stream(ctx, sendC)
	if err != nil {
		t.Fatalf("call error: %+v", err)
	}
	defer stop()

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
			if min, have := 2, clientPingRounds; min > have {
				t.Fatalf("client ping rounds: want at least %d, have %d", min, have)
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

func TestStreamWS_heartbeat_server(t *testing.T) {
	serverPingRounds := 0
	sTidy, err := prepareServer(
		address,
		ws.WithServerPing(50*time.Millisecond, 500*time.Millisecond, func(context.Context) (msg []byte, deadline time.Time) {
			serverPingRounds++
			return []byte("ping"), time.Now().Add(time.Second)
		}),
	)
	if err != nil {
		t.Fatalf("unable to prepare server: %+v", err)
	}
	defer sTidy()

	client := prepareClient(address)

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
		time.Sleep(500 * time.Millisecond) // wait for server to process all messages
	}()
	receive, stop, err := client.Stream(ctx, sendC)
	if err != nil {
		t.Fatalf("call error: %+v", err)
	}
	defer stop()

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
			if min, have := 2, serverPingRounds; min > have {
				t.Fatalf("server ping rounds: want at least %d, have %d", min, have)
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

func TestStreamWS_heartbeat_both(t *testing.T) {
	serverPingRounds := 0
	sTidy, err := prepareServer(
		address,
		ws.WithServerPing(50*time.Millisecond, 500*time.Millisecond, func(context.Context) (msg []byte, deadline time.Time) {
			serverPingRounds++
			return []byte("ping"), time.Now().Add(time.Second)
		}),
	)
	if err != nil {
		t.Fatalf("unable to prepare server: %+v", err)
	}
	defer sTidy()

	clientPingRounds := 0
	client := prepareClient(
		address,
		ws.WithClientPing(50*time.Millisecond, 500*time.Millisecond, func(context.Context) (msg []byte, deadline time.Time) {
			clientPingRounds++
			return []byte("ping"), time.Now().Add(time.Second)
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
		time.Sleep(time.Second) // wait for server to process all messages
	}()
	receive, stop, err := client.Stream(ctx, sendC)
	if err != nil {
		t.Fatalf("call error: %+v", err)
	}
	defer stop()

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
			if min, have := 2, serverPingRounds; min > have {
				t.Fatalf("server ping rounds: want at least %d, have %d", min, have)
			}
			if min, have := 2, clientPingRounds; min > have {
				t.Fatalf("client ping rounds: want at least %d, have %d", min, have)
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
