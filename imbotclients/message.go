package imbotclients

import (
	"github.com/juggleim/imbot-sdk-go/imbotclients/pbdefines/pbobjs"
	"github.com/juggleim/imbot-sdk-go/models"
	"github.com/juggleim/imbot-sdk-go/models/messages"
	"github.com/juggleim/imbot-sdk-go/utils"
	"google.golang.org/protobuf/proto"
)

// SendMessage 按会话类型选择 topic 并 Publish（p_msg / g_msg / c_msg / pc_msg），
// 与 Android PBData.sendMessageData 一致；targetId 为 conver.Conversation；UpMsg 中的 SubChannel 等由调用方填入。
func (client *ImBotClient) SendMessage(conver *models.Conversation, upMsg *pbobjs.UpMsg) (utils.ClientErrorCode, *pbobjs.PublishAckMsgBody) {
	if conver == nil || upMsg == nil || conver.Conversation == "" {
		return utils.ClientErrorCode_Unknown, nil
	}
	var topic string
	switch conver.ConversationType {
	case pbobjs.ChannelType_Private:
		topic = "p_msg"
	case pbobjs.ChannelType_Group:
		topic = "g_msg"
	case pbobjs.ChannelType_Chatroom:
		topic = "c_msg"
	case pbobjs.ChannelType_PublicChannel:
		topic = "pc_msg"
	default:
		return utils.ClientErrorCode_Unknown, nil
	}
	data, err := proto.Marshal(upMsg)
	if err != nil {
		return utils.ClientErrorCode_Unknown, nil
	}
	return client.Publish(topic, conver.Conversation, data)
}

type IMessageListener interface {
	OnMessageReceive(msg *models.Message)
	OnMessageRecall(msg *models.Message)
	OnMessageUpdate(msg *models.Message)
	OnMessageDelete(conver *models.Conversation, msgIds []string)
	OnMessageClear(conver *models.Conversation, time int64, senderId string)
	OnMessageReactionAdd(conver *models.Conversation, reaction *models.MessageReaction)
	OnMessageReactionRemove(conver *models.Conversation, reaction *models.MessageReaction)
	OnMessageSetTop(message *models.Message, operatorId string, isTop bool)
}

func (client *ImBotClient) SyncMsgs(req *pbobjs.SyncMsgReq) (utils.ClientErrorCode, *pbobjs.DownMsgSet) {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("sync_msgs", client.UserId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.DownMsgSet{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return utils.ClientErrorCode_Unknown, nil
}

func (client *ImBotClient) AddMsgExset(req *pbobjs.MsgExt) utils.ClientErrorCode {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("msg_exset", req.MsgId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil {
		if qryAck.Code == imErrorCodeSuccess {
			return utils.ClientErrorCode_Success
		}
		return utils.ClientErrorCode(qryAck.Code)
	}
	return code
}

func (client *ImBotClient) AddMessageListener(listener IMessageListener) {
	if listener == nil {
		return
	}
	client.messageListeners = append(client.messageListeners, listener)
}

func (client *ImBotClient) notifyMessageReceive(downMsg *pbobjs.DownMsg) *models.Message {
	msg := client.downMsgToMessage(downMsg)
	if msg == nil {
		return nil
	}
	for _, listener := range client.messageListeners {
		if listener != nil {
			listener.OnMessageReceive(msg)
		}
	}
	return msg
}

func (client *ImBotClient) notifyMessageRecall(msg *models.Message) {
	if msg == nil {
		return
	}
	for _, listener := range client.messageListeners {
		if listener != nil {
			listener.OnMessageRecall(msg)
		}
	}
}

func (client *ImBotClient) notifyMessageUpdate(msg *models.Message) {
	if msg == nil {
		return
	}
	for _, listener := range client.messageListeners {
		if listener != nil {
			listener.OnMessageUpdate(msg)
		}
	}
}

func (client *ImBotClient) notifyMessageDelete(conver *models.Conversation, msgIds []string) {
	if conver == nil || len(msgIds) == 0 {
		return
	}
	for _, listener := range client.messageListeners {
		if listener != nil {
			listener.OnMessageDelete(conver, msgIds)
		}
	}
}

func (client *ImBotClient) notifyMessageClear(conver *models.Conversation, timestamp int64, senderId string) {
	if conver == nil {
		return
	}
	for _, listener := range client.messageListeners {
		if listener != nil {
			listener.OnMessageClear(conver, timestamp, senderId)
		}
	}
}

func (client *ImBotClient) notifyMessageReactionAdd(conver *models.Conversation, reaction *models.MessageReaction) {
	if conver == nil || reaction == nil {
		return
	}
	for _, listener := range client.messageListeners {
		if listener != nil {
			listener.OnMessageReactionAdd(conver, reaction)
		}
	}
}

func (client *ImBotClient) notifyMessageReactionRemove(conver *models.Conversation, reaction *models.MessageReaction) {
	if conver == nil || reaction == nil {
		return
	}
	for _, listener := range client.messageListeners {
		if listener != nil {
			listener.OnMessageReactionRemove(conver, reaction)
		}
	}
}

func (client *ImBotClient) notifyMessageSetTop(msg *models.Message, operatorId string, isTop bool) {
	if msg == nil {
		return
	}
	for _, listener := range client.messageListeners {
		if listener != nil {
			listener.OnMessageSetTop(msg, operatorId, isTop)
		}
	}
}

func (client *ImBotClient) downMsgToMessage(downMsg *pbobjs.DownMsg) *models.Message {
	if downMsg == nil {
		return nil
	}
	msg := &models.Message{
		Conversation:   conversationFromDownMsg(downMsg),
		MsgId:          downMsg.MsgId,
		HasRead:        downMsg.IsRead,
		MsgTime:        downMsg.MsgTime,
		SenderId:       downMsg.SenderId,
		MsgType:        downMsg.MsgType,
		MsgContent:     decodeMessageContent(downMsg.MsgType, downMsg.MsgContent),
		GroupId:        groupIdFromDownMsg(downMsg),
		ReferedMessage: client.downMsgToMessage(downMsg.ReferMsg),
		MentionInfo:    mentionInfoFromPB(downMsg.MentionInfo),
	}
	return msg
}

func decodeMessageContent(msgType string, data []byte) models.MessageContentInterface {
	var content models.MessageContentInterface
	switch msgType {
	case messages.MessageContentTypeText:
		content = messages.NewTextMessage("")
	case messages.MessageContentTypeImage:
		content = messages.NewImageMessage()
	case messages.MessageContentTypeFile:
		content = messages.NewFileMessage()
	case messages.MessageContentTypeVideo:
		content = messages.NewVideoMessage()
	case messages.MessageContentTypeVoice:
		content = messages.NewVoiceMessage()
	case messages.MessageContentTypeStreamText:
		content = messages.NewStreamTextMessage()
	case messages.MessageContentTypeRecallInfo:
		content = messages.NewRecallInfoMessage()
	case messages.MessageContentTypeMerge:
		content = messages.NewMergeMessage("", nil, nil, nil)
	case messages.MessageContentTypeThumbnailPackedImage:
		content = messages.NewThumbnailPackedImageMessage()
	case messages.MessageContentTypeSnapshotPackedVideo:
		content = messages.NewSnapshotPackedVideoMessage()
	default:
		unknown := messages.NewUnknownMessage(msgType)
		unknown.Content = string(data)
		return unknown
	}
	if err := content.Decode(data); err != nil {
		unknown := messages.NewUnknownMessage(msgType)
		unknown.Content = string(data)
		return unknown
	}
	return content
}

func conversationFromDownMsg(downMsg *pbobjs.DownMsg) *models.Conversation {
	if downMsg == nil {
		return nil
	}
	return &models.Conversation{
		Conversation:     downMsg.TargetId,
		ConversationType: downMsg.ChannelType,
		SubChannel:       downMsg.SubChannel,
	}
}

func mentionInfoFromPB(info *pbobjs.MentionInfo) *models.MessageMentionInfo {
	if info == nil {
		return nil
	}
	targetUserIds := make([]string, 0, len(info.TargetUsers))
	for _, user := range info.TargetUsers {
		if user != nil && user.UserId != "" {
			targetUserIds = append(targetUserIds, user.UserId)
		}
	}
	return &models.MessageMentionInfo{
		MentionType:   info.MentionType,
		TargetUserIds: targetUserIds,
	}
}

func groupIdFromDownMsg(downMsg *pbobjs.DownMsg) string {
	if downMsg == nil {
		return ""
	}
	if downMsg.ChannelType == pbobjs.ChannelType_Group {
		return downMsg.TargetId
	}
	if downMsg.GroupInfo != nil {
		return downMsg.GroupInfo.GroupId
	}
	return ""
}
