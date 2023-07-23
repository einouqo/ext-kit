package grpc

import (
	"context"

	"google.golang.org/protobuf/proto"
)

type DecodeFunc[T any] func(context.Context, proto.Message) (T, error)

type EncodeFunc[T any] func(context.Context, T) (proto.Message, error)
