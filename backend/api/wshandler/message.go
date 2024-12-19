package wshandler

type wsRequestMessageType int
type wsResponseMessageType int

const (
	wsRequestMessageInvalid wsRequestMessageType = iota
	wsRequestMessageAwareness
)

const (
	wsResponseMessageInvalid wsResponseMessageType = iota
	wsResponseMessageAwareness
)

type WSRequestMessage struct {
	Type    wsRequestMessageType `json:"t"`
	Message any                  `json:"m"`
}

type WSResponseMessage struct {
	Type    wsResponseMessageType `json:"t"`
	Message any                   `json:"m"`
}
