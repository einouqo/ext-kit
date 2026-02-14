package transport

import (
	"context"

	"github.com/go-kit/kit/transport"
)

type ErrorHandlers []transport.ErrorHandler

var _ transport.ErrorHandler = (ErrorHandlers)(nil)

func (e ErrorHandlers) Handle(ctx context.Context, err error) {
	for _, h := range e {
		h.Handle(ctx, err)
	}
}
