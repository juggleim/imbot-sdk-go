package imbotclients

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/juggleim/imbot-sdk-go/imbotclients/pbdefines/pbobjs"
	"github.com/juggleim/imbot-sdk-go/utils"
	"google.golang.org/protobuf/proto"
)

// 与 Android HeartbeatManager 一致：10s 发一次 ping，10s 检测一次；超过 2 个心跳周期未收到任何下行则判定超时。
const (
	heartbeatInterval       = 10 * time.Second
	heartbeatCheckInterval  = 10 * time.Second
	heartbeatReceiveTimeout = 2 * heartbeatInterval
)

// 与 Android IntervalGenerator 一致：首次 300ms，返回后下限提到 1000ms，之后每次翻倍直到 32s 封顶。
type reconnectBackoff struct {
	mu   sync.Mutex
	next int // 毫秒，下一次 sleep 使用的值（Java 先返回当前再推进）
}

func (b *reconnectBackoff) reset() {
	b.mu.Lock()
	b.next = 300
	b.mu.Unlock()
}

func (b *reconnectBackoff) nextDelay() time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := b.next
	if b.next < 1000 {
		b.next = 1000
	} else if b.next < 32000 {
		b.next *= 2
	}
	return time.Duration(out) * time.Millisecond
}

func (client *ImBotClient) resetReconnectBackoff() {
	client.reconnectBackoff.reset()
}

func (client *ImBotClient) markHeartbeatRX() {
	client.lastRxUnixMilli.Store(time.Now().UnixMilli())
}

func (client *ImBotClient) sendPingFireAndForget() {
	if client.state != utils.State_connected {
		return
	}
	pingMsg := &pbobjs.ImWebsocketMsg{
		Version: version1,
		Cmd:     cmdPing,
		Qos:     qosNeedAck,
	}
	wsMsgBs, _ := proto.Marshal(pingMsg)
	_ = client.WriteMessage(wsMsgBs)
}

func (client *ImBotClient) stopHeartbeat() {
	client.heartbeatCancelMu.Lock()
	if client.heartbeatCancel != nil {
		client.heartbeatCancel()
		client.heartbeatCancel = nil
	}
	client.heartbeatCancelMu.Unlock()
}

func (client *ImBotClient) startHeartbeat() {
	client.stopHeartbeat()
	ctx, cancel := context.WithCancel(context.Background())
	client.heartbeatCancelMu.Lock()
	client.heartbeatCancel = cancel
	client.heartbeatCancelMu.Unlock()

	go func() {
		pingTicker := time.NewTicker(heartbeatInterval)
		checkTicker := time.NewTicker(heartbeatCheckInterval)
		defer pingTicker.Stop()
		defer checkTicker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-pingTicker.C:
				if client.state != utils.State_connected {
					return
				}
				client.sendPingFireAndForget()
			case <-checkTicker.C:
				if client.state != utils.State_connected {
					return
				}
				last := client.lastRxUnixMilli.Load()
				if last == 0 {
					continue
				}
				if time.Since(time.UnixMilli(last)) >= heartbeatReceiveTimeout {
					client.handleHeartbeatTimeout()
					return
				}
			}
		}
	}()
}

func (client *ImBotClient) handleHeartbeatTimeout() {
	client.stopHeartbeat()
	client.pendingDisconnectCode.Store(int32(utils.ClientErrorCode_PingTimeout))
	client.closeConn()
}

func (client *ImBotClient) beginReconnectBackoff() {
	if !client.AutoReconnect || client.suppressAutoReconnect.Load() {
		return
	}
	if client.Token == "" {
		return
	}
	if !atomic.CompareAndSwapInt32(&client.reconnectBusy, 0, 1) {
		return
	}
	go client.runReconnectBackoff()
}

func (client *ImBotClient) runReconnectBackoff() {
	defer atomic.StoreInt32(&client.reconnectBusy, 0)

	for client.AutoReconnect && !client.suppressAutoReconnect.Load() && client.Token != "" {
		if client.GetState() != utils.State_Disconnect {
			return
		}
		d := client.reconnectBackoff.nextDelay()
		time.Sleep(d)
		if client.suppressAutoReconnect.Load() {
			return
		}
		if client.GetState() != utils.State_Disconnect {
			return
		}
		code, ack := client.Connect(client.Token)
		if code == utils.ClientErrorCode_Success {
			return
		}
		if terminalConnectFailure(code, ack) {
			return
		}
	}
}

func terminalConnectFailure(code utils.ClientErrorCode, ack *pbobjs.ConnectAckMsgBody) bool {
	if code == utils.ClientErrorCode_ConnectExisted {
		return true
	}
	if ack != nil && ack.Code != imErrorCodeSuccess {
		return true
	}
	return false
}

func (client *ImBotClient) handleReadLoopEnded(_ error) {
	client.stopHeartbeat()
	hadSession := client.handshakeComplete.Swap(false)
	alreadyDisconnected := client.state == utils.State_Disconnect
	client.closeConn()
	code := utils.ClientErrorCode_ConnectClosed
	if v := client.pendingDisconnectCode.Swap(0); v != 0 {
		code = utils.ClientErrorCode(v)
	}
	if !alreadyDisconnected {
		client.changeConnectionStatus(utils.State_Disconnect, code)
	}
	if !alreadyDisconnected && hadSession && client.AutoReconnect && !client.suppressAutoReconnect.Load() && client.Token != "" {
		client.beginReconnectBackoff()
	}
}
