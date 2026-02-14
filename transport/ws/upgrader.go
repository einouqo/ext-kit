package ws

import (
	"net/http"
	"time"

	"github.com/fasthttp/websocket"
)

type ErrorWriter = func(w http.ResponseWriter, r *http.Request, status int, reason error)
type CheckOrigin = func(r *http.Request) bool

type Upgrader interface {
	SetHandshakeTimeout(time.Duration)
	SetReadBufferSize(int)
	SetWriteBufferSize(int)
	SetWriteBufferPool(BufferPool)
	SetSubprotocols(protocols []string)
	SetErrorWriter(ErrorWriter)
	SetCheckOrigin(check CheckOrigin)
	SetEnableCompression(enabled bool)
}

type upgrader struct {
	ws websocket.Upgrader
}

var _ Upgrader = (*upgrader)(nil)

func (u *upgrader) SetHandshakeTimeout(timeout time.Duration) { u.ws.HandshakeTimeout = timeout }
func (u *upgrader) SetReadBufferSize(size int)                { u.ws.ReadBufferSize = size }
func (u *upgrader) SetWriteBufferSize(size int)               { u.ws.WriteBufferSize = size }
func (u *upgrader) SetWriteBufferPool(pool BufferPool)        { u.ws.WriteBufferPool = pool }
func (u *upgrader) SetSubprotocols(protocols []string)        { u.ws.Subprotocols = protocols }
func (u *upgrader) SetErrorWriter(w ErrorWriter)              { u.ws.Error = w }
func (u *upgrader) SetCheckOrigin(check CheckOrigin)          { u.ws.CheckOrigin = check }
func (u *upgrader) SetEnableCompression(enabled bool)         { u.ws.EnableCompression = enabled }

func (u *upgrader) Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*websocket.Conn, error) {
	return u.ws.Upgrade(w, r, responseHeader)
}
