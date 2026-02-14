package ws

import (
	"time"

	"github.com/fasthttp/websocket"
)

type WriteMod uint8

const (
	WriteModPlain WriteMod = iota + 1
	WriteModPrepared
)

type connConfig struct {
	read struct {
		timeout time.Duration
	}

	write struct {
		timeout time.Duration
		mod     WriteMod
	}
}

type conn struct {
	ws     *websocket.Conn
	config connConfig
}

var (
	_ Tuner = (*conn)(nil)
)

func (c *conn) EnableWriteCompression(enable bool)  { c.ws.EnableWriteCompression(enable) }
func (c *conn) SetCompressionLevel(level int) error { return c.ws.SetCompressionLevel(level) }
func (c *conn) SetReadLimit(limit int64)            { c.ws.SetReadLimit(limit) }

func (c *conn) PongHandler() func(string) error     { return c.ws.PongHandler() }
func (c *conn) SetPongHandler(h func(string) error) { c.ws.SetPongHandler(h) }

func (c *conn) ReadMessage() (messageType int, p []byte, err error) {
	if err := c.updateReadDeadline(c.ws); err != nil {
		return 0, nil, err
	}
	return c.ws.ReadMessage()
}

func (c *conn) updateWriteDeadline(conn *websocket.Conn) error {
	if c.config.write.timeout > 0 {
		deadline := time.Now().Add(c.config.write.timeout)
		return conn.SetWriteDeadline(deadline)
	}
	return nil
}

func (c *conn) WriteMessage(messageType int, data []byte) error {
	if err := c.updateWriteDeadline(c.ws); err != nil {
		return err
	}
	switch c.config.write.mod {
	case WriteModPlain:
		w, err := c.ws.NextWriter(messageType)
		if err != nil {
			return err
		}
		defer func() { _ = w.Close() }()
		_, err = w.Write(data)
		return err
	case WriteModPrepared:
		pm, err := websocket.NewPreparedMessage(messageType, data)
		if err != nil {
			return err
		}
		return c.ws.WritePreparedMessage(pm)
	default:
		return c.ws.WriteMessage(messageType, data)
	}
}

func (c *conn) updateReadDeadline(conn *websocket.Conn) error {
	if c.config.read.timeout > 0 {
		deadline := time.Now().Add(c.config.read.timeout)
		return conn.SetReadDeadline(deadline)
	}
	return nil
}

func (c *conn) WriteControl(messageType int, data []byte, deadline time.Time) error {
	return c.ws.WriteControl(messageType, data, deadline)
}

func (c *conn) Close() error { return c.ws.Close() }
