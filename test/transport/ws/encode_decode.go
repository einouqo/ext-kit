package ws

import (
	"context"
	"encoding/json"

	"github.com/einouqo/ext-kit/test/transport/_service"
	"github.com/einouqo/ext-kit/transport/ws"
)

func encodeRequest(_ context.Context, req service.EchoRequest) ([]byte, ws.MessageType, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, 0, err
	}
	return data, ws.TextMessageType, nil
}

func decodeRequest(_ context.Context, _ ws.MessageType, data []byte) (service.EchoRequest, error) {
	var req service.EchoRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return service.EchoRequest{}, err
	}
	return req, nil
}

func encodeResponse(_ context.Context, resp service.EchoResponse) ([]byte, ws.MessageType, error) {
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, 0, err
	}
	return data, ws.TextMessageType, nil
}

func decodeResponse(_ context.Context, _ ws.MessageType, data []byte) (service.EchoResponse, error) {
	var resp service.EchoResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return service.EchoResponse{}, err
	}
	return resp, nil
}
