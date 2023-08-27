package ws

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/fasthttp/websocket"
)

//nolint:interfacebloat
type Dialler interface {
	SetNetDial(dial func(network, addr string) (net.Conn, error))
	SetNetDialContext(func(ctx context.Context, network, addr string) (net.Conn, error))
	SetNetDialTLSContext(func(ctx context.Context, network, addr string) (net.Conn, error))
	SetProxy(func(*http.Request) (*url.URL, error))
	SetTLSClientConfig(*tls.Config)
	SetHandshakeTimeout(time.Duration)
	SetReadBufferSize(int)
	SetWriteBufferSize(int)
	SetWriteBufferPool(websocket.BufferPool)
	SetSubprotocols([]string)
	SetEnableCompression(bool)
	SetJar(http.CookieJar)
}

type dialler struct {
	*websocket.Dialer
}

func (d dialler) SetNetDial(dial func(network, addr string) (net.Conn, error)) {
	d.NetDial = dial
}

func (d dialler) SetNetDialContext(dial func(ctx context.Context, network, addr string) (net.Conn, error)) {
	d.NetDialContext = dial
}

func (d dialler) SetNetDialTLSContext(dial func(ctx context.Context, network, addr string) (net.Conn, error)) {
	d.NetDialTLSContext = dial
}

func (d dialler) SetProxy(proxy func(*http.Request) (*url.URL, error)) {
	d.Proxy = proxy
}

func (d dialler) SetTLSClientConfig(cfg *tls.Config) {
	d.TLSClientConfig = cfg
}

func (d dialler) SetHandshakeTimeout(timeout time.Duration) {
	d.HandshakeTimeout = timeout
}

func (d dialler) SetReadBufferSize(size int) {
	d.ReadBufferSize = size
}

func (d dialler) SetWriteBufferSize(size int) {
	d.WriteBufferSize = size
}

func (d dialler) SetWriteBufferPool(pool websocket.BufferPool) {
	d.WriteBufferPool = pool
}

func (d dialler) SetSubprotocols(protocols []string) {
	d.Subprotocols = protocols
}

func (d dialler) SetEnableCompression(enable bool) {
	d.EnableCompression = enable
}

func (d dialler) SetJar(jar http.CookieJar) {
	d.Jar = jar
}
