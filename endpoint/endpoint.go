package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
)

var (
	ErrTypeMismatch = errors.New("type mismatch")
)

type Unary[IN, OUT any] func(ctx context.Context, request IN) (response OUT, err error)

func (e Unary[IN, OUT]) Kit() endpoint.Endpoint {
	return func(ctx context.Context, request any) (response any, err error) {
		req, ok := request.(IN)
		if !ok && request != nil {
			return nil, ErrTypeMismatch
		}
		return e(ctx, req)
	}
}

func FromKit[IN, OUT any](e endpoint.Endpoint) Unary[IN, OUT] {
	return func(ctx context.Context, in IN) (out OUT, err error) {
		resp, err := e(ctx, in)
		if err != nil {
			return out, err
		}
		out, ok := resp.(OUT)
		if !ok {
			return out, ErrTypeMismatch
		}
		return out, nil
	}
}

type InnerStream[IN, OUT any] func(ctx context.Context, request IN) (Receive[OUT], Stop, error)

type OuterStream[IN, OUT any] func(ctx context.Context, receiver <-chan IN) (response OUT, err error)

type BiStream[IN, OUT any] func(ctx context.Context, receiver <-chan IN) (Receive[OUT], Stop, error)

type Endpoint[IN, OUT any] interface {
	Unary[IN, OUT] | InnerStream[IN, OUT] | OuterStream[IN, OUT] | BiStream[IN, OUT]
}
