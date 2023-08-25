package ws

import (
	"context"
	"errors"
	"net"
	"syscall"
	"time"

	"github.com/fasthttp/websocket"
)

type hbConfig struct {
	period, await time.Duration
	pinging       Pinging
}

func heartbeat(
	ctx context.Context,
	cfg hbConfig,
	conn controller,
	done <-chan struct{},
) error {
	pongCh := make(chan struct{})
	handler := conn.PongHandler()
	conn.SetPongHandler(func(msg string) error {
		select {
		case pongCh <- struct{}{}:
		case <-done:
		}
		return handler(msg)
	})

	ticker := time.NewTicker(cfg.period)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return nil
		case <-ticker.C:
			msg, deadline := cfg.pinging(ctx)
			err := conn.WriteControl(websocket.PingMessage, msg, deadline)
			switch {
			case errors.Is(err, net.ErrClosed):
				return nil
			case errors.Is(err, syscall.EPIPE): // broken pipe can appear on closed underlying tcp connection by peer
				return nil
			case errors.Is(err, websocket.ErrCloseSent):
				return nil
			case err != nil:
				return err
			}
		}
		select {
		case <-done:
			return nil
		case <-time.After(cfg.await):
			return context.DeadlineExceeded
		case <-pongCh:
			ticker.Reset(cfg.period)
		}
	}
}
