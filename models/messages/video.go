package messages

import (
	"encoding/json"

	"github.com/juggleim/imbot-sdk-go/models"
)

type VideoMessage struct {
	models.MediaMessageContent `json:"-"`
	Url                        string `json:"url,omitempty"`
	LocalPath                  string `json:"local,omitempty"`
	SnapshotUrl                string `json:"poster,omitempty"`
	SnapshotLocalPath          string `json:"snapshotLocalPath,omitempty"`
	Height                     int    `json:"height"`
	Width                      int    `json:"width"`
	Size                       int64  `json:"size"`
	Duration                   int    `json:"duration"`
	Extra                      string `json:"extra,omitempty"`
}

func NewVideoMessage() *VideoMessage {
	return &VideoMessage{MediaMessageContent: newMediaMessageContent(MessageContentTypeVideo)}
}

func (msg *VideoMessage) Encode() ([]byte, error) {
	return json.Marshal(msg)
}

func (msg *VideoMessage) Decode(data []byte) error {
	msg.MediaMessageContent = newMediaMessageContent(MessageContentTypeVideo)
	return json.Unmarshal(data, msg)
}

func (msg *VideoMessage) ConversationDigest() string {
	return "[Video]"
}
