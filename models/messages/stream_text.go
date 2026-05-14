package messages

import (
	"encoding/json"

	"github.com/juggleim/imbot-sdk-go/models"
)

type StreamTextMessage struct {
	models.MessageContent `json:"-"`
	Content               string `json:"content,omitempty"`
	IsFinished            bool   `json:"is_finished"`
	Seq                   int    `json:"seq"`
}

func NewStreamTextMessage() *StreamTextMessage {
	return &StreamTextMessage{MessageContent: models.NewMessageContent(MessageContentTypeStreamText)}
}

func (msg *StreamTextMessage) Encode() ([]byte, error) {
	return json.Marshal(msg)
}

func (msg *StreamTextMessage) Decode(data []byte) error {
	msg.MessageContent = models.NewMessageContent(MessageContentTypeStreamText)
	return json.Unmarshal(data, msg)
}

func (msg *StreamTextMessage) ConversationDigest() string {
	return msg.Content
}

func (msg *StreamTextMessage) GetSearchContent() string {
	return msg.Content
}
