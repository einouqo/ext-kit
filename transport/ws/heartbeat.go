package ws

import (
	"context"
	"errors"
	"net"
	"syscall"
	"time"

	"github.com/fasthttp/websocket"
)

type heartbeatConfig struct {
	period, await time.Duration
	pinging       Pinging
}

func heartbeat(
	ctx context.Context,
	cfg heartbeatConfig,
	conn *conn,
) error {
	pong := make(chan struct{})
	handler := conn.PongHandler()
	conn.SetPongHandler(func(msg string) error {
		select {
		case pong <- struct{}{}:
		case <-ctx.Done():
		}
		return handler(msg)
	})

	ticker := time.NewTicker(cfg.period)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			msg, deadline := cfg.pinging(ctx)
			err := conn.WriteControl(websocket.PingMessage, msg, deadline)
			switch {
			case
				errors.Is(err, net.ErrClosed),
				errors.Is(err, syscall.EPIPE), // broken pipe can appear on closed underlying TCP connection by peer
				errors.Is(err, websocket.ErrCloseSent):
				return nil
			case err != nil:
				return err
			}
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(cfg.await):
			return context.DeadlineExceeded
		case <-pong:
			ticker.Reset(cfg.period)
		}
	}
}
