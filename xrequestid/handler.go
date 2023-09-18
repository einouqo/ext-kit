package xrequestid

import (
	"context"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/tap"
)

type Generator interface {
	Generate() XRequestID
}

type Handler struct {
	gen    Generator
	header string
}

func New(opts ...Option) *Handler {
	h := &Handler{
		gen:    ShortuuidGenerator{},
		header: "X-Request-ID",
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *Handler) InTapHandler(ctx context.Context, _ *tap.Info) (context.Context, error) {
	ctx = h.populate(ctx)
	return ctx, nil
}

func (h *Handler) UnaryInterceptor(
	ctx context.Context,
	req any,
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	ctx = h.populate(ctx)
	return handler(ctx, req)
}

func (h *Handler) PopulateHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		xid := r.Header.Get(h.header)
		if xid == "" {
			xid = h.gen.Generate()
		}
		ctx := Populate(r.Context(), xid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) populate(parent context.Context) context.Context {
	_, ok := Get(parent)
	if ok {
		return parent
	}
	xid := h.gen.Generate()
	ctx := Populate(parent, xid)
	return ctx
}
