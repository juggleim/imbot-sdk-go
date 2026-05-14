package models

import (
	"encoding/json"

	"github.com/juggleim/imbot-sdk-go/imbotclients/pbdefines/pbobjs"
)

const (
	MessageContentTypeUnknown = "jg:unknown"
)

const (
	MessageFlagNone        = 0
	MessageFlagIsCmd       = 1
	MessageFlagIsCountable = 2
	MessageFlagIsStatus    = 4
	MessageFlagIsSave      = 8
	MessageFlagIsModified  = 16
	MessageFlagIsMerged    = 32
	MessageFlagIsMute      = 64
	MessageFlagIsBroadcast = 128

	DefaultMessageFlags = MessageFlagIsCountable | MessageFlagIsSave
)

type MessageContentInterface interface {
	GetContentType() string
	Encode() ([]byte, error)
	Decode(data []byte) error
	ConversationDigest() string
	GetFlags() int
	GetSearchContent() string
}

type MessageContent struct {
	MsgType string `json:"-"`
}

func NewMessageContent(msgType string) MessageContent {
	if msgType == "" {
		msgType = MessageContentTypeUnknown
	}
	return MessageContent{MsgType: msgType}
}

func (content *MessageContent) GetContentType() string {
	if content == nil || content.MsgType == "" {
		return MessageContentTypeUnknown
	}
	return content.MsgType
}

func (content *MessageContent) Encode() ([]byte, error) {
	return json.Marshal(struct{}{})
}

func (content *MessageContent) Decode(data []byte) error {
	return nil
}

func (content *MessageContent) ConversationDigest() string {
	return ""
}

func (content *MessageContent) GetFlags() int {
	return DefaultMessageFlags
}

func (content *MessageContent) GetSearchContent() string {
	return ""
}

type MediaMessageContent struct {
	MessageContent `json:"-"`
	LocalPath      string `json:"local,omitempty"`
	Url            string `json:"url,omitempty"`
}

type MessageMentionInfo struct {
	MentionType   pbobjs.MentionType
	TargetUserIds []string
}

type Message struct {
	Conversation   *Conversation
	MsgId          string
	HasRead        bool
	MsgTime        int64
	SenderId       string
	MsgType        string
	MsgContent     MessageContentInterface
	GroupId        string
	ReferedMessage *Message
	MentionInfo    *MessageMentionInfo
}

type MessageReaction struct {
	MessageId string
	ItemList  []*MessageReactionItem
}

type MessageReactionItem struct {
	ReactionId   string
	UserInfoList []*UserInfo
}
