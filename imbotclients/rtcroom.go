package imbotclients

import (
	"github.com/juggleim/imbot-sdk-go/imbotclients/pbdefines/pbobjs"
	"github.com/juggleim/imbot-sdk-go/utils"
	"google.golang.org/protobuf/proto"
)

// RtcJoinAlreadyInRoom 与 simulator/wsclients 中 JoinRtcRoom 的特殊成功码一致（已在房间内等场景）。
const RtcJoinAlreadyInRoom utils.ClientErrorCode = 16002

func (client *ImBotClient) CreateRtcRoom(room *pbobjs.RtcInviteReq) (utils.ClientErrorCode, *pbobjs.RtcRoom) {
	data, _ := proto.Marshal(room)
	code, qryAck := client.Query("rtc_create", client.UserId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.RtcRoom{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return code, nil
}

func (client *ImBotClient) DestroyRtcRoom(roomId string) utils.ClientErrorCode {
	code, _ := client.Query("rtc_destroy", roomId, []byte{})
	return code
}

func (client *ImBotClient) JoinRtcRoom(room *pbobjs.RtcInviteReq) (utils.ClientErrorCode, *pbobjs.RtcRoom) {
	data, _ := proto.Marshal(room)
	code, qryAck := client.Query("rtc_join", room.RoomId, data)
	if (code == utils.ClientErrorCode_Success || code == RtcJoinAlreadyInRoom) && qryAck != nil {
		resp := &pbobjs.RtcRoom{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return code, resp
	}
	return code, nil
}

func (client *ImBotClient) QuitRtcRoom(roomId string) utils.ClientErrorCode {
	code, _ := client.Query("rtc_quit", roomId, []byte{})
	return code
}

func (client *ImBotClient) QryRtcRoom(roomId string) (utils.ClientErrorCode, *pbobjs.RtcRoom) {
	code, qryAck := client.Query("rtc_qry", roomId, nil)
	if code == utils.ClientErrorCode_Success && qryAck != nil {
		resp := &pbobjs.RtcRoom{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return code, resp
	}
	return code, nil
}

func (client *ImBotClient) RtcRoomPing(roomId string) utils.ClientErrorCode {
	code, _ := client.Query("rtc_ping", roomId, nil)
	return code
}

func (client *ImBotClient) RtcInvite(req *pbobjs.RtcInviteReq) (utils.ClientErrorCode, *pbobjs.RtcAuth) {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("rtc_invite", client.UserId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.RtcAuth{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return code, nil
}

func (client *ImBotClient) QryRtcMemberRooms() (utils.ClientErrorCode, *pbobjs.RtcMemberRooms) {
	code, qryAck := client.Query("rtc_member_rooms", client.UserId, nil)
	if code == utils.ClientErrorCode_Success && qryAck != nil {
		resp := &pbobjs.RtcMemberRooms{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return code, resp
	}
	return code, nil
}

func (client *ImBotClient) RtcHangup(roomId string) utils.ClientErrorCode {
	code, qryAck := client.Query("rtc_hangup", roomId, nil)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		return utils.ClientErrorCode_Success
	}
	if code != utils.ClientErrorCode_Success {
		return code
	}
	if qryAck != nil {
		return utils.ClientErrorCode(qryAck.Code)
	}
	return utils.ClientErrorCode_Unknown
}
