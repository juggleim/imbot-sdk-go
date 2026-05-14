package imbotclients

import (
	"github.com/juggleim/imbot-sdk-go/imbotclients/pbdefines/pbobjs"
	"github.com/juggleim/imbot-sdk-go/utils"
	"google.golang.org/protobuf/proto"
)

func (client *ImBotClient) QryHistoryMsgs(req *pbobjs.QryHisMsgsReq) (utils.ClientErrorCode, *pbobjs.DownMsgSet) {
	return client.queryDownMsgSet("qry_hismsgs", req.TargetId, req)
}

func (client *ImBotClient) QryFirstUnreadMsg(req *pbobjs.QryFirstUnreadMsgReq) (utils.ClientErrorCode, *pbobjs.DownMsg) {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("qry_first_unread_msg", req.TargetId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		msg := &pbobjs.DownMsg{}
		if err := proto.Unmarshal(qryAck.Data, msg); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, msg
	}
	return code, nil
}

func (client *ImBotClient) DelHisMsgs(req *pbobjs.DelHisMsgsReq) utils.ClientErrorCode {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("del_msg", req.TargetId, data)
	return ackCode(code, qryAck)
}

func (client *ImBotClient) RecallMsg(req *pbobjs.RecallMsgReq) (utils.ClientErrorCode, *pbobjs.QueryAckMsgBody) {
	data, _ := proto.Marshal(req)
	return client.Query("recall_msg", req.TargetId, data)
}

func (client *ImBotClient) ModifyMsg(req *pbobjs.ModifyMsgReq) (utils.ClientErrorCode, *pbobjs.QueryAckMsgBody) {
	data, _ := proto.Marshal(req)
	return client.Query("modify_msg", req.TargetId, data)
}

func (client *ImBotClient) MarkReadMsg(req *pbobjs.MarkReadReq) (utils.ClientErrorCode, *pbobjs.QueryAckMsgBody) {
	data, _ := proto.Marshal(req)
	return client.Query("mark_read", client.UserId, data)
}

func (client *ImBotClient) QryHisMsgsByIds(targetId string, req *pbobjs.QryHisMsgByIdsReq) (utils.ClientErrorCode, *pbobjs.DownMsgSet) {
	return client.queryDownMsgSet("qry_hismsg_by_ids", targetId, req)
}

func (client *ImBotClient) QryReadMsgDetail(req *pbobjs.QryReadDetailReq) (utils.ClientErrorCode, *pbobjs.QryReadDetailResp) {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("qry_read_detail", req.TargetId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.QryReadDetailResp{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return code, nil
}

func (client *ImBotClient) CleanHisMsgs(req *pbobjs.CleanHisMsgReq) utils.ClientErrorCode {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("clean_hismsg", client.UserId, data)
	return ackCode(code, qryAck)
}

func (client *ImBotClient) QryMergedMsgs(msgId string, req *pbobjs.QryMergedMsgsReq) (utils.ClientErrorCode, *pbobjs.DownMsgSet) {
	return client.queryDownMsgSet("qry_merged_msgs", msgId, req)
}

func (client *ImBotClient) BatchTranslate(req *pbobjs.TransReq) (utils.ClientErrorCode, *pbobjs.TransReq) {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("batch_trans", client.UserId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.TransReq{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return code, nil
}

func (client *ImBotClient) MsgSearch(req *pbobjs.SearchMsgsReq) (utils.ClientErrorCode, *pbobjs.SearchMsgsResp) {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("msg_search", req.TargetId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.SearchMsgsResp{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return code, nil
}

func (client *ImBotClient) MsgGlobalSearch(req *pbobjs.SearchMsgsReq) (utils.ClientErrorCode, *pbobjs.BatchSearchMsgsResp) {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("msg_global_search", client.UserId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.BatchSearchMsgsResp{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return code, nil
}

func (client *ImBotClient) SetTopMsg(req *pbobjs.TopMsgReq) utils.ClientErrorCode {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("set_top_msg", req.TargetId, data)
	return ackCode(code, qryAck)
}

func (client *ImBotClient) DelTopMsg(req *pbobjs.TopMsgReq) utils.ClientErrorCode {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("del_top_msg", req.TargetId, data)
	return ackCode(code, qryAck)
}

func (client *ImBotClient) SubStreamMsg(req *pbobjs.SubStreamMsgsReq) utils.ClientErrorCode {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("sup_stream_msg", client.UserId, data)
	return ackCode(code, qryAck)
}

func (client *ImBotClient) queryDownMsgSet(method, targetId string, req proto.Message) (utils.ClientErrorCode, *pbobjs.DownMsgSet) {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query(method, targetId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.DownMsgSet{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return code, nil
}
