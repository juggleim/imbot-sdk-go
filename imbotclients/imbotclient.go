package imbotclients

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/juggleim/imbot-sdk-go/imbotclients/pbdefines/pbobjs"
	"github.com/juggleim/imbot-sdk-go/utils"
	"google.golang.org/protobuf/proto"
)

const (
	protoID  = "jug9le1m"
	version1 = 1

	cmdConnect      = 0
	cmdConnectAck   = 1
	cmdDisconnect   = 2
	cmdPublish      = 3
	cmdPublishAck   = 4
	cmdQuery        = 5
	cmdQueryAck     = 6
	cmdQueryConfirm = 7
	cmdPing         = 8
	cmdPong         = 9

	qosNoAck   = 0
	qosNeedAck = 1

	imErrorCodeSuccess       = 0
	imErrorCodeConnectLogout = 10004
)

type ImBotClient struct {
	Address  string
	AppKey   string
	Token    string
	Platform string

	// AutoReconnect 为 true（默认）时，在非主动断开场景下按退避策略自动重连，行为对齐 Android ConnectionManager + IntervalGenerator。
	AutoReconnect bool

	DisconnectCallback  func(code utils.ClientErrorCode, disMsg *pbobjs.DisconnectMsgBody)
	OnMessageCallBack   func(msg *pbobjs.DownMsg)
	OnStreamMsgCallBack func(msg *pbobjs.StreamDownMsg)

	connectChgListeners []IConnectionStatusChangeListener
	converChgListeners  []IConversationChangeListener
	messageListeners    []IMessageListener

	UserId string

	conn            *websocket.Conn
	state           utils.ConnectState
	accssorCache    sync.Map
	myIndex         uint16
	connAckAccessor *utils.DataAccessor
	manualPong      chan struct{}
	lock            *sync.RWMutex
	converCache     sync.Map

	lastRxUnixMilli       atomic.Int64
	pendingDisconnectCode atomic.Int32
	handshakeComplete     atomic.Bool
	suppressAutoReconnect atomic.Bool
	reconnectBackoff      reconnectBackoff
	reconnectBusy         int32

	heartbeatCancelMu sync.Mutex
	heartbeatCancel   context.CancelFunc

	inboxTime        int64
	sendboxTime      int64
	totalUnreadCount int
}

func NewImBotClient(address, appkey string) *ImBotClient {
	return &ImBotClient{
		AppKey:          appkey,
		Address:         address,
		Platform:        "Bot",
		AutoReconnect:   true,
		accssorCache:    sync.Map{},
		connAckAccessor: utils.NewDataAccessor(),
		manualPong:      make(chan struct{}, 1),
		lock:            &sync.RWMutex{},
	}
}

func (client *ImBotClient) GetState() utils.ConnectState {
	return client.state
}

func (client *ImBotClient) WriteMessage(data []byte) error {
	client.lock.Lock()
	defer client.lock.Unlock()
	if client.state != utils.State_connected || client.conn == nil {
		return fmt.Errorf("not connected")
	}
	return client.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (client *ImBotClient) startListener() {
	for {
		client.lock.RLock()
		c := client.conn
		client.lock.RUnlock()
		if c == nil {
			break
		}
		_, msgBs, err := c.ReadMessage()
		if err != nil {
			fmt.Println(err)
			client.handleReadLoopEnded(err)
			break
		}
		client.markHeartbeatRX()
		wsImMsg := &pbobjs.ImWebsocketMsg{}
		err = proto.Unmarshal(msgBs, wsImMsg)
		if err != nil {
			fmt.Println("failed to decode pb data:", err)
			continue
		}
		go client.handleMsg(wsImMsg)
	}
	fmt.Println("Stop bot client listener.")
}

func (client *ImBotClient) handleMsg(wsImMsg *pbobjs.ImWebsocketMsg) {
	switch wsImMsg.Cmd {
	case cmdConnectAck:
		client.OnConnectAck(wsImMsg.GetConnectAckMsgBody())
	case cmdDisconnect:
		client.OnDisconnect(wsImMsg.GetDisconnectMsgBody())
	case cmdPublish:
		client.OnPublish(wsImMsg.GetPublishMsgBody(), int(wsImMsg.Qos))
	case cmdPublishAck:
		client.OnPublishAck(wsImMsg.GetPubAckMsgBody())
	case cmdQueryAck:
		client.OnQueryAck(wsImMsg.GetQryAckMsgBody())
	case cmdPong:
		client.OnPong(wsImMsg)
	}
}

func (client *ImBotClient) OnConnectAck(msg *pbobjs.ConnectAckMsgBody) {
	client.connAckAccessor.Put(msg)
}

func (client *ImBotClient) OnPublishAck(msg *pbobjs.PublishAckMsgBody) {
	if msg == nil {
		return
	}
	dataAccessor, ok := client.accssorCache.LoadAndDelete(msg.Index)
	if ok {
		dataAccessor.(*utils.DataAccessor).Put(msg)
	}
}

func (client *ImBotClient) OnQueryAck(msg *pbobjs.QueryAckMsgBody) {
	if msg == nil {
		return
	}
	dataAccessor, ok := client.accssorCache.LoadAndDelete(msg.Index)
	if ok {
		dataAccessor.(*utils.DataAccessor).Put(msg)
	}
}

func (client *ImBotClient) OnDisconnect(msg *pbobjs.DisconnectMsgBody) {
	client.suppressAutoReconnect.Store(true)
	client.stopHeartbeat()
	if msg == nil {
		msg = &pbobjs.DisconnectMsgBody{}
	}
	if client.DisconnectCallback != nil {
		client.DisconnectCallback(trans2ClientErrorCode(msg.Code), msg)
	}
	client.pendingDisconnectCode.Store(int32(trans2ClientErrorCode(msg.Code)))
	client.closeConn()
	// 状态变更由读循环 handleReadLoopEnded 统一触发，避免与断线读错误重复回调
}

func (client *ImBotClient) OnPong(_ *pbobjs.ImWebsocketMsg) {
	select {
	case client.manualPong <- struct{}{}:
	default:
	}
}

func (client *ImBotClient) OnPublish(msg *pbobjs.PublishMsgBody, needAck int) {
	if msg == nil {
		return
	}
	switch msg.Topic {
	case "msg":
		downMsg := pbobjs.DownMsg{}
		err := proto.Unmarshal(msg.Data, &downMsg)
		if err == nil {
			client.handleDownMsg(&downMsg)
		} else {
			fmt.Println("msg unmarshal error:", err)
		}
	case "ntf":
		ntf := pbobjs.Notify{}
		err := proto.Unmarshal(msg.Data, &ntf)
		if err != nil {
			fmt.Println("ntf unmarshal error:", err)
			break
		}
		switch ntf.Type {
		case pbobjs.NotifyType_Msg:
			isContinue := true
			for isContinue {
				code, downSet := client.SyncMsgs(&pbobjs.SyncMsgReq{
					SyncTime:        client.inboxTime,
					SendBoxSyncTime: client.sendboxTime,
				})
				if code == utils.ClientErrorCode_Success && downSet != nil {
					for _, downMsg := range downSet.Msgs {
						client.handleDownMsg(downMsg)
					}
					isContinue = !downSet.IsFinished
				} else {
					fmt.Println("ntf pull msg error, code:", code)
					isContinue = false
				}
			}
		case pbobjs.NotifyType_ChatroomMsg:
			fmt.Println("ntf chatroom msg ignored: SyncChatroomReq is not included in current pbdefines")
		}
	case "stream_msg":
		streamMsg := pbobjs.StreamDownMsg{}
		err := proto.Unmarshal(msg.Data, &streamMsg)
		if err == nil && client.OnStreamMsgCallBack != nil {
			client.OnStreamMsgCallBack(&streamMsg)
		}
	default:
		fmt.Println(msg.Topic, msg.Data)
	}
	if needAck > 0 {
		ackMsg := &pbobjs.ImWebsocketMsg{
			Version: version1,
			Cmd:     cmdPublishAck,
			Qos:     qosNoAck,
			Testof: &pbobjs.ImWebsocketMsg_PubAckMsgBody{
				PubAckMsgBody: &pbobjs.PublishAckMsgBody{
					Index: msg.Index,
				},
			},
		}
		wsMsgBs, _ := proto.Marshal(ackMsg)
		_ = client.WriteMessage(wsMsgBs)
	}
}

func (client *ImBotClient) handleDownMsg(downMsg *pbobjs.DownMsg) {
	if downMsg == nil {
		return
	}
	if client.OnMessageCallBack != nil {
		client.OnMessageCallBack(downMsg)
	}
	msg := client.notifyMessageReceive(downMsg)
	client.notifyConversationForMessage(msg, downMsg)
	if downMsg.IsSend {
		client.sendboxTime = downMsg.MsgTime
	} else {
		client.inboxTime = downMsg.MsgTime
	}
}

func (client *ImBotClient) Publish(method, targetId string, data []byte) (utils.ClientErrorCode, *pbobjs.PublishAckMsgBody) {
	if client.state != utils.State_connected {
		return utils.ClientErrorCode_ConnectClosed, nil
	}
	index := int32(client.getMyIndex())
	protoMsg := &pbobjs.ImWebsocketMsg{
		Version: version1,
		Cmd:     cmdPublish,
		Qos:     qosNeedAck,
		Testof: &pbobjs.ImWebsocketMsg_PublishMsgBody{
			PublishMsgBody: &pbobjs.PublishMsgBody{
				Index:    index,
				Topic:    method,
				TargetId: targetId,
				Data:     data,
			},
		},
	}
	dataAccessor := utils.NewDataAccessor()
	client.accssorCache.Store(index, dataAccessor)

	wsMsgBs, _ := proto.Marshal(protoMsg)
	_ = client.WriteMessage(wsMsgBs)
	obj, err := dataAccessor.GetWithTimeout(10 * time.Second)
	if err != nil {
		return utils.ClientErrorCode_SendTimeout, nil
	}
	pubAck := obj.(*pbobjs.PublishAckMsgBody)
	return trans2ClientErrorCode(pubAck.Code), pubAck
}

func (client *ImBotClient) Ping() utils.ClientErrorCode {
	if client.state != utils.State_connected {
		return utils.ClientErrorCode_ConnectClosed
	}
	select {
	case <-client.manualPong:
	default:
	}
	pingMsg := &pbobjs.ImWebsocketMsg{
		Version: version1,
		Cmd:     cmdPing,
		Qos:     qosNeedAck,
	}
	wsMsgBs, _ := proto.Marshal(pingMsg)
	_ = client.WriteMessage(wsMsgBs)
	select {
	case <-client.manualPong:
		return utils.ClientErrorCode_Success
	case <-time.After(15 * time.Second):
		return utils.ClientErrorCode_PingTimeout
	}
}

func (client *ImBotClient) Query(method, targetId string, data []byte) (utils.ClientErrorCode, *pbobjs.QueryAckMsgBody) {
	if client.state != utils.State_connected {
		return utils.ClientErrorCode_ConnectClosed, nil
	}
	index := int32(client.getMyIndex())
	protoMsg := &pbobjs.ImWebsocketMsg{
		Version: version1,
		Cmd:     cmdQuery,
		Qos:     qosNeedAck,
		Testof: &pbobjs.ImWebsocketMsg_QryMsgBody{
			QryMsgBody: &pbobjs.QueryMsgBody{
				Index:    index,
				Topic:    method,
				TargetId: targetId,
				Data:     data,
			},
		},
	}
	dataAccessor := utils.NewDataAccessor()
	client.accssorCache.Store(index, dataAccessor)

	wsMsgBs, _ := proto.Marshal(protoMsg)
	_ = client.WriteMessage(wsMsgBs)
	obj, err := dataAccessor.GetWithTimeout(10 * time.Second)
	if err != nil {
		return utils.ClientErrorCode_QueryTimeout, nil
	}
	queryAck := obj.(*pbobjs.QueryAckMsgBody)
	return trans2ClientErrorCode(queryAck.Code), queryAck
}

func (client *ImBotClient) QueryConfirm(index int32) utils.ClientErrorCode {
	if client.state != utils.State_connected {
		return utils.ClientErrorCode_ConnectClosed
	}
	confirmMsg := &pbobjs.ImWebsocketMsg{
		Version: version1,
		Cmd:     cmdQueryConfirm,
		Qos:     qosNoAck,
		Testof: &pbobjs.ImWebsocketMsg_QryConfirmMsgBody{
			QryConfirmMsgBody: &pbobjs.QueryConfirmMsgBody{
				Index: index,
			},
		},
	}
	wsMsgBs, _ := proto.Marshal(confirmMsg)
	_ = client.WriteMessage(wsMsgBs)
	return utils.ClientErrorCode_Success
}

func (client *ImBotClient) getMyIndex() uint16 {
	client.myIndex = client.myIndex + 1
	return client.myIndex
}

func (client *ImBotClient) wsURL() (*url.URL, error) {
	if strings.HasPrefix(client.Address, "wss://") {
		return &url.URL{Scheme: "wss", Host: client.Address[6:], Path: "/imbot"}, nil
	}
	if strings.HasPrefix(client.Address, "ws://") {
		return &url.URL{Scheme: "ws", Host: client.Address[5:], Path: "/imbot"}, nil
	}
	return nil, fmt.Errorf("unsupported websocket address: %s", client.Address)
}

func (client *ImBotClient) closeConn() {
	client.lock.Lock()
	defer client.lock.Unlock()
	if client.conn != nil {
		_ = client.conn.Close()
		client.conn = nil
	}
}

func trans2ClientErrorCode(code int32) utils.ClientErrorCode {
	if code == imErrorCodeSuccess {
		return utils.ClientErrorCode_Success
	}
	return utils.ClientErrorCode(code)
}

func ackCode(code utils.ClientErrorCode, ack *pbobjs.QueryAckMsgBody) utils.ClientErrorCode {
	if code != utils.ClientErrorCode_Success {
		return code
	}
	if ack != nil && ack.Code != imErrorCodeSuccess {
		return utils.ClientErrorCode(ack.Code)
	}
	return utils.ClientErrorCode_Success
}
