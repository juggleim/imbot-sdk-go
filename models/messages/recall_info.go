package messages

import (
	"encoding/json"

	"github.com/juggleim/imbot-sdk-go/models"
)

type RecallInfoMessage struct {
	models.MessageContent `json:"-"`
	Extra                 map[string]string `json:"exts,omitempty"`
}

func NewRecallInfoMessage() *RecallInfoMessage {
	return &RecallInfoMessage{MessageContent: models.NewMessageContent(MessageContentTypeRecallInfo)}
}

func (msg *RecallInfoMessage) Encode() ([]byte, error) {
	return json.Marshal(msg)
}

func (msg *RecallInfoMessage) Decode(data []byte) error {
	msg.MessageContent = models.NewMessageContent(MessageContentTypeRecallInfo)
	return json.Unmarshal(data, msg)
}
