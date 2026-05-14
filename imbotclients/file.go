package imbotclients

import (
	"github.com/juggleim/imbot-sdk-go/imbotclients/pbdefines/pbobjs"
	"github.com/juggleim/imbot-sdk-go/utils"
	"google.golang.org/protobuf/proto"
)

func (client *ImBotClient) GetFileCred(req *pbobjs.QryFileCredReq) (utils.ClientErrorCode, *pbobjs.QryFileCredResp) {
	data, _ := proto.Marshal(req)
	code, qryAck := client.Query("file_cred", client.UserId, data)
	if code == utils.ClientErrorCode_Success && qryAck != nil && qryAck.Code == imErrorCodeSuccess {
		resp := &pbobjs.QryFileCredResp{}
		if err := proto.Unmarshal(qryAck.Data, resp); err != nil {
			return utils.ClientErrorCode_Unknown, nil
		}
		return utils.ClientErrorCode_Success, resp
	}
	return code, nil
}
