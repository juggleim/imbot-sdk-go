package imbotclients

import (
	"github.com/juggleim/imbot-sdk-go/imbotclients/pbdefines/pbobjs"
	"github.com/juggleim/imbot-sdk-go/models"
	"github.com/juggleim/imbot-sdk-go/utils"
	"google.golang.org/protobuf/proto"
)

// JoinChatroom 对应 IChatroomManager.joinChatroom(String chatroomId)。
func (client *ImBotClient) JoinChatroom(chatroomId string) utils.ClientErrorCode {
	return client.JoinChatroomWithOptions(chatroomId, -1, false)
}

// JoinChatroomWithPrevCount 对应 joinChatroom(String chatroomId, int prevMessageCount)。
// prevMessageCount 与 Android 一致：仅客户端入参语义，当前 ChatroomReq 不包含该字段，不会随请求下发。
func (client *ImBotClient) JoinChatroomWithPrevCount(chatroomId string, prevMessageCount int) utils.ClientErrorCode {
	return client.JoinChatroomWithOptions(chatroomId, prevMessageCount, false)
}

// JoinChatroomWithOptions 对应 joinChatroom(String chatroomId, int prevMessageCount, boolean isAutoCreate)。
func (client *ImBotClient) JoinChatroomWithOptions(chatroomId string, prevMessageCount int, isAutoCreate bool) utils.ClientErrorCode {
	_ = prevMessageCount
	data, _ := proto.Marshal(&pbobjs.ChatroomReq{
		ChatId:       chatroomId,
		IsAutoCreate: isAutoCreate,
	})
	code, _ := client.Query("c_join", chatroomId, data)
	return code
}

// QuitChatroom 对应 quitChatroom。
func (client *ImBotClient) QuitChatroom(chatroomId string) utils.ClientErrorCode {
	data, _ := proto.Marshal(&pbobjs.ChatroomReq{
		ChatId: chatroomId,
	})
	code, _ := client.Query("c_quit", chatroomId, data)
	return code
}

// SendChatroomMsg 发送聊天室消息，内部走 SendMessage（ChannelType_Chatroom，topic: c_msg）。
func (client *ImBotClient) SendChatroomMsg(chatroomId string, upMsg *pbobjs.UpMsg) (utils.ClientErrorCode, *pbobjs.PublishAckMsgBody) {
	return client.SendMessage(&models.Conversation{
		ConversationType: pbobjs.ChannelType_Chatroom,
		Conversation:     chatroomId,
	}, upMsg)
}

// SetAttributes 对应 IChatroomManager.setAttributes（topic: c_batch_add_att，与 Android PBData 一致）。
func (client *ImBotClient) SetAttributes(chatroomId string, attributes map[string]string) (utils.ClientErrorCode, *pbobjs.ChatAttBatchResp) {
	if len(attributes) == 0 {
		return utils.ClientErrorCode_Success, &pbobjs.ChatAttBatchResp{}
	}
	atts := make([]*pbobjs.ChatAttReq, 0, len(attributes))
	for k, v := range attributes {
		atts = append(atts, &pbobjs.ChatAttReq{
			Key:     k,
			Value:   v,
			IsForce: false,
		})
	}
	req := &pbobjs.ChatAttBatchReq{Atts: atts}
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("c_batch_add_att", chatroomId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.ChatAttBatchResp{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return code, nil
}

// RemoveAttributes 对应 IChatroomManager.removeAttributes（topic: c_batch_del_att）。
func (client *ImBotClient) RemoveAttributes(chatroomId string, keys []string) (utils.ClientErrorCode, *pbobjs.ChatAttBatchResp) {
	if len(keys) == 0 {
		return utils.ClientErrorCode_Success, &pbobjs.ChatAttBatchResp{}
	}
	atts := make([]*pbobjs.ChatAttReq, 0, len(keys))
	for _, k := range keys {
		atts = append(atts, &pbobjs.ChatAttReq{
			Key:     k,
			IsForce: false,
		})
	}
	req := &pbobjs.ChatAttBatchReq{Atts: atts}
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("c_batch_del_att", chatroomId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.ChatAttBatchResp{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return code, nil
}

func (client *ImBotClient) SyncChatroomMsgs(req *pbobjs.SyncChatroomReq) (utils.ClientErrorCode, *pbobjs.SyncChatroomMsgResp) {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("c_sync_msgs", req.ChatroomId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.SyncChatroomMsgResp{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return utils.ClientErrorCode_Unknown, nil
}

func (client *ImBotClient) SyncChatroomExts(req *pbobjs.SyncChatroomReq) (utils.ClientErrorCode, *pbobjs.SyncChatroomAttResp) {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("c_sync_atts", req.ChatroomId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.SyncChatroomAttResp{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return utils.ClientErrorCode_Unknown, nil
}

// AddChatAtt 单条属性写入（topic: c_add_att）；批量场景请用 SetAttributes。
func (client *ImBotClient) AddChatAtt(targetId string, att *pbobjs.ChatAttReq) (utils.ClientErrorCode, *pbobjs.ChatAttResp) {
	data, _ := proto.Marshal(att)
	code, qryAck := client.Query("c_add_att", targetId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.ChatAttResp{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return code, nil
}

// DelChatAtt 单条属性删除（topic: c_del_att）；批量删除请用 RemoveAttributes。
func (client *ImBotClient) DelChatAtt(targetId string, att *pbobjs.ChatAttReq) (utils.ClientErrorCode, *pbobjs.ChatAttResp) {
	data, _ := proto.Marshal(att)
	code, qryAck := client.Query("c_del_att", targetId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.ChatAttResp{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return code, nil
}

// JoinChatroomAutoCreate 兼容旧 API：JoinChatroom(chatroomId, isAutoCreate)。
func (client *ImBotClient) JoinChatroomAutoCreate(chatroomId string, isAutoCreate bool) utils.ClientErrorCode {
	return client.JoinChatroomWithOptions(chatroomId, -1, isAutoCreate)
}
