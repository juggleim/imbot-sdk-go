package webhookclients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/juggleim/imbot-sdk-go/imbotclients/pbdefines/pbobjs"
	"github.com/juggleim/imbot-sdk-go/models"
	"github.com/juggleim/imbot-sdk-go/models/messages"
)

// Handler 返回处理 webhook 回调的 http.Handler，便于挂载到调用方已有的 mux/框架上。
// 服务端 botmsg 以 POST + JSON(InboundMessage) 推送消息，并带请求头
// appkey 与 Authorization: Bearer <ApiKey>。
func (c *Client) Handler() http.Handler {
	return http.HandlerFunc(c.serveCallback)
}

// StartReceiver 在 addr 上启动一个独立 HTTP server 接收回调，path 为空时默认 "/"。
// 该调用会阻塞直到 server 退出（ListenAndServe 返回），通常放到单独 goroutine 中运行。
func (c *Client) StartReceiver(addr, path string) error {
	if path == "" {
		path = "/"
	}
	mux := http.NewServeMux()
	mux.Handle(path, c.Handler())
	srv := &http.Server{Addr: addr, Handler: mux}

	c.serverLock.Lock()
	c.server = srv
	c.serverLock.Unlock()

	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

// StopReceiver 优雅关闭由 StartReceiver 启动的 HTTP server。
func (c *Client) StopReceiver(ctx context.Context) error {
	c.serverLock.Lock()
	srv := c.server
	c.server = nil
	c.serverLock.Unlock()
	if srv == nil {
		return nil
	}
	return srv.Shutdown(ctx)
}

func (c *Client) serveCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !c.verifyAuth(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body failed", http.StatusBadRequest)
		return
	}
	inbound := &InboundMessage{}
	if err := json.Unmarshal(body, inbound); err != nil {
		http.Error(w, "illegal body", http.StatusBadRequest)
		return
	}

	c.dispatchInbound(inbound)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"code":0,"msg":"success"}`))
}

// verifyAuth 校验回调请求头中的 Authorization: Bearer <ApiKey>。
// 当 c.ApiKey 为空时不做校验（视为未启用鉴权）。
func (c *Client) verifyAuth(r *http.Request) bool {
	if c.ApiKey == "" {
		return true
	}
	auth := r.Header.Get("Authorization")
	expected := fmt.Sprintf("Bearer %s", c.ApiKey)
	return auth == expected
}

func (c *Client) dispatchInbound(inbound *InboundMessage) {
	c.listenerLock.RLock()
	rawCb := c.onInboundRaw
	listeners := make([]IInboundMessageListener, len(c.listeners))
	copy(listeners, c.listeners)
	c.listenerLock.RUnlock()

	if rawCb != nil {
		rawCb(inbound)
	}
	if len(listeners) == 0 {
		return
	}
	msg := inboundToMessage(inbound)
	for _, listener := range listeners {
		if listener != nil {
			listener.OnMessageReceive(msg)
		}
	}
}

// inboundToMessage 把回调请求体转换为与 WebSocket 端一致的 *models.Message。
// 注意：回调载荷不包含群会话 id（receiver 为 Bot 自身），因此 Conversation 取 Sender，
// 适用于单聊回复场景；群聊回复请改用 SendMessage 显式指定群会话。
func inboundToMessage(in *InboundMessage) *models.Message {
	if in == nil {
		return nil
	}
	return &models.Message{
		Conversation: &models.Conversation{
			ConversationType: pbobjs.ChannelType(in.ConverType),
			ConversationId:   in.Sender,
		},
		MsgId:       in.MsgId,
		MsgTime:     in.MsgTime,
		SenderId:    in.Sender,
		MsgType:     in.MsgType,
		MsgContent:  messages.DecodeContent(in.MsgType, []byte(in.MsgContent)),
		MentionInfo: inboundMentionToModel(in.MentionInfo),
	}
}

func inboundMentionToModel(info *MentionInfo) *models.MessageMentionInfo {
	if info == nil {
		return nil
	}
	var mentionType pbobjs.MentionType
	switch strings.ToLower(info.MentionType) {
	case "mention_all":
		mentionType = pbobjs.MentionType_All
	case "mention_someone":
		mentionType = pbobjs.MentionType_Someone
	case "mention_all_someone":
		mentionType = pbobjs.MentionType_AllAndSomeone
	default:
		mentionType = pbobjs.MentionType_MentionDefault
	}
	return &models.MessageMentionInfo{
		MentionType:   mentionType,
		TargetUserIds: info.TargetUserIds,
	}
}
