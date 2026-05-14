package messages

import (
	"encoding/json"

	"github.com/juggleim/imbot-sdk-go/imbotclients/pbdefines/pbobjs"
	"github.com/juggleim/imbot-sdk-go/models"
)

type MergeMessage struct {
	models.MessageContent `json:"-"`
	Title                 string
	ContainerMsgId        string
	Conversation          *models.Conversation
	MessageIdList         []string
	PreviewList           []*MergeMessagePreviewUnit
	Extra                 string
}

type MergeMessagePreviewUnit struct {
	PreviewContent string
	Sender         *models.UserInfo
}

func NewMergeMessage(title string, conversation *models.Conversation, messageIdList []string, previewList []*MergeMessagePreviewUnit) *MergeMessage {
	if len(messageIdList) > 100 {
		messageIdList = messageIdList[:100]
	}
	if len(previewList) > 10 {
		previewList = previewList[:10]
	}
	return &MergeMessage{
		MessageContent: models.NewMessageContent(MessageContentTypeMerge),
		Title:          title,
		Conversation:   conversation,
		MessageIdList:  messageIdList,
		PreviewList:    previewList,
	}
}

func (msg *MergeMessage) Encode() ([]byte, error) {
	dto := mergeMessageDTO{
		Title:          msg.Title,
		ContainerMsgId: msg.ContainerMsgId,
		MessageIdList:  msg.MessageIdList,
		Extra:          msg.Extra,
	}
	if msg.Conversation != nil {
		dto.ConversationId = msg.Conversation.Conversation
		dto.ConversationType = int32(msg.Conversation.ConversationType)
		dto.SubChannel = msg.Conversation.SubChannel
	}
	for _, preview := range msg.PreviewList {
		unit := mergePreviewDTO{Content: preview.PreviewContent}
		if preview.Sender != nil {
			unit.UserId = preview.Sender.UserId
			unit.UserName = preview.Sender.UserName
			unit.Portrait = preview.Sender.UserPortrait
		}
		dto.PreviewList = append(dto.PreviewList, unit)
	}
	return json.Marshal(dto)
}

func (msg *MergeMessage) Decode(data []byte) error {
	msg.MessageContent = models.NewMessageContent(MessageContentTypeMerge)
	var dto mergeMessageDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return err
	}
	msg.Title = dto.Title
	msg.ContainerMsgId = dto.ContainerMsgId
	msg.MessageIdList = dto.MessageIdList
	msg.Extra = dto.Extra
	if dto.ConversationId != "" {
		msg.Conversation = &models.Conversation{
			Conversation:     dto.ConversationId,
			ConversationType: pbobjs.ChannelType(dto.ConversationType),
			SubChannel:       dto.SubChannel,
		}
	}
	msg.PreviewList = nil
	for _, unit := range dto.PreviewList {
		msg.PreviewList = append(msg.PreviewList, &MergeMessagePreviewUnit{
			PreviewContent: unit.Content,
			Sender: &models.UserInfo{
				UserId:       unit.UserId,
				UserName:     unit.UserName,
				UserPortrait: unit.Portrait,
			},
		})
	}
	return nil
}

func (msg *MergeMessage) ConversationDigest() string {
	return "[Merge]"
}

type mergeMessageDTO struct {
	Title            string            `json:"title,omitempty"`
	ContainerMsgId   string            `json:"containerMsgId,omitempty"`
	ConversationId   string            `json:"conversationId,omitempty"`
	ConversationType int32             `json:"conversationType,omitempty"`
	SubChannel       string            `json:"sub_channel,omitempty"`
	MessageIdList    []string          `json:"messageIdList,omitempty"`
	PreviewList      []mergePreviewDTO `json:"previewList,omitempty"`
	Extra            string            `json:"extra,omitempty"`
}

type mergePreviewDTO struct {
	Content  string `json:"content,omitempty"`
	UserId   string `json:"userId,omitempty"`
	UserName string `json:"userName,omitempty"`
	Portrait string `json:"portrait,omitempty"`
}
