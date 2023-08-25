package ws

import (
	"testing"

	"github.com/einouqo/ext-kit/test/transport/_service"
	"github.com/einouqo/ext-kit/transport/ws"
)

func TestEncodeDecodeImplementation(*testing.T) {
	var (
		_ ws.DecodeFunc[service.EchoRequest]  = decodeRequest
		_ ws.DecodeFunc[service.EchoResponse] = decodeResponse
		_ ws.EncodeFunc[service.EchoRequest]  = encodeRequest
		_ ws.EncodeFunc[service.EchoResponse] = encodeResponse
	)
}
