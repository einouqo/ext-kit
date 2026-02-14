//go:build integration

package grpc

import (
	"context"

	"google.golang.org/grpc/metadata"
)

type ctxTestKeyType struct{}

var (
	ctxTestKey = ctxTestKeyType{}

	mdIncomingTestKey = "md-incoming-test"
	mdOuterTestKey    = "md-outer-test"
	mdTestHeadersKey  = "md-test-headers"
	mdTestTrailersKey = "md-test-trailers"
)

func populateTestValueFromMD(ctx context.Context, md metadata.MD) context.Context {
	return context.WithValue(ctx, ctxTestKey, md.Get(mdIncomingTestKey))
}

func populateTestHeaderTrailer(ctx context.Context, headers *metadata.MD, trailers *metadata.MD) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	headers.Append(mdTestHeadersKey, md.Get(mdTestHeadersKey)...)
	trailers.Append(mdTestTrailersKey, md.Get(mdTestTrailersKey)...)
	tests, ok := ctx.Value(ctxTestKey).([]string)
	if !ok {
		return ctx
	}
	trailers.Append(mdOuterTestKey, tests...)
	return ctx
}
