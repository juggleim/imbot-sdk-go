package webhookclients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/juggleim/imbot-sdk-go/models"
	"github.com/juggleim/imbot-sdk-go/utils"
)

const (
	// defaultPrefix 与服务端 launcher 中 botApiRouters.Route(engine, "botgateway") 一致。
	defaultPrefix = "botgateway"

	headerAppKey    = "appkey"
	headerToken     = "token"
	headerRequestId = "request-id"
)

// IInboundMessageListener 接收 webhook 回调解码后的消息。
// 与 WebSocket 端 IMessageListener 不同：Bot API Gateway 的 webhook 回调只推送新消息，
// 不涉及撤回/编辑/已读等事件，因此这里只暴露 OnMessageReceive。
type IInboundMessageListener interface {
	OnMessageReceive(msg *models.Message)
}

// Client 是 Webhook 形态的 Bot 客户端：
//   - 主动能力（发消息、查用户、配置 webhook）通过 HTTP 调用 botapigateway；
//   - 被动接收消息通过内置 HTTP server 接收 botmsg 的回调。
//
// 它与 imbotclients(WebSocket) 完全独立，互不依赖。
type Client struct {
	// BaseURL 为 botapigateway 的基础地址，例如 http://127.0.0.1:8080 。
	// SDK 会自动拼接路由前缀（默认 botgateway）。
	BaseURL string
	AppKey  string
	Token   string
	// Prefix 为网关路由前缀，默认 botgateway，可按部署情况覆盖。
	Prefix string

	// ApiKey 同时用于两处：SetWebhook 时上报给服务端，以及校验回调请求头
	// Authorization: Bearer <ApiKey>。为空时不校验。
	ApiKey string

	// HTTPClient 供主动请求使用，nil 时使用默认（10s 超时）。
	HTTPClient *http.Client

	// UserId 在首次成功调用 QryUserInfo 后写入（Bot 自身的 userId）。
	UserId string

	listeners    []IInboundMessageListener
	onInboundRaw func(*InboundMessage)
	listenerLock sync.RWMutex
	server       *http.Server
	serverLock   sync.Mutex
}

// NewImBotWebhookClient 创建一个 Webhook 形态的 Bot 客户端。
// baseURL 为网关基础地址，appKey、token 用于鉴权；无需再调用 Connect。
func NewImBotWebhookClient(baseURL, appKey, token string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		AppKey:  appKey,
		Token:   token,
		Prefix:  defaultPrefix,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// AddMessageListener 注册回调消息监听器。
func (c *Client) AddMessageListener(listener IInboundMessageListener) {
	if listener == nil {
		return
	}
	c.listenerLock.Lock()
	c.listeners = append(c.listeners, listener)
	c.listenerLock.Unlock()
}

// SetInboundRawCallback 注册底层回调，拿到未解码的原始请求体。
func (c *Client) SetInboundRawCallback(f func(*InboundMessage)) {
	c.listenerLock.Lock()
	c.onInboundRaw = f
	c.listenerLock.Unlock()
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

func (c *Client) url(path string) string {
	prefix := c.Prefix
	if prefix == "" {
		prefix = defaultPrefix
	}
	return fmt.Sprintf("%s/%s/%s", c.BaseURL, prefix, strings.TrimLeft(path, "/"))
}

// doJSON 发送一个 JSON 请求并把响应 data 解析到 out（out 可为 nil）。
func (c *Client) doJSON(method, path string, body interface{}, out interface{}) utils.ClientErrorCode {
	var reqBody io.Reader
	if body != nil {
		bs, err := json.Marshal(body)
		if err != nil {
			return utils.ClientErrorCode_Unknown
		}
		reqBody = bytes.NewReader(bs)
	}
	req, err := http.NewRequest(method, c.url(path), reqBody)
	if err != nil {
		return utils.ClientErrorCode_SocketFailed
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(headerAppKey, c.AppKey)
	req.Header.Set(headerToken, c.Token)

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return utils.ClientErrorCode_SocketFailed
	}
	defer resp.Body.Close()
	respBs, err := io.ReadAll(resp.Body)
	if err != nil {
		return utils.ClientErrorCode_SocketFailed
	}
	envelope := &apiResp{}
	if out != nil {
		envelope.Data = out
	}
	if err := json.Unmarshal(respBs, envelope); err != nil {
		return utils.ClientErrorCode_Unknown
	}
	if envelope.Code != 0 {
		return utils.ClientErrorCode(envelope.Code)
	}
	return utils.ClientErrorCode_Success
}
