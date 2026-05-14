package imbotclients

import (
	"github.com/juggleim/imbot-sdk-go/imbotclients/pbdefines/pbobjs"
	"github.com/juggleim/imbot-sdk-go/utils"
	"google.golang.org/protobuf/proto"
)

// FetchUserInfo 对应 IUserInfoManager.fetchUserInfo（topic: qry_user_info；TargetId 为被查询用户，与 Android PBData 一致）。
func (client *ImBotClient) FetchUserInfo(userId string) (utils.ClientErrorCode, *pbobjs.UserInfo) {
	req := &pbobjs.UserIdReq{UserId: userId}
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("qry_user_info", userId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.UserInfo{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return code, nil
}

// FetchGroupInfo 对应 fetchGroupInfo（topic: qry_group_info）。
func (client *ImBotClient) FetchGroupInfo(groupId string) (utils.ClientErrorCode, *pbobjs.GroupInfo) {
	req := &pbobjs.GroupInfoReq{GroupId: groupId}
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("qry_group_info", groupId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.GroupInfo{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return code, nil
}

// FetchFriendInfo 对应 fetchFriendInfo（topic: qry_friend_infos；TargetId 为当前用户）。
func (client *ImBotClient) FetchFriendInfo(friendUserId string) (utils.ClientErrorCode, *pbobjs.FriendInfo) {
	req := &pbobjs.FriendIdsReq{FriendIds: []string{friendUserId}}
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("qry_friend_infos", client.UserId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.FriendInfos{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		if len(resp.Items) > 0 {
			return utils.ClientErrorCode_Success, resp.Items[0]
		}
		return utils.ClientErrorCode_Success, nil
	}
	return code, nil
}

// GetUserStatus 对应 IUserInfoManager.getUserStatus（topic: qry_user_status）。
func (client *ImBotClient) GetUserStatus(req *pbobjs.UserIdsReq) (utils.ClientErrorCode, *pbobjs.UserStatusList) {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("qry_user_status", client.UserId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.UserStatusList{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return code, nil
}

// SubscribeUserStatus 订阅用户在线状态（topic: sub_user_status）。
func (client *ImBotClient) SubscribeUserStatus(req *pbobjs.SubUsersReq) (utils.ClientErrorCode, *pbobjs.UserStatusList) {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("sub_user_status", client.UserId, data)
	if code != utils.ClientErrorCode_Success || qryAck == nil {
		return code, nil
	}
	if qryAck.Code != imErrorCodeSuccess {
		return utils.ClientErrorCode(qryAck.Code), nil
	}
	resp := &pbobjs.UserStatusList{}
	if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
		return utils.ClientErrorCode_Unknown, nil
	}
	return utils.ClientErrorCode_Success, resp
}

// UnsubscribeUserStatus 取消订阅（topic: unsub_user_status，与 statussubscriptions starter 一致）。
func (client *ImBotClient) UnsubscribeUserStatus(req *pbobjs.SubUsersReq) utils.ClientErrorCode {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("unsub_user_status", client.UserId, data)
	return ackCode(code, qryAck)
}

// PubUserStatus 对应 pub_user_status 发布。
func (client *ImBotClient) PubUserStatus(upMsg *pbobjs.UpMsg) (utils.ClientErrorCode, *pbobjs.PublishAckMsgBody) {
	data, _ := proto.Marshal(upMsg)
	return client.Publish("pub_user_status", client.UserId, data)
}

// SetUserUndisturb 对应 set_user_undisturb。
func (client *ImBotClient) SetUserUndisturb(req *pbobjs.UserUndisturb) utils.ClientErrorCode {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("set_user_undisturb", client.UserId, data)
	return ackCode(code, qryAck)
}

// GetUserUndisturb 对应 get_user_undisturb。
func (client *ImBotClient) GetUserUndisturb() (utils.ClientErrorCode, *pbobjs.UserUndisturb) {
	data, _ := proto.Marshal(&pbobjs.Nil{})
	code, qryAck := client.Query("get_user_undisturb", client.UserId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.UserUndisturb{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return code, nil
}

// --- 兼容旧方法名 ---

func (client *ImBotClient) QryUserInfo(req *pbobjs.UserIdReq) (utils.ClientErrorCode, *pbobjs.UserInfo) {
	if req == nil {
		return utils.ClientErrorCode_Unknown, nil
	}
	return client.FetchUserInfo(req.GetUserId())
}

func (client *ImBotClient) SubUsers(req *pbobjs.SubUsersReq) (utils.ClientErrorCode, *pbobjs.UserStatusList) {
	return client.SubscribeUserStatus(req)
}

func (client *ImBotClient) UnSubUsers(req *pbobjs.SubUsersReq) utils.ClientErrorCode {
	return client.UnsubscribeUserStatus(req)
}
