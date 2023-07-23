package grpc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/einouqo/ext-kit/endpoint"
	"github.com/einouqo/ext-kit/test/service"
)

const (
	addressTestGRPC string = ":8801"
)

func TestUnaryGRPC_ok(t *testing.T) {
	client, tidy, err := prepareTestGRPC(addressTestGRPC)
	if err != nil {
		t.Fatalf("unable to prepare test: %+v", err)
	}
	defer tidy()

	ctx := context.Background()

	req := service.EchoRequest{Message: "unary"}
	resp, err := client.Unary(ctx, req)
	if err != nil {
		t.Fatalf("call error: %+v", err)
	}
	if len(resp.Messages) != 1 {
		t.Fatalf("message: want 1, have %d", len(resp.Messages))
	}
	if want, have := req.Message, resp.Messages[0]; want != have {
		t.Fatalf("message: want %q', have %q", want, have)
	}
}

func TestUnaryGRPC_error(t *testing.T) {
	client, tidy, err := prepareTestGRPC(addressTestGRPC)
	if err != nil {
		t.Fatalf("unable to prepare test: %+v", err)
	}
	defer tidy()

	ctx := context.Background()

	req := service.EchoRequest{Message: "unary", IsError: true}
	_, err = client.Unary(ctx, req)
	if err == nil {
		t.Fatal("want error, have nil")
	}
	if !strings.Contains(err.Error(), req.Message) {
		t.Fatalf("want %q to contain %q", err.Error(), req.Message)
	}
}

func TestInnerStreamGRPC_ok(t *testing.T) {
	client, tidy, err := prepareTestGRPC(addressTestGRPC)
	if err != nil {
		t.Fatalf("unable to prepare test: %+v", err)
	}
	defer tidy()

	ctx := context.Background()

	n := 10
	req := service.EchoRequest{
		Message: "inner",
		Repeat:  uint(n),
	}
	receive, stop, err := client.InnerStream(ctx, req)
	if err != nil {
		t.Fatalf("request: %+v", err)
	}
	defer stop()
	i := 0
	for {
		msg, err := receive()
		switch {
		case errors.Is(err, endpoint.StreamDone):
			if want, have := n, i; want != have {
				t.Fatalf("iteration: want %d, have %d", want, have)
			}
			return
		case err != nil:
			t.Fatalf("request: %+v", err)
		}
		if i > n {
			t.Fatalf("iteration: want less than %d, have %d", n, i)
		}
		if len(msg.Messages) != 1 {
			t.Fatalf("message: want exactly 1 message, have %d", len(msg.Messages))
		}
		if want, have := req.Message, msg.Messages[0]; want != have {
			t.Fatalf("message: want %q', have %q", want, have)
		}
		i++
	}
}

func TestInnerStreamGRPC_error(t *testing.T) {
	client, tidy, err := prepareTestGRPC(addressTestGRPC)
	if err != nil {
		t.Fatalf("unable to prepare test: %+v", err)
	}
	defer tidy()

	var (
		ctx = context.Background()
		i   = 0
		req = service.EchoRequest{
			Message: "inner",
			IsError: true,
		}
	)

	receive, stop, err := client.InnerStream(ctx, req)
	if err != nil {
		t.Fatalf("request: %+v", err)
	}
	defer stop()
	for {
		_, err := receive()
		switch {
		case errors.Is(err, endpoint.StreamDone):
			t.Fatal("want error, have stream done")
		case err != nil:
			if !strings.Contains(err.Error(), req.Message) {
				t.Fatalf("want %q to contain %q", err.Error(), req.Message)
			}
			return
		}
		if i > 0 {
			t.Fatal("iterations: expect only one iteration")
		}
		i++
	}
}

func TestInnerStreamGRPC_cancel(t *testing.T) {
	client, tidy, err := prepareTestGRPC(addressTestGRPC)
	if err != nil {
		t.Fatalf("unable to prepare test: %+v", err)
	}
	defer tidy()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	i := 0
	req := service.EchoRequest{Message: "inner", Repeat: 2, Latency: 10 * time.Millisecond}

	receive, stop, err := client.InnerStream(ctx, req)
	if err != nil {
		t.Fatalf("request: %+v", err)
	}
	defer stop()
	for {
		msg, err := receive()
		switch {
		case errors.Is(err, endpoint.StreamDone):
			t.Fatal("want error, have stream done")
		case err != nil:
			if !strings.Contains(err.Error(), context.Canceled.Error()) {
				t.Fatalf("want %q to contain %q", err.Error(), context.Canceled.Error())
			}
			if want, have := 1, i; want != have {
				t.Fatalf("iteration: want %d, have %d", want, have)
			}
			return
		}
		cancel()
		if i > 1 {
			t.Fatal("iterations: expect two iterations")
		}
		if len(msg.Messages) != 1 {
			t.Fatalf("message: want exactly 1 message, have %d", len(msg.Messages))
		}
		if want, have := req.Message, msg.Messages[0]; want != have {
			t.Fatalf("message: want %q', have %q", want, have)
		}
		i++
	}
}

func TestOuterStreamGRPC_ok(t *testing.T) {
	client, tidy, err := prepareTestGRPC(addressTestGRPC)
	if err != nil {
		t.Fatalf("unable to prepare test: %+v", err)
	}
	defer tidy()

	ctx := context.Background()

	n := 10
	msgs := make([]string, 0, (n+1)/2) // expect triangular number
	sendC := make(chan service.EchoRequest)
	go func() {
		defer close(sendC)
		for i := 0; i < n; i++ {
			req := service.EchoRequest{
				Message: fmt.Sprintf("outer %d", i),
				Repeat:  uint(i),
			}
			sendC <- req
			for j := 0; j < int(req.Repeat); j++ {
				msgs = append(msgs, req.Message)
			}
		}
	}()
	resp, err := client.OuterStream(ctx, sendC)
	if err != nil {
		t.Fatalf("call error: %+v", err)
	}
	if !slices.Equal(resp.Messages, msgs) {
		t.Fatalf("want %q, have %q", msgs, resp.Messages)
	}
}

func TestOuterStreamGRPC_error(t *testing.T) {
	client, tidy, err := prepareTestGRPC(addressTestGRPC)
	if err != nil {
		t.Fatalf("unable to prepare test: %+v", err)
	}
	defer tidy()

	ctx := context.Background()

	n := 10
	errIter := n / 2
	errMsg := ""
	sendC := make(chan service.EchoRequest)
	go func() {
		defer close(sendC)
		for i := 0; i < n; i++ {
			req := service.EchoRequest{
				Message: fmt.Sprintf("outer %d", i),
			}
			if i == errIter {
				req.IsError = true
				errMsg = req.Message
			} else {
				req.Repeat = uint(i)
			}
			sendC <- req
			if req.IsError {
				break
			}
		}
	}()
	_, err = client.OuterStream(ctx, sendC)
	if err == nil {
		t.Fatal("want error, have nil")
	}
	if !strings.Contains(err.Error(), errMsg) {
		t.Fatalf("want %q to contain %q", err.Error(), errMsg)
	}
}

func TestOuterStreamGRPC_cancel(t *testing.T) {
	client, tidy, err := prepareTestGRPC(addressTestGRPC)
	if err != nil {
		t.Fatalf("unable to prepare test: %+v", err)
	}
	defer tidy()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	n := 10
	cancelIter := n / 2
	sendC := make(chan service.EchoRequest)
	go func() {
		defer close(sendC)
		for i := 0; i < n; i++ {
			req := service.EchoRequest{
				Message: fmt.Sprintf("outer %d", i),
				Repeat:  uint(i),
			}
			if i == cancelIter {
				cancel()
				return
			}
			sendC <- req
		}
	}()
	_, err = client.OuterStream(ctx, sendC)
	if err == nil {
		t.Fatal("want error, have nil")
	}
	if !strings.Contains(err.Error(), context.Canceled.Error()) {
		t.Fatalf("want %q to contain %q", err.Error(), context.Canceled.Error())
	}
}

func TestBiStreamGRPC_ok(t *testing.T) {
	client, tidy, err := prepareTestGRPC(addressTestGRPC)
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
	}()
	receive, stop, err := client.BiStream(ctx, sendC)
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

func TestBiStreamGRPC_error(t *testing.T) {
	client, tidy, err := prepareTestGRPC(addressTestGRPC)
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
			mu.Lock()
			if i == errIter {
				req.IsError = true
				errMsg = req.Message
			} else {
				req.Repeat = uint(i)
				for j := 0; j < int(req.Repeat); j++ {
					msgs = append(msgs, req.Message)
				}
			}
			sendC <- req
			mu.Unlock()
		}
	}()

	receive, stop, err := client.BiStream(ctx, sendC)
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
			if !strings.Contains(err.Error(), errMsg) {
				t.Fatalf("want %s to contain %s", err.Error(), errMsg)
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

func TestBiStreamGRPC_stop(t *testing.T) {
	client, tidy, err := prepareTestGRPC(addressTestGRPC)
	if err != nil {
		t.Fatalf("unable to prepare test: %+v", err)
	}
	defer tidy()

	ctx := context.Background()

	n := 10
	nn := n * (n - 1) / 2 // expect triangular like number
	cancelIter := n / 2
	sendC := make(chan service.EchoRequest)

	receive, stop, err := client.BiStream(ctx, sendC)
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
			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("want %s to be grpc status", err)
			}
			if st.Code() != codes.Canceled {
				t.Fatalf("want %s to be canceled", st.Code())
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

func TestBiStreamGRPC_cancel(t *testing.T) {
	client, tidy, err := prepareTestGRPC(addressTestGRPC)
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
				return
			}
			sendC <- req
		}
	}()

	receive, stop, err := client.BiStream(ctx, sendC)
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
			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("want %s to be grpc status", err)
			}
			if st.Code() != codes.Canceled {
				t.Fatalf("want %s to be canceled", st.Code())
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
