package messages

import (
	"encoding/json"

	"github.com/juggleim/imbot-sdk-go/models"
)

type ImageMessage struct {
	models.MediaMessageContent `json:"-"`
	Url                        string `json:"url,omitempty"`
	LocalPath                  string `json:"local,omitempty"`
	ThumbnailUrl               string `json:"thumbnail,omitempty"`
	ThumbnailLocalPath         string `json:"thumbnailLocalPath,omitempty"`
	Height                     int    `json:"height"`
	Width                      int    `json:"width"`
	Extra                      string `json:"extra,omitempty"`
	Size                       int64  `json:"size"`
}

func NewImageMessage() *ImageMessage {
	return &ImageMessage{MediaMessageContent: newMediaMessageContent(MessageContentTypeImage)}
}

func (msg *ImageMessage) Encode() ([]byte, error) {
	return json.Marshal(msg)
}

func (msg *ImageMessage) Decode(data []byte) error {
	msg.MediaMessageContent = newMediaMessageContent(MessageContentTypeImage)
	return json.Unmarshal(data, msg)
}

func (msg *ImageMessage) ConversationDigest() string {
	return "[Image]"
}
