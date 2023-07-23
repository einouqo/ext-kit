package ws

import "github.com/fasthttp/websocket"

type MessageType int

const (
	TextMessageType   MessageType = websocket.TextMessage
	BinaryMessageType MessageType = websocket.BinaryMessage
)

func (mt MessageType) String() string {
	switch mt {
	case TextMessageType:
		return "text"
	case BinaryMessageType:
		return "binary"
	default:
		return "unknown"
	}
}

func (mt MessageType) fastsocket() int {
	return int(mt)
}

type CloseCode int

const (
	NormalClosureCloseCode           CloseCode = websocket.CloseNormalClosure
	GoingAwayCloseCode               CloseCode = websocket.CloseGoingAway
	ProtocolErrorCloseCode           CloseCode = websocket.CloseProtocolError
	UnsupportedDataCloseCode         CloseCode = websocket.CloseUnsupportedData
	NoStatusReceivedCloseCode        CloseCode = websocket.CloseNoStatusReceived
	AbnormalClosureCloseCode         CloseCode = websocket.CloseAbnormalClosure
	InvalidFramePayloadDataCloseCode CloseCode = websocket.CloseInvalidFramePayloadData
	PolicyViolationCloseCode         CloseCode = websocket.ClosePolicyViolation
	MessageTooBigCloseCode           CloseCode = websocket.CloseMessageTooBig
	MandatoryExtensionCloseCode      CloseCode = websocket.CloseMandatoryExtension
	InternalServerErrCloseCode       CloseCode = websocket.CloseInternalServerErr
	ServiceRestartCloseCode          CloseCode = websocket.CloseServiceRestart
	TryAgainLaterCloseCode           CloseCode = websocket.CloseTryAgainLater
	TLSHandshakeCloseCode            CloseCode = websocket.CloseTLSHandshake
)

func (c CloseCode) String() string {
	switch c {
	case NormalClosureCloseCode:
		return "normal_closure"
	case GoingAwayCloseCode:
		return "going_away"
	case ProtocolErrorCloseCode:
		return "protocol_error"
	case UnsupportedDataCloseCode:
		return "unsupported_data"
	case NoStatusReceivedCloseCode:
		return "no_status_received"
	case AbnormalClosureCloseCode:
		return "abnormal_closure"
	case InvalidFramePayloadDataCloseCode:
		return "invalid_frame_payload_data"
	case PolicyViolationCloseCode:
		return "policy_violation"
	case MessageTooBigCloseCode:
		return "message_too_big"
	case MandatoryExtensionCloseCode:
		return "mandatory_extension"
	case InternalServerErrCloseCode:
		return "internal_server_err"
	case ServiceRestartCloseCode:
		return "service_restart"
	case TryAgainLaterCloseCode:
		return "try_again_later"
	case TLSHandshakeCloseCode:
		return "tls_handshake"
	default:
		return "unknown"
	}
}

func (c CloseCode) fastsocket() int {
	return int(c)
}
