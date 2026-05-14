package messages

import (
	"encoding/json"

	"github.com/juggleim/imbot-sdk-go/models"
)

type ThumbnailPackedImageMessage struct {
	models.MediaMessageContent `json:"-"`
	ImageUri                   string `json:"imageUri,omitempty"`
	LocalPath                  string `json:"local,omitempty"`
	Content                    string `json:"content,omitempty"`
	Height                     int    `json:"height"`
	Width                      int    `json:"width"`
	Extra                      string `json:"extra,omitempty"`
	Size                       int64  `json:"size"`
}

func NewThumbnailPackedImageMessage() *ThumbnailPackedImageMessage {
	return &ThumbnailPackedImageMessage{MediaMessageContent: newMediaMessageContent(MessageContentTypeThumbnailPackedImage)}
}

func (msg *ThumbnailPackedImageMessage) Encode() ([]byte, error) {
	return json.Marshal(msg)
}

func (msg *ThumbnailPackedImageMessage) Decode(data []byte) error {
	msg.MediaMessageContent = newMediaMessageContent(MessageContentTypeThumbnailPackedImage)
	return json.Unmarshal(data, msg)
}

func (msg *ThumbnailPackedImageMessage) ConversationDigest() string {
	return "[Image]"
}
