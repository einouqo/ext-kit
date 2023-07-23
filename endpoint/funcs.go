package endpoint

import "errors"

var (
	StreamDone = errors.New("stream done")
)

type Receive[T any] func() (T, error)

type Stop = func()
