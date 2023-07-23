package ws

import "context"

type DecodeFunc[T any] func(context.Context, MessageType, []byte) (T, error)

type EncodeFunc[T any] func(context.Context, T) ([]byte, MessageType, error)
