// Webhook 模式示例：初始化 Bot 客户端，监听回调消息，收到后回复 "ok"。
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/juggleim/imbot-sdk-go/models"
	"github.com/juggleim/imbot-sdk-go/models/messages"
	"github.com/juggleim/imbot-sdk-go/webhookclients"
)

const (
	baseURL = "http://127.0.0.1:9003"
	appKey  = "appkey"
	token   = "CgZhcHBrZXkaIM-hDRn0gGcnc7X5Mp3SbsVHnJb53v4actvR-j9d0V4O"

	// apiKey 用于校验 IM 服务回调请求的 Authorization: Bearer <apiKey>。
	apiKey = "webhook-secret"

	// listenAddr / callbackPath 为本地接收回调的地址与路径；
	// callbackURL 需为 IM 服务能访问到的公网/内网可达地址。
	listenAddr   = ":9000"
	callbackPath = "/callback"
	callbackURL  = "http://127.0.0.1:9000/callback"
)

type replyListener struct {
	client *webhookclients.Client
}

// OnMessageReceive 收到任意消息后回复一条文本 "ok"（单聊回复，会话取发送方）。
func (l replyListener) OnMessageReceive(msg *models.Message) {
	log.Printf("received: from=%s type=%s", msg.SenderId, msg.MsgType)

	reply := messages.NewTextMessage("ok")
	payload, err := reply.Encode()
	if err != nil {
		log.Printf("encode reply failed: %v", err)
		return
	}
	req := &webhookclients.SendMessageReq{
		Conversation: &webhookclients.Conversation{
			ConversationType: int(msg.Conversation.ConversationType),
			ConversationId:   msg.Conversation.ConversationId,
		},
		MsgType:    reply.GetContentType(),
		MsgContent: string(payload),
	}
	code, ack := l.client.SendMessage(req)
	if code != 0 {
		log.Printf("reply failed: code=%d", code)
		return
	}
	log.Printf("replied ok, msgId=%s", ack.MsgId)
}

func main() {
	c := webhookclients.NewImBotWebhookClient(baseURL, appKey, token)
	c.ApiKey = apiKey

	code, self := c.QryUserInfo("")
	if code != 0 {
		log.Fatalf("qry self failed: code=%d", code)
	}
	log.Printf("bot userId=%s nickname=%s", self.UserId, self.Nickname)

	// 向服务端注册回调地址，IM 服务会把消息 POST 到这里。
	if code := c.SetWebhook(callbackURL, apiKey, false); code != 0 {
		log.Fatalf("set webhook failed: code=%d", code)
	}

	c.AddMessageListener(replyListener{client: c})

	go func() {
		if err := c.StartReceiver(listenAddr, callbackPath); err != nil {
			log.Fatalf("receiver exited: %v", err)
		}
	}()

	fmt.Printf("webhook bot started, listening on %s%s, press Ctrl+C to exit\n", listenAddr, callbackPath)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
