package webhookclients

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/juggleim/imbot-sdk-go/imbotclients/pbdefines/pbobjs"
	"github.com/juggleim/imbot-sdk-go/models"
	"github.com/juggleim/imbot-sdk-go/models/messages"
	"github.com/juggleim/imbot-sdk-go/utils"
)

// fakeGateway 模拟 botapigateway 的若干路由。
func fakeGateway(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/botgateway/bots/info", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("appkey") != "ak" || r.Header.Get("token") != "tk" {
			writeJSON(w, map[string]any{"code": 20004, "msg": "auth fail"})
			return
		}
		writeJSON(w, map[string]any{"code": 0, "msg": "success", "data": map[string]any{
			"user_id": "bot1", "nickname": "Bot One",
		}})
	})
	mux.HandleFunc("/botgateway/messages/send", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		req := &SendMessageReq{}
		if err := json.Unmarshal(body, req); err != nil {
			t.Errorf("bad send body: %v", err)
		}
		if req.Conversation == nil || req.Conversation.ConversationId != "u2" {
			t.Errorf("unexpected conversation: %+v", req.Conversation)
		}
		if req.IsStorage == nil || !*req.IsStorage || req.IsCount == nil || !*req.IsCount {
			t.Errorf("flags not split correctly: %+v", req)
		}
		writeJSON(w, map[string]any{"code": 0, "msg": "success", "data": map[string]any{"msg_id": "m100"}})
	})
	return httptest.NewServer(mux)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	bs, _ := json.Marshal(v)
	_, _ = w.Write(bs)
}

func TestQryUserInfo(t *testing.T) {
	srv := fakeGateway(t)
	defer srv.Close()
	c := NewImBotWebhookClient(srv.URL, "ak", "tk")

	code, info := c.QryUserInfo("")
	if code != utils.ClientErrorCode_Success {
		t.Fatalf("code=%d", code)
	}
	if info.UserId != "bot1" || c.UserId != "bot1" {
		t.Fatalf("unexpected user info: %+v", info)
	}

	bad := NewImBotWebhookClient(srv.URL, "ak", "wrong")
	if code, _ := bad.QryUserInfo(""); code == utils.ClientErrorCode_Success {
		t.Fatalf("expected auth failure")
	}
}

func TestSendMessage(t *testing.T) {
	srv := fakeGateway(t)
	defer srv.Close()
	c := NewImBotWebhookClient(srv.URL, "ak", "tk")

	text := messages.NewTextMessage("hi")
	payload, _ := text.Encode()
	isTrue := true
	req := &SendMessageReq{
		Conversation: &Conversation{ConversationType: int(pbobjs.ChannelType_Private), ConversationId: "u2"},
		MsgType:      text.GetContentType(),
		MsgContent:   string(payload),
		IsStorage:    &isTrue,
		IsCount:      &isTrue,
	}

	code, ack := c.SendMessage(req)
	if code != utils.ClientErrorCode_Success {
		t.Fatalf("code=%d", code)
	}
	if ack.MsgId != "m100" {
		t.Fatalf("unexpected msgId: %s", ack.MsgId)
	}
}

func TestReceiverDispatchAndAuth(t *testing.T) {
	c := NewImBotWebhookClient("http://unused", "ak", "tk")
	c.ApiKey = "secret"

	var got *models.Message
	c.AddMessageListener(listenerFunc(func(m *models.Message) { got = m }))

	inbound := &InboundMessage{
		Sender: "u2", Receiver: "bot1", ConverType: int(pbobjs.ChannelType_Private),
		MsgType: messages.MessageContentTypeText, MsgContent: `{"content":"hello"}`,
		MsgId: "m1", MsgTime: 123,
	}
	body, _ := json.Marshal(inbound)

	// 缺少 Authorization 应被拒绝。
	rec := httptest.NewRecorder()
	c.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/cb", strings.NewReader(string(body))))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if got != nil {
		t.Fatalf("listener should not fire on auth failure")
	}

	// 带正确 Authorization 应解码并分发。
	req := httptest.NewRequest(http.MethodPost, "/cb", strings.NewReader(string(body)))
	req.Header.Set("Authorization", "Bearer secret")
	rec = httptest.NewRecorder()
	c.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got == nil {
		t.Fatal("listener did not fire")
	}
	if got.SenderId != "u2" || got.Conversation.ConversationId != "u2" {
		t.Fatalf("unexpected message: %+v", got)
	}
	text, ok := got.MsgContent.(*messages.TextMessage)
	if !ok || text.Content != "hello" {
		t.Fatalf("content not decoded: %+v", got.MsgContent)
	}
}

type listenerFunc func(*models.Message)

func (f listenerFunc) OnMessageReceive(msg *models.Message) { f(msg) }
