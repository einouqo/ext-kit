package ws

import (
	"net/http"
	"time"

	"github.com/fasthttp/websocket"
)

type Upgrader interface {
	SetHandshakeTimeout(timeout time.Duration)
	SetReadBufferSize(size int)
	SetWriteBufferSize(size int)
	SetWriteBufferPool(pool websocket.BufferPool)
	SetSubprotocols(protocols []string)
	SetErrorWriter(writer func(w http.ResponseWriter, r *http.Request, status int, reason error))
	SetCheckOrigin(check func(r *http.Request) bool)
	SetEnableCompression(enabled bool)
}

type upgrader struct {
	*websocket.Upgrader
}

func (u upgrader) SetHandshakeTimeout(timeout time.Duration) {
	u.HandshakeTimeout = timeout
}

func (u upgrader) SetReadBufferSize(size int) {
	u.ReadBufferSize = size
}

func (u upgrader) SetWriteBufferSize(size int) {
	u.WriteBufferSize = size
}

func (u upgrader) SetWriteBufferPool(pool websocket.BufferPool) {
	u.WriteBufferPool = pool
}

func (u upgrader) SetSubprotocols(protocols []string) {
	u.Subprotocols = protocols
}

func (u upgrader) SetErrorWriter(writer func(w http.ResponseWriter, r *http.Request, status int, reason error)) {
	u.Error = writer
}

func (u upgrader) SetCheckOrigin(check func(r *http.Request) bool) {
	u.CheckOrigin = check
}

func (u upgrader) SetEnableCompression(enabled bool) {
	u.EnableCompression = enabled
}
