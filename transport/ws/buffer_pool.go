package ws

import "github.com/fasthttp/websocket"

type pool[T any] interface {
	Get() T
	Put(T)
}

type BufferPool = pool[any]

var _ BufferPool = websocket.BufferPool(nil)
