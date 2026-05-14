package messages

import (
	"encoding/json"

	"github.com/juggleim/imbot-sdk-go/models"
)

type TextMessage struct {
	models.MessageContent `json:"-"`
	Content               string `json:"content,omitempty"`
	Extra                 string `json:"extra,omitempty"`
}

func NewTextMessage(content string) *TextMessage {
	return &TextMessage{
		MessageContent: models.NewMessageContent(MessageContentTypeText),
		Content:        content,
	}
}

func (msg *TextMessage) Encode() ([]byte, error) {
	return json.Marshal(msg)
}

func (msg *TextMessage) Decode(data []byte) error {
	msg.MessageContent = models.NewMessageContent(MessageContentTypeText)
	return json.Unmarshal(data, msg)
}

func (msg *TextMessage) ConversationDigest() string {
	return msg.Content
}

func (msg *TextMessage) GetSearchContent() string {
	return msg.Content
}
