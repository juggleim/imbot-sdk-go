package messages

import (
	"encoding/json"

	"github.com/juggleim/imbot-sdk-go/models"
)

type VoiceMessage struct {
	models.MediaMessageContent `json:"-"`
	Url                        string `json:"url,omitempty"`
	LocalPath                  string `json:"local,omitempty"`
	Duration                   int    `json:"duration"`
	Extra                      string `json:"extra,omitempty"`
}

func NewVoiceMessage() *VoiceMessage {
	return &VoiceMessage{MediaMessageContent: newMediaMessageContent(MessageContentTypeVoice)}
}

func (msg *VoiceMessage) Encode() ([]byte, error) {
	return json.Marshal(msg)
}

func (msg *VoiceMessage) Decode(data []byte) error {
	msg.MediaMessageContent = newMediaMessageContent(MessageContentTypeVoice)
	return json.Unmarshal(data, msg)
}

func (msg *VoiceMessage) ConversationDigest() string {
	return "[Voice]"
}
