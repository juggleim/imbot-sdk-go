package messages

import (
	"encoding/json"

	"github.com/juggleim/imbot-sdk-go/models"
)

type SnapshotPackedVideoMessage struct {
	models.MediaMessageContent `json:"-"`
	SightUrl                   string `json:"sightUrl,omitempty"`
	LocalPath                  string `json:"local,omitempty"`
	Content                    string `json:"content,omitempty"`
	Height                     int    `json:"height"`
	Width                      int    `json:"width"`
	Size                       int64  `json:"size"`
	Duration                   int    `json:"duration"`
	Name                       string `json:"name,omitempty"`
	Extra                      string `json:"extra,omitempty"`
}

func NewSnapshotPackedVideoMessage() *SnapshotPackedVideoMessage {
	return &SnapshotPackedVideoMessage{MediaMessageContent: newMediaMessageContent(MessageContentTypeSnapshotPackedVideo)}
}

func (msg *SnapshotPackedVideoMessage) Encode() ([]byte, error) {
	return json.Marshal(msg)
}

func (msg *SnapshotPackedVideoMessage) Decode(data []byte) error {
	msg.MediaMessageContent = newMediaMessageContent(MessageContentTypeSnapshotPackedVideo)
	return json.Unmarshal(data, msg)
}

func (msg *SnapshotPackedVideoMessage) ConversationDigest() string {
	return "[Video]"
}
