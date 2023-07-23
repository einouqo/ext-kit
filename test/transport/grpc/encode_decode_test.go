package grpc

import (
	"testing"

	"github.com/einouqo/ext-kit/test/service"
	"github.com/einouqo/ext-kit/transport/grpc"
)

func TestEncodeDecodeImplementation(*testing.T) {
	var (
		_ grpc.DecodeFunc[service.EchoRequest]  = decodeRequest
		_ grpc.DecodeFunc[service.EchoResponse] = decodeResponse
		_ grpc.EncodeFunc[service.EchoRequest]  = encodeRequest
		_ grpc.EncodeFunc[service.EchoResponse] = encodeResponse
	)
}
