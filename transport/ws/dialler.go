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

type Dial = func(network, addr string) (net.Conn, error)
type DialContext = func(ctx context.Context, network, addr string) (net.Conn, error)
type Proxy = func(*http.Request) (*url.URL, error)

type Dialler interface {
	SetNetDial(Dial)
	SetNetDialContext(DialContext)
	SetNetDialTLSContext(DialContext)
	SetProxy(Proxy)
	SetTLSClientConfig(*tls.Config)
	SetHandshakeTimeout(time.Duration)
	SetReadBufferSize(int)
	SetWriteBufferSize(int)
	SetWriteBufferPool(BufferPool)
	SetSubprotocols([]string)
	SetEnableCompression(bool)
	SetJar(http.CookieJar)
}

type dialler struct {
	ws *websocket.Dialer
}

var _ Dialler = (*dialler)(nil)

func (d *dialler) SetNetDial(dial Dial)                      { d.ws.NetDial = dial }
func (d *dialler) SetNetDialContext(dial DialContext)        { d.ws.NetDialContext = dial }
func (d *dialler) SetNetDialTLSContext(dial DialContext)     { d.ws.NetDialTLSContext = dial }
func (d *dialler) SetProxy(proxy Proxy)                      { d.ws.Proxy = proxy }
func (d *dialler) SetTLSClientConfig(cfg *tls.Config)        { d.ws.TLSClientConfig = cfg }
func (d *dialler) SetHandshakeTimeout(timeout time.Duration) { d.ws.HandshakeTimeout = timeout }
func (d *dialler) SetReadBufferSize(size int)                { d.ws.ReadBufferSize = size }
func (d *dialler) SetWriteBufferSize(size int)               { d.ws.WriteBufferSize = size }
func (d *dialler) SetWriteBufferPool(pool BufferPool)        { d.ws.WriteBufferPool = pool }
func (d *dialler) SetSubprotocols(protocols []string)        { d.ws.Subprotocols = protocols }
func (d *dialler) SetEnableCompression(enable bool)          { d.ws.EnableCompression = enable }
func (d *dialler) SetJar(jar http.CookieJar)                 { d.ws.Jar = jar }
