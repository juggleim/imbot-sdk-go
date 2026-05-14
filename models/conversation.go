package models

import "github.com/juggleim/imbot-sdk-go/imbotclients/pbdefines/pbobjs"

type ConversationInfo struct {
	Conversation  *Conversation
	LatestMessage *Message
	UnreadCount   int
	SortTime      int64
	IsTop         bool
	Mute          bool
	Draft         string
	MentionInfo   *ConversationMentionInfo
	DisplayName   string
	Alias         string
	Portrait      string
}

type Conversation struct {
	Conversation     string
	SubChannel       string
	ConversationType pbobjs.ChannelType
}

type ConversationMentionInfo struct {
	SenderId    string
	MsgId       string
	MsgTime     int64
	MentionType pbobjs.MentionType
}
