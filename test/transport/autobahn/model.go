package main

import "github.com/einouqo/ext-kit/transport/ws"

type Message struct {
	Type    ws.MessageType
	Payload []byte
}
