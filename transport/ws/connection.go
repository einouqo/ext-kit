package ws

import (
	"time"

	"github.com/fasthttp/websocket"
)

type connection interface {
	Tuner
	rw
	controller
	closer
}

type Tuner interface {
	EnableWriteCompression(enable bool)
	SetCompressionLevel(level int) error
	SetReadLimit(limit int64)
}

type rw interface {
	SetReadDeadline(t time.Time) error
	ReadMessage() (messageType int, p []byte, err error)
	SetWriteDeadline(t time.Time) error
	WriteMessage(messageType int, data []byte) error
	WritePreparedMessage(pm *websocket.PreparedMessage) error
}

type controller interface {
	WriteControl(messageType int, data []byte, deadline time.Time) error
	PongHandler() func(string) error
	SetPongHandler(h func(string) error)
}

type closer interface {
	Close() error
}

type enhConfig struct {
	read struct {
		timeout time.Duration
	}

	write struct {
		timeout  time.Duration
		prepared bool
	}
}

type enhConn struct {
	connection
	cfg enhConfig
}

func (c *enhConn) ReadMessage() (messageType int, p []byte, err error) {
	if err := c.updateReadDeadline(c.connection); err != nil {
		return 0, nil, err
	}
	return c.connection.ReadMessage()
}

func (c *enhConn) WriteMessage(messageType int, data []byte) error {
	if err := c.updateWriteDeadline(c.connection); err != nil {
		return err
	}
	if c.cfg.write.prepared {
		return c.writePrepared(messageType, data)
	}
	return c.connection.WriteMessage(messageType, data)
}

func (c *enhConn) writePrepared(messageType int, data []byte) error {
	pm, err := websocket.NewPreparedMessage(messageType, data)
	if err != nil {
		return err
	}
	return c.connection.WritePreparedMessage(pm)
}

func (c *enhConn) updateWriteDeadline(conn connection) error {
	if c.cfg.write.timeout > 0 {
		deadline := time.Now().Add(c.cfg.write.timeout)
		return conn.SetWriteDeadline(deadline)
	}
	return nil
}

func (c *enhConn) updateReadDeadline(conn connection) error {
	if c.cfg.read.timeout > 0 {
		deadline := time.Now().Add(c.cfg.read.timeout)
		return conn.SetReadDeadline(deadline)
	}
	return nil
}
