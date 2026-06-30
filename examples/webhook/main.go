// Webhook 模式示例：通过 HTTP 调用 botapigateway 发送消息，
// 并启动本地 HTTP server 接收 IM 服务推送的回调消息。
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

type inboundListener struct {
	client *webhookclients.Client
}

// OnMessageReceive 收到回调消息后，若是文本则原样回一条。
func (l inboundListener) OnMessageReceive(msg *models.Message) {
	text, ok := msg.MsgContent.(*messages.TextMessage)
	if !ok {
		log.Printf("received non-text message: from=%s type=%s", msg.SenderId, msg.MsgType)
		return
	}
	log.Printf("received text: from=%s content=%s", msg.SenderId, text.Content)

	reply := messages.NewTextMessage("echo: " + text.Content)
	payload, err := reply.Encode()
	if err != nil {
		log.Printf("encode reply failed: %v", err)
		return
	}
	// 单聊回复：会话取消息发送方。
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
	log.Printf("replied msgId=%s", ack.MsgId)
}

func main() {
	base := "http://127.0.0.1:8080" // botapigateway 基础地址
	appKey := "your-appkey"
	token := "your-token"

	c := webhookclients.NewImBotWebhookClient(base, appKey, token)
	c.ApiKey = "your-webhook-apikey"

	code, self := c.QryUserInfo("")
	if code != 0 {
		log.Fatalf("qry self failed: code=%d", code)
	}
	log.Printf("bot userId=%s nickname=%s", self.UserId, self.Nickname)

	// 注册对外回调地址（IM 服务会把消息 POST 到这里）。
	if code := c.SetWebhook("http://your-bot-host:9000/callback", c.ApiKey, false); code != 0 {
		log.Fatalf("set webhook failed: code=%d", code)
	}

	c.AddMessageListener(inboundListener{client: c})

	go func() {
		if err := c.StartReceiver(":9000", "/callback"); err != nil {
			log.Fatalf("receiver exited: %v", err)
		}
	}()

	fmt.Println("webhook bot started, press Ctrl+C to exit")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
