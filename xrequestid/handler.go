package xrequestid

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/tap"
)

type Generator interface {
	Generate() XRequestID
}

type Handler struct {
	gen Generator
}

func New(opts ...Option) *Handler {
	h := new(Handler)
	for _, opt := range opts {
		opt(h)
	}
	h.prepare()
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

func (h *Handler) populate(parent context.Context) context.Context {
	_, ok := Get(parent)
	if ok {
		return parent
	}
	xid := h.gen.Generate()
	ctx := Populate(parent, xid)
	return ctx
}

func (h *Handler) prepare() {
	if h.gen == nil {
		h.gen = ShortuuidGenerator{}
	}
}
