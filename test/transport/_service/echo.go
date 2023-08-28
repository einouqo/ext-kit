package service

import (
	"context"
	"errors"
	"time"

	"github.com/einouqo/ext-kit/endpoint"
)

type EchoRequest struct {
	Message string        `json:"message"`
	Latency time.Duration `json:"latency"`
	IsError bool          `json:"is_error"`
	Repeat  uint          `json:"repeat"`
}

type EchoResponse struct {
	Messages []string `json:"messages"`
}

type Echo struct{}

func NewEcho() *Echo {
	return new(Echo)
}

func (s Echo) Once(_ context.Context, req EchoRequest) (EchoResponse, error) {
	if req.Latency > 0 {
		time.Sleep(req.Latency)
	}
	if req.IsError {
		return EchoResponse{}, errors.New(req.Message)
	}
	return EchoResponse{Messages: []string{req.Message}}, nil
}

func (s Echo) Inner(ctx context.Context, req EchoRequest) (endpoint.Receive[EchoResponse], error) {
	respC := make(chan EchoResponse)
	errC := make(chan error, 1)
	go func() {
		defer close(respC)
		defer close(errC)
		if req.IsError {
			errC <- errors.New(req.Message)
			return
		}
		for i := 0; i < int(req.Repeat); i++ {
			if req.Latency > 0 {
				time.Sleep(req.Latency)
			}
			select {
			case <-ctx.Done():
				return
			case respC <- EchoResponse{Messages: []string{req.Message}}:
				// nothing
			}
		}
	}()
	return func() (EchoResponse, error) {
		resp, ok := <-respC
		if ok {
			return resp, nil
		}
		err, ok := <-errC
		if ok {
			return EchoResponse{}, err
		}
		return EchoResponse{}, endpoint.StreamDone
	}, nil
}

func (s Echo) Outer(_ context.Context, receiver <-chan EchoRequest) (EchoResponse, error) {
	msgs := make([]string, 0, 1<<8)
	for req := range receiver {
		if req.Latency > 0 {
			time.Sleep(req.Latency)
		}
		if req.IsError {
			return EchoResponse{}, errors.New(req.Message)
		}
		for i := 0; i < int(req.Repeat); i++ {
			msgs = append(msgs, req.Message)
		}
	}
	return EchoResponse{Messages: msgs}, nil
}

func (s Echo) Bi(ctx context.Context, receiver <-chan EchoRequest) (endpoint.Receive[EchoResponse], error) {
	respC := make(chan EchoResponse)
	errC := make(chan error, 1)
	go func() {
		defer close(respC)
		defer close(errC)
		i := 0
		for req := range receiver {
			i++
			if req.Latency > 0 {
				time.Sleep(req.Latency)
			}
			if req.IsError {
				errC <- errors.New(req.Message)
				return
			}
			for i := 0; i < int(req.Repeat); i++ {
				select {
				case <-ctx.Done():
					return
				case respC <- EchoResponse{Messages: []string{req.Message}}:
					// nothing
				}
			}
		}
	}()
	return func() (EchoResponse, error) {
		if resp, ok := <-respC; ok {
			return resp, nil
		}
		if err, ok := <-errC; ok {
			return EchoResponse{}, err
		}
		return EchoResponse{}, endpoint.StreamDone
	}, nil
}
