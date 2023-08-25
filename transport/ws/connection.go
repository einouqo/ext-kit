package ws

import (
	"time"

	"github.com/fasthttp/websocket"
)

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
	*websocket.Conn
	cfg enhConfig
}

func enhance(preset enhPreset, cfg enhConfig, conn *websocket.Conn) (*enhConn, error) {
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
	if err := c.updateReadDeadline(c.Conn); err != nil {
		return 0, nil, err
	}
	return c.Conn.ReadMessage()
}

func (c *enhConn) WriteMessage(messageType int, data []byte) error {
	if err := c.updateWriteDeadline(c.Conn); err != nil {
		return err
	}
	if c.cfg.write.prepared {
		return c.writePrepared(messageType, data)
	}
	return c.Conn.WriteMessage(messageType, data)
}

func (c *enhConn) writePrepared(messageType int, data []byte) error {
	pm, err := websocket.NewPreparedMessage(messageType, data)
	if err != nil {
		return err
	}
	return c.Conn.WritePreparedMessage(pm)
}

func (c *enhConn) updateWriteDeadline(conn *websocket.Conn) error {
	if c.cfg.write.timeout > 0 {
		deadline := time.Now().Add(c.cfg.write.timeout)
		return conn.SetWriteDeadline(deadline)
	}
	return nil
}

func (c *enhConn) updateReadDeadline(conn *websocket.Conn) error {
	if c.cfg.read.timeout > 0 {
		deadline := time.Now().Add(c.cfg.read.timeout)
		return conn.SetReadDeadline(deadline)
	}
	return nil
}
