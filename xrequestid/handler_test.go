package xrequestid

import (
	"context"
	"net/http"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/tap"

	"github.com/einouqo/ext-kit/util"
)

func TestHandler_implement(t *testing.T) {
	h := New()
	var _ tap.ServerInHandle = h.InTapHandler
	var _ grpc.UnaryServerInterceptor = h.UnaryInterceptor
	var _ util.Middleware[http.Handler] = h.PopulateHTTP
}

func TestHandler_InTapHandler(t *testing.T) {
	h := New()
	ctx := context.Background()
	ctx, err := h.InTapHandler(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if _, ok := Get(ctx); !ok {
		t.Fatal("expected request id in context")
	}

	const expected = "test"
	ctx = Populate(ctx, expected)
	ctx, err = h.InTapHandler(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if got, ok := Get(ctx); !ok || got != expected {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestHandler_UnaryInterceptor(t *testing.T) {
	h := New()
	ctx := context.Background()
	_, err := h.UnaryInterceptor(ctx, nil, nil, func(ctx context.Context, req any) (any, error) {
		if _, ok := Get(ctx); !ok {
			t.Fatal("expected request id in context")
		}
		return req, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	const expected = "test"
	ctx = Populate(ctx, expected)
	_, err = h.UnaryInterceptor(ctx, nil, nil, func(ctx context.Context, req any) (any, error) {
		if got, ok := Get(ctx); !ok || got != expected {
			t.Fatalf("expected %s, got %s", expected, got)
		}
		return req, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestHandler_PopulateHTTP(t *testing.T) {
	const header = "X-Test-ID"
	h := New(WithHeaderHTTP(header))
	h.PopulateHTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := Get(r.Context()); !ok {
			t.Fatal("expected request id in context")
		}
	})).ServeHTTP(nil, &http.Request{})

	const expected = "test"
	headers := make(http.Header)
	headers.Add(header, expected)
	h.PopulateHTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, ok := Get(r.Context()); !ok || got != expected {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	})).ServeHTTP(nil, &http.Request{Header: headers})
}
