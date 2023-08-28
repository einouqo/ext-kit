package ws

import (
	"io"
	"time"

	"github.com/fasthttp/websocket"
)

type WriteMod uint8

const (
	WriteModPlain WriteMod = iota + 1
	WriteModPrepared
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
	NextWriter(messageType int) (io.WriteCloser, error)
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
		timeout time.Duration
		mod     WriteMod
	}
}

type enhConn struct {
	connection
	cfg enhConfig
}

func (c enhConn) ReadMessage() (messageType int, p []byte, err error) {
	if err := c.updateReadDeadline(c.connection); err != nil {
		return 0, nil, err
	}
	return c.connection.ReadMessage()
}

func (c enhConn) WriteMessage(messageType int, data []byte) error {
	if err := c.updateWriteDeadline(c.connection); err != nil {
		return err
	}
	switch c.cfg.write.mod {
	case WriteModPlain:
		return c.writePlain(messageType, data)
	case WriteModPrepared:
		return c.writePrepared(messageType, data)
	default:
		return c.connection.WriteMessage(messageType, data)
	}
}

func (c enhConn) writePlain(messageType int, data []byte) error {
	w, err := c.connection.NextWriter(messageType)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	return w.Close()
}

func (c enhConn) writePrepared(messageType int, data []byte) error {
	pm, err := websocket.NewPreparedMessage(messageType, data)
	if err != nil {
		return err
	}
	return c.connection.WritePreparedMessage(pm)
}

func (c enhConn) updateWriteDeadline(conn connection) error {
	if c.cfg.write.timeout > 0 {
		deadline := time.Now().Add(c.cfg.write.timeout)
		return conn.SetWriteDeadline(deadline)
	}
	return nil
}

func (c enhConn) updateReadDeadline(conn connection) error {
	if c.cfg.read.timeout > 0 {
		deadline := time.Now().Add(c.cfg.read.timeout)
		return conn.SetReadDeadline(deadline)
	}
	return nil
}
