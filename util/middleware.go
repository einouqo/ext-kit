package util

type Middleware[T any] func(T) T

func Chain[T any](outer Middleware[T], others ...Middleware[T]) Middleware[T] {
	return func(next T) T {
		for i := len(others) - 1; i >= 0; i-- { // reverse
			next = others[i](next)
		}
		return outer(next)
	}
}
