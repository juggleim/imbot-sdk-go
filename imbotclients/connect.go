package imbotclients

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/juggleim/imbot-sdk-go/imbotclients/pbdefines/pbobjs"
	"github.com/juggleim/imbot-sdk-go/utils"
	"google.golang.org/protobuf/proto"
)

type IConnectionStatusChangeListener interface {
	OnStatusChange(status utils.ConnectState, code utils.ClientErrorCode)
}

func (client *ImBotClient) Connect(token string) (utils.ClientErrorCode, *pbobjs.ConnectAckMsgBody) {
	if client.state != utils.State_Disconnect {
		return utils.ClientErrorCode_ConnectExisted, nil
	}
	client.stopHeartbeat()
	client.handshakeComplete.Store(false)
	client.pendingDisconnectCode.Store(0)
	client.suppressAutoReconnect.Store(false)

	client.Token = token
	connectMsgBody := &pbobjs.ConnectMsgBody{
		ProtoId:    protoID,
		SdkVersion: "1.0.0",
		Appkey:     client.AppKey,
		Token:      client.Token,
		Platform:   client.Platform,
	}

	u, err := client.wsURL()
	if err != nil {
		fmt.Println(err)
		client.changeConnectionStatus(utils.State_Disconnect, utils.ClientErrorCode_SocketFailed)
		return utils.ClientErrorCode_SocketFailed, nil
	}
	header := http.Header{}
	header.Add("x-appkey", client.AppKey)
	header.Add("x-token", client.Token)

	c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		fmt.Println("addr:", u.String(), "err:", err)
		client.changeConnectionStatus(utils.State_Disconnect, utils.ClientErrorCode_SocketFailed)
		return utils.ClientErrorCode_SocketFailed, nil
	}
	client.lock.Lock()
	client.conn = c
	client.lock.Unlock()

	wsMsg := &pbobjs.ImWebsocketMsg{
		Version: version1,
		Cmd:     cmdConnect,
		Qos:     qosNeedAck,
		Testof: &pbobjs.ImWebsocketMsg_ConnectMsgBody{
			ConnectMsgBody: connectMsgBody,
		},
	}
	wsMsgBs, _ := proto.Marshal(wsMsg)
	err = c.WriteMessage(websocket.BinaryMessage, wsMsgBs)
	if err != nil {
		fmt.Println(err)
		client.closeConn()
		client.changeConnectionStatus(utils.State_Disconnect, utils.ClientErrorCode_ConnectTimeout)
		return utils.ClientErrorCode_ConnectTimeout, nil
	}

	client.changeConnectionStatus(utils.State_Connecting, utils.ClientErrorCode_Success)
	go client.startListener()

	connAckObj, err := client.connAckAccessor.GetWithTimeout(10 * time.Second)
	if err != nil {
		fmt.Println(err)
		client.closeConn()
		client.changeConnectionStatus(utils.State_Disconnect, utils.ClientErrorCode_ConnectTimeout)
		return utils.ClientErrorCode_ConnectTimeout, nil
	}
	connAck := connAckObj.(*pbobjs.ConnectAckMsgBody)
	clientCode := trans2ClientErrorCode(connAck.Code)
	if connAck.Code == imErrorCodeSuccess {
		client.UserId = connAck.UserId
		client.markHeartbeatRX()
		client.resetReconnectBackoff()
		client.handshakeComplete.Store(true)
		client.changeConnectionStatus(utils.State_connected, clientCode)
		client.startHeartbeat()
		return clientCode, connAck
	}
	client.closeConn()
	client.changeConnectionStatus(utils.State_Disconnect, clientCode)
	return clientCode, connAck
}

func (client *ImBotClient) Reconnect() (utils.ClientErrorCode, *pbobjs.ConnectAckMsgBody) {
	if client.state == utils.State_Connecting {
		return utils.ClientErrorCode_ConnectExisted, nil
	}
	client.suppressAutoReconnect.Store(false)
	client.closeConn()
	client.changeConnectionStatus(utils.State_Disconnect, utils.ClientErrorCode_Success)
	return client.Connect(client.Token)
}

func (client *ImBotClient) Disconnect() {
	client.suppressAutoReconnect.Store(true)
	client.stopHeartbeat()
	if client.conn != nil {
		disMsg := &pbobjs.ImWebsocketMsg{
			Version: version1,
			Cmd:     cmdDisconnect,
			Qos:     qosNoAck,
			Testof: &pbobjs.ImWebsocketMsg_DisconnectMsgBody{
				DisconnectMsgBody: &pbobjs.DisconnectMsgBody{
					Code: imErrorCodeSuccess,
				},
			},
		}
		wsMsgBs, _ := proto.Marshal(disMsg)
		_ = client.WriteMessage(wsMsgBs)
	}
	client.closeConn()
}

func (client *ImBotClient) Logout() {
	client.suppressAutoReconnect.Store(true)
	client.stopHeartbeat()
	if client.conn != nil {
		disMsg := &pbobjs.ImWebsocketMsg{
			Version: version1,
			Cmd:     cmdDisconnect,
			Qos:     qosNoAck,
			Testof: &pbobjs.ImWebsocketMsg_DisconnectMsgBody{
				DisconnectMsgBody: &pbobjs.DisconnectMsgBody{
					Code: imErrorCodeConnectLogout,
				},
			},
		}
		wsMsgBs, _ := proto.Marshal(disMsg)
		_ = client.WriteMessage(wsMsgBs)
	}
	client.closeConn()
}

func (client *ImBotClient) AddConnectionStatusChangeListener(listener IConnectionStatusChangeListener) {
	if listener == nil {
		return
	}
	client.connectChgListeners = append(client.connectChgListeners, listener)
}

func (client *ImBotClient) changeConnectionStatus(status utils.ConnectState, code utils.ClientErrorCode) {
	client.state = status
	for _, listener := range client.connectChgListeners {
		if listener != nil {
			listener.OnStatusChange(status, code)
		}
	}
}
