package messages

import (
	"encoding/json"

	"github.com/juggleim/imbot-sdk-go/models"
)

type FileMessage struct {
	models.MediaMessageContent `json:"-"`
	Name                       string `json:"name,omitempty"`
	Url                        string `json:"url,omitempty"`
	LocalPath                  string `json:"local,omitempty"`
	Size                       int64  `json:"size"`
	Type                       string `json:"type,omitempty"`
	Extra                      string `json:"extra,omitempty"`
}

func NewFileMessage() *FileMessage {
	return &FileMessage{MediaMessageContent: newMediaMessageContent(MessageContentTypeFile)}
}

func (msg *FileMessage) Encode() ([]byte, error) {
	return json.Marshal(msg)
}

func (msg *FileMessage) Decode(data []byte) error {
	msg.MediaMessageContent = newMediaMessageContent(MessageContentTypeFile)
	return json.Unmarshal(data, msg)
}

func (msg *FileMessage) ConversationDigest() string {
	return "[File]"
}

func (msg *FileMessage) GetSearchContent() string {
	return msg.Name
}
