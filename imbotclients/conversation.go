package imbotclients

import (
	"fmt"

	"github.com/juggleim/imbot-sdk-go/imbotclients/pbdefines/pbobjs"
	"github.com/juggleim/imbot-sdk-go/models"
	"github.com/juggleim/imbot-sdk-go/utils"
	"google.golang.org/protobuf/proto"
)

type IConversationChangeListener interface {
	OnConversationInfoAdd([]*models.ConversationInfo)
	OnConversationInfoUpdate([]*models.ConversationInfo)
	OnConversationInfoDelete([]*models.ConversationInfo)
	OnTotalUnreadMessageCountUpdate(count int)
}

func (client *ImBotClient) AddConversationChangeListener(listener IConversationChangeListener) {
	if listener == nil {
		return
	}
	client.converChgListeners = append(client.converChgListeners, listener)
}

// GetConversation 对应 Android 侧拉取单会话（WebSocket topic: qry_conver）。
func (client *ImBotClient) GetConversation(req *pbobjs.QryConverReq) (utils.ClientErrorCode, *pbobjs.Conversation) {
	data, _ := proto.Marshal(req)
	code, ack := client.Query("qry_conver", client.UserId, data)
	if code == utils.ClientErrorCode_Success && ack != nil && ack.Code == imErrorCodeSuccess {
		resp := &pbobjs.Conversation{}
		if err := proto.Unmarshal(ack.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return code, nil
}

// GetTotalUnreadCount 对应 IConversationManager 中服务端查询总未读（topic: qry_total_unread_count）。
func (client *ImBotClient) GetTotalUnreadCount(req *pbobjs.QryTotalUnreadCountReq) (utils.ClientErrorCode, *pbobjs.QryTotalUnreadCountResp) {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("qry_total_unread_count", client.UserId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.QryTotalUnreadCountResp{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return code, nil
}

// GetConversations 分页查询会话列表（topic: qry_convers）。
func (client *ImBotClient) GetConversations(req *pbobjs.QryConversationsReq) (utils.ClientErrorCode, *pbobjs.QryConversationsResp) {
	return client.queryConversations("qry_convers", req)
}

// GetPcConversations 查询 PC 会话列表（topic: qry_pc_convers）。
func (client *ImBotClient) GetPcConversations(req *pbobjs.QryConversationsReq) (utils.ClientErrorCode, *pbobjs.QryConversationsResp) {
	return client.queryConversations("qry_pc_convers", req)
}

// GetTopConversations 查询置顶会话（topic: qry_top_convers）。
func (client *ImBotClient) GetTopConversations(req *pbobjs.QryTopConversReq) (utils.ClientErrorCode, *pbobjs.QryConversationsResp) {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("qry_top_convers", client.UserId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.QryConversationsResp{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return code, nil
}

// ClearUnreadCount 对应 clear_unread。
func (client *ImBotClient) ClearUnreadCount(req *pbobjs.ClearUnreadReq) utils.ClientErrorCode {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("clear_unread", client.UserId, data)
	return ackCode(code, qryAck)
}

// ClearTotalUnreadCount 对应 clear_total_unread（与 Android PBData.clearTotalUnreadCountData 一致，body 为 QryTotalUnreadCountReq）。
func (client *ImBotClient) ClearTotalUnreadCount(req *pbobjs.QryTotalUnreadCountReq) utils.ClientErrorCode {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("clear_total_unread", client.UserId, data)
	return ackCode(code, qryAck)
}

// GetMentionMsgs 对应 qry_mention_msgs。
func (client *ImBotClient) GetMentionMsgs(req *pbobjs.QryMentionMsgsReq) (utils.ClientErrorCode, *pbobjs.QryMentionMsgsResp) {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("qry_mention_msgs", client.UserId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.QryMentionMsgsResp{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return code, nil
}

// SyncConversations 对应 sync_convers。
func (client *ImBotClient) SyncConversations(req *pbobjs.QryConversationsReq) (utils.ClientErrorCode, *pbobjs.QryConversationsResp) {
	return client.queryConversations("sync_convers", req)
}

// SetMute 对应会话免打扰 / mute（WebSocket: undisturb_convers，与 Android setMute -> disturbData 一致）。
func (client *ImBotClient) SetMute(req *pbobjs.UndisturbConversReq) utils.ClientErrorCode {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("undisturb_convers", client.UserId, data)
	return ackCode(code, qryAck)
}

// SetConversationTop 对应 top_convers（与 Android setConversationTop 一致）。
func (client *ImBotClient) SetConversationTop(req *pbobjs.ConversationsReq) utils.ClientErrorCode {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("top_convers", client.UserId, data)
	return ackCode(code, qryAck)
}

// DeleteConversations 对应 del_convers。
func (client *ImBotClient) DeleteConversations(req *pbobjs.ConversationsReq) utils.ClientErrorCode {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("del_convers", client.UserId, data)
	return ackCode(code, qryAck)
}

// SetUnread 对应 mark_unread。
func (client *ImBotClient) SetUnread(req *pbobjs.ConversationsReq) utils.ClientErrorCode {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("mark_unread", client.UserId, data)
	return ackCode(code, qryAck)
}

// CreateConversationInfo 对应 add_conver。
func (client *ImBotClient) CreateConversationInfo(req *pbobjs.Conversation) utils.ClientErrorCode {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("add_conver", client.UserId, data)
	return ackCode(code, qryAck)
}

// CreateConversationTag 对应 create_user_conver_tags。
func (client *ImBotClient) CreateConversationTag(tagId, tagName string) utils.ClientErrorCode {
	req := &pbobjs.UserConverTags{
		Tags: []*pbobjs.ConverTag{
			{Tag: tagId, TagName: tagName},
		},
	}
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("create_user_conver_tags", client.UserId, data)
	return ackCode(code, qryAck)
}

// DestroyConversationTag 对应 del_user_conver_tags。
func (client *ImBotClient) DestroyConversationTag(tagId string) utils.ClientErrorCode {
	req := &pbobjs.UserConverTags{
		Tags: []*pbobjs.ConverTag{{Tag: tagId}},
	}
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("del_user_conver_tags", client.UserId, data)
	return ackCode(code, qryAck)
}

// UpdateConversationTagName 与 Android 一致：复用 create_user_conver_tags 携带新名称。
func (client *ImBotClient) UpdateConversationTagName(tagId, tagName string) utils.ClientErrorCode {
	return client.CreateConversationTag(tagId, tagName)
}

// GetConversationTagList 对应 qry_user_conver_tags。
func (client *ImBotClient) GetConversationTagList() (utils.ClientErrorCode, *pbobjs.UserConverTags) {
	code, qryAck := client.Query("qry_user_conver_tags", client.UserId, nil)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.UserConverTags{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return code, nil
}

// AddConversationsToTag 对应 tag_add_convers。
func (client *ImBotClient) AddConversationsToTag(req *pbobjs.TagConvers) utils.ClientErrorCode {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("tag_add_convers", client.UserId, data)
	return ackCode(code, qryAck)
}

// RemoveConversationsFromTag 对应 tag_del_convers。
func (client *ImBotClient) RemoveConversationsFromTag(req *pbobjs.TagConvers) utils.ClientErrorCode {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("tag_del_convers", client.UserId, data)
	return ackCode(code, qryAck)
}

func (client *ImBotClient) queryConversations(method string, req *pbobjs.QryConversationsReq) (utils.ClientErrorCode, *pbobjs.QryConversationsResp) {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query(method, client.UserId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.QryConversationsResp{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return code, nil
}

func (client *ImBotClient) notifyConversationForMessage(msg *models.Message, downMsg *pbobjs.DownMsg) {
	if msg == nil || msg.Conversation == nil {
		return
	}
	info := conversationInfoFromMessage(msg, downMsg)
	key := conversationKey(msg.Conversation)
	if _, loaded := client.converCache.LoadOrStore(key, info); loaded {
		client.converCache.Store(key, info)
		client.notifyConversationInfoUpdate([]*models.ConversationInfo{info})
	} else {
		client.notifyConversationInfoAdd([]*models.ConversationInfo{info})
	}
	if downMsg != nil && !downMsg.IsSend && !downMsg.IsRead {
		client.totalUnreadCount++
		client.notifyTotalUnreadMessageCountUpdate(client.totalUnreadCount)
	}
}

func (client *ImBotClient) notifyConversationInfoAdd(convers []*models.ConversationInfo) {
	if len(convers) == 0 {
		return
	}
	for _, listener := range client.converChgListeners {
		if listener != nil {
			listener.OnConversationInfoAdd(convers)
		}
	}
}

func (client *ImBotClient) notifyConversationInfoUpdate(convers []*models.ConversationInfo) {
	if len(convers) == 0 {
		return
	}
	for _, listener := range client.converChgListeners {
		if listener != nil {
			listener.OnConversationInfoUpdate(convers)
		}
	}
}

func (client *ImBotClient) notifyConversationInfoDelete(convers []*models.ConversationInfo) {
	if len(convers) == 0 {
		return
	}
	for _, listener := range client.converChgListeners {
		if listener != nil {
			listener.OnConversationInfoDelete(convers)
		}
	}
}

func (client *ImBotClient) notifyTotalUnreadMessageCountUpdate(count int) {
	for _, listener := range client.converChgListeners {
		if listener != nil {
			listener.OnTotalUnreadMessageCountUpdate(count)
		}
	}
}

func conversationInfoFromMessage(msg *models.Message, downMsg *pbobjs.DownMsg) *models.ConversationInfo {
	info := &models.ConversationInfo{
		Conversation:  msg.Conversation,
		LatestMessage: msg,
		SortTime:      msg.MsgTime,
	}
	if downMsg == nil {
		return info
	}
	info.UnreadCount = unreadCountFromDownMsg(downMsg)
	if downMsg.MentionInfo != nil {
		info.MentionInfo = &models.ConversationMentionInfo{
			SenderId:    downMsg.SenderId,
			MsgId:       downMsg.MsgId,
			MsgTime:     downMsg.MsgTime,
			MentionType: downMsg.MentionInfo.MentionType,
		}
	}
	if downMsg.TargetUserInfo != nil {
		info.DisplayName = downMsg.TargetUserInfo.Nickname
		info.Portrait = downMsg.TargetUserInfo.UserPortrait
	}
	if downMsg.GroupInfo != nil {
		info.DisplayName = downMsg.GroupInfo.GroupName
		info.Portrait = downMsg.GroupInfo.GroupPortrait
		info.Mute = downMsg.GroupInfo.IsMute > 0
	}
	return info
}

func unreadCountFromDownMsg(downMsg *pbobjs.DownMsg) int {
	if downMsg == nil || downMsg.IsSend || downMsg.IsRead {
		return 0
	}
	return 1
}

func conversationKey(conver *models.Conversation) string {
	if conver == nil {
		return ""
	}
	return fmt.Sprintf("%d:%s:%s", conver.ConversationType, conver.Conversation, conver.SubChannel)
}
