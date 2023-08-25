package ws

import (
	"time"

	"github.com/fasthttp/websocket"
)

type connection interface {
	tuner
	rw
	controller
	closer
}

type tuner interface {
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

type enhPreset struct {
	read struct {
		limit int64
	}

	write struct {
		compression struct {
			enable bool
			level  int
		}
	}
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

func enhance(preset enhPreset, cfg enhConfig, conn connection) (*enhConn, error) {
	if preset.write.compression.enable {
		conn.EnableWriteCompression(true)
		if err := conn.SetCompressionLevel(preset.write.compression.level); err != nil {
			return nil, err
		}
	}
	if preset.read.limit > 0 {
		conn.SetReadLimit(preset.read.limit)
	}
	return &enhConn{conn, cfg}, nil
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
