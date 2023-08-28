package main

import (
	"context"

	"github.com/einouqo/ext-kit/endpoint"
)

type Streamer struct{}

func (Streamer) Echo(_ context.Context, receiver <-chan Message) (endpoint.Receive[Message], error) {
	return func() (Message, error) {
		if resp, ok := <-receiver; ok {
			return resp, nil
		}
		return Message{}, endpoint.StreamDone
	}, nil
}
