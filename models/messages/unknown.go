package messages

import "github.com/juggleim/imbot-sdk-go/models"

type UnknownMessage struct {
	models.MessageContent `json:"-"`
	MessageType           string
	Content               string
	Flags                 int
}

func NewUnknownMessage(messageType string) *UnknownMessage {
	return &UnknownMessage{
		MessageContent: models.NewMessageContent(models.MessageContentTypeUnknown),
		MessageType:    messageType,
	}
}

func (msg *UnknownMessage) Encode() ([]byte, error) {
	return []byte(msg.Content), nil
}

func (msg *UnknownMessage) Decode(data []byte) error {
	msg.Content = string(data)
	return nil
}

func (msg *UnknownMessage) GetContentType() string {
	if msg.MessageType != "" {
		return msg.MessageType
	}
	return models.MessageContentTypeUnknown
}

func (msg *UnknownMessage) GetFlags() int {
	return msg.Flags
}
