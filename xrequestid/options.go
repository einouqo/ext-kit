package xrequestid

type Option func(*Handler)

func WithGenerator(gen Generator) Option {
	return func(handler *Handler) {
		handler.gen = gen
	}
}
