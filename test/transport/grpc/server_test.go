//go:build integration
// +build integration

package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"

	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/einouqo/ext-kit/test/transport/grpc/pb"
	kitgrpc "github.com/einouqo/ext-kit/transport/grpc"
)

const (
	addressTestServer string = ":8802"
)

func TestUnaryGRPC_before_after(t *testing.T) {
	sTidy, err := prepareServer(
		addressTestServer,
		kitgrpc.WithServerBefore(populateTestValueFromMD),
		kitgrpc.WithServerAfter(populateTestHeaderTrailer),
	)
	if err != nil {
		t.Fatalf("unable to prepare server: %+v", err)
	}
	defer sTidy()
	client, cTidy, err := prepareClientPB(addressTestServer)
	if err != nil {
		t.Fatalf("unable to prepare test: %+v", err)
	}
	defer cTidy()

	headers, trailers, tests := []string{"header_one", "header_two"}, []string{"trail_one", "trail_two"}, []string{"test_one", "test_two"}
	ctx := context.Background()
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		mdTestHeadersKey:  headers,
		mdTestTrailersKey: trailers,
		mdIncomingTestKey: tests,
	})

	var hs, ts metadata.MD
	req := pb.EchoRequest{
		Payload: &pb.EchoPayload{
			Msg: "inner",
		},
	}
	resp, err := client.Unary(
		ctx, &req,
		grpc.Header(&hs),
		grpc.Trailer(&ts),
	)
	if err != nil {
		t.Fatalf("call error: %+v", err)
	}
	if len(resp.GetPayloads()) != 1 {
		t.Fatalf("message: want 1, have %d", len(resp.GetPayloads()))
	}
	if want, have := req.GetPayload().GetMsg(), resp.GetPayloads()[0].GetMsg(); want != have {
		t.Fatalf("message: want %q', have %q", want, have)
	}
	if want, have := headers, hs[mdTestHeadersKey]; !slices.Equal(want, have) {
		t.Fatalf("header: want %q, have %q", want, have)
	}
	if want, have := trailers, ts[mdTestTrailersKey]; !slices.Equal(want, have) {
		t.Fatalf("trailer: want %q, have %q", want, have)
	}
}

func TestInnerStreamGRPC_before_after(t *testing.T) {
	sTidy, err := prepareServer(
		addressTestServer,
		kitgrpc.WithServerBefore(populateTestValueFromMD),
		kitgrpc.WithServerAfter(populateTestHeaderTrailer),
	)
	if err != nil {
		t.Fatalf("unable to prepare server: %+v", err)
	}
	defer sTidy()
	client, cTidy, err := prepareClientPB(addressTestServer)
	if err != nil {
		t.Fatalf("unable to prepare test: %+v", err)
	}
	defer cTidy()

	headers, trailers, tests := []string{"header_one", "header_two"}, []string{"trail_one", "trail_two"}, []string{"test_one", "test_two"}
	ctx := context.Background()
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		mdTestHeadersKey:  headers,
		mdTestTrailersKey: trailers,
		mdIncomingTestKey: tests,
	})

	n := 10
	req := pb.EchoRequest{
		Payload: &pb.EchoPayload{
			Msg: "inner",
		},
		Options: &pb.EchoRequest_Repeat{Repeat: uint32(n)},
	}
	stream, err := client.InnerStream(ctx, &req)
	if err != nil {
		t.Fatalf("request: %+v", err)
	}

	hs, err := stream.Header()
	if err != nil {
		t.Fatalf("header: %+v", err)
	}
	if want, have := headers, hs[mdTestHeadersKey]; !slices.Equal(want, have) {
		t.Fatalf("header: want %q, have %q", want, have)
	}
	i := 0
	for {
		resp, err := stream.Recv()
		switch {
		case errors.Is(err, io.EOF):
			if want, have := n, i; want != have {
				t.Fatalf("iteration: want %d, have %d", want, have)
			}
			ts := stream.Trailer()
			if want, have := trailers, ts[mdTestTrailersKey]; !slices.Equal(want, have) {
				t.Fatalf("trailer: want %q, have %q", want, have)
			}
			if want, have := tests, ts[mdOuterTestKey]; !slices.Equal(want, have) {
				t.Fatalf("trailer: want %q, have %q", want, have)
			}
			return
		case err != nil:
			t.Fatalf("request: %+v", err)
		}
		if i > n {
			t.Fatalf("iteration: want less than %d, have %d", n, i)
		}
		if len(resp.GetPayloads()) != 1 {
			t.Fatalf("message: want exactly 1 message, have %d", len(resp.GetPayloads()))
		}
		if want, have := req.GetPayload().GetMsg(), resp.GetPayloads()[0].GetMsg(); want != have {
			t.Fatalf("message: want %q', have %q", want, have)
		}
		i++
	}
}

func TestOuterStreamGRPC_before_after(t *testing.T) {
	sTidy, err := prepareServer(
		addressTestServer,
		kitgrpc.WithServerBefore(populateTestValueFromMD),
		kitgrpc.WithServerAfter(populateTestHeaderTrailer),
	)
	if err != nil {
		t.Fatalf("unable to prepare server: %+v", err)
	}
	defer sTidy()
	client, cTidy, err := prepareClientPB(addressTestServer)
	if err != nil {
		t.Fatalf("unable to prepare test: %+v", err)
	}
	defer cTidy()

	headers, trailers, tests := []string{"header_one", "header_two"}, []string{"trail_one", "trail_two"}, []string{"test_one", "test_two"}
	ctx := context.Background()
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		mdTestHeadersKey:  headers,
		mdTestTrailersKey: trailers,
		mdIncomingTestKey: tests,
	})

	stream, err := client.OuterStream(ctx)
	if err != nil {
		t.Fatalf("unable to create stream: %+v", err)
	}

	n := 10
	msgs := make([]string, 0, (n+1)/2) // expect triangular number
	for i := 0; i < n; i++ {
		req := &pb.EchoRequest{
			Payload: &pb.EchoPayload{
				Msg: fmt.Sprintf("bi %d", i),
			},
			Options: &pb.EchoRequest_Repeat{Repeat: uint32(i)},
		}
		if err := stream.SendMsg(req); err != nil {
			t.Fatalf("unable to send message: %+v", err)
		}
		for j := 0; j < int(req.GetRepeat()); j++ {
			msgs = append(msgs, req.GetPayload().GetMsg())
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		t.Fatalf("call error: %+v", err)
	}
	rmsgs := make([]string, 0, len(resp.GetPayloads()))
	for _, p := range resp.GetPayloads() {
		rmsgs = append(rmsgs, p.GetMsg())
	}
	if !slices.Equal(rmsgs, msgs) {
		t.Fatalf("want %q, have %q", msgs, rmsgs)
	}

	hs, err := stream.Header()
	if err != nil {
		t.Fatalf("header: %+v", err)
	}
	if want, have := headers, hs[mdTestHeadersKey]; !slices.Equal(want, have) {
		t.Fatalf("header: want %q, have %q", want, have)
	}

	ts := stream.Trailer()
	if want, have := trailers, ts[mdTestTrailersKey]; !slices.Equal(want, have) {
		t.Fatalf("trailer: want %q, have %q", want, have)
	}
	if want, have := tests, ts[mdOuterTestKey]; !slices.Equal(want, have) {
		t.Fatalf("trailer: want %q, have %q", want, have)
	}
}

func TestServerBiStreamGRPC_before_after(t *testing.T) {
	sTidy, err := prepareServer(
		addressTestServer,
		kitgrpc.WithServerBefore(populateTestValueFromMD),
		kitgrpc.WithServerAfter(populateTestHeaderTrailer),
	)
	if err != nil {
		t.Fatalf("unable to prepare server: %+v", err)
	}
	defer sTidy()
	client, cTidy, err := prepareClientPB(addressTestServer)
	if err != nil {
		t.Fatalf("unable to prepare test: %+v", err)
	}
	defer cTidy()

	headers, trailers, tests := []string{"header_one", "header_two"}, []string{"trail_one", "trail_two"}, []string{"test_one", "test_two"}
	ctx := context.Background()
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		mdTestHeadersKey:  headers,
		mdTestTrailersKey: trailers,
		mdIncomingTestKey: tests,
	})

	stream, err := client.BiStream(ctx)
	if err != nil {
		t.Fatalf("unable to create stream: %+v", err)
	}

	n := 10
	nn := n * (n - 1) / 2 // expect triangular like number
	mu := sync.Mutex{}
	msgs := make([]string, 0, nn)
	for i := 0; i < n; i++ {
		req := &pb.EchoRequest{
			Payload: &pb.EchoPayload{
				Msg: fmt.Sprintf("bi %d", i),
			},
			Options: &pb.EchoRequest_Repeat{Repeat: uint32(i)},
		}
		if err := stream.SendMsg(req); err != nil {
			t.Fatalf("unable to send message: %+v", err)
		}
		mu.Lock()
		for j := 0; j < int(req.GetRepeat()); j++ {
			msgs = append(msgs, req.GetPayload().GetMsg())
		}
		mu.Unlock()
	}
	if err := stream.CloseSend(); err != nil {
		t.Fatalf("unable to close stream: %+v", err)
	}

	hs, err := stream.Header()
	if err != nil {
		t.Fatalf("header: %+v", err)
	}
	if want, have := headers, hs[mdTestHeadersKey]; !slices.Equal(want, have) {
		t.Fatalf("header: want %q, have %q", want, have)
	}
	i := 0
	for {
		resp, err := stream.Recv()
		switch {
		case errors.Is(err, io.EOF):
			if want, have := nn, i; want != have {
				t.Fatalf("iteration: want %d, have %d", want, have)
			}
			if len(msgs) != 0 {
				t.Fatalf("message: want empty, have %q", msgs)
			}
			ts := stream.Trailer()
			if want, have := trailers, ts[mdTestTrailersKey]; !slices.Equal(want, have) {
				t.Fatalf("trailer: want %q, have %q", want, have)
			}
			if want, have := tests, ts[mdOuterTestKey]; !slices.Equal(want, have) {
				t.Fatalf("trailer: want %q, have %q", want, have)
			}
			return
		case err != nil:
			t.Fatalf("request: %+v", err)
		}
		if i > nn {
			t.Fatalf("iteration: want less than %d, have %d", nn, i)
		}
		if len(resp.Payloads) != 1 {
			t.Fatalf("message: want exactly 1 message, have %d", len(resp.Payloads))
		}
		mu.Lock()
		if !slices.Contains(msgs, resp.Payloads[0].Msg) {
			t.Fatalf("message: want %q to contain in %q", resp.Payloads[0].Msg, msgs)
		}
		once := true
		msgs = slices.DeleteFunc(msgs, func(s string) bool {
			del := s == resp.Payloads[0].Msg
			if del && once {
				defer func() { once = false }()
			}
			return del && once
		})
		mu.Unlock()
		i++
	}
}
