package xrequestid

import "github.com/lithammer/shortuuid/v4"

type ShortuuidGenerator struct{}

var (
	_ Generator = new(ShortuuidGenerator)
)

func (ShortuuidGenerator) Generate() XRequestID {
	return shortuuid.New()
}
