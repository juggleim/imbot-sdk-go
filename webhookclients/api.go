package webhookclients

import (
	"net/http"

	"github.com/juggleim/imbot-sdk-go/utils"
)

// SendMessageReq 直接以网关请求体发送，适合不依赖 pbobjs 的调用方。
func (c *Client) SendMessage(req *SendMessageReq) (utils.ClientErrorCode, *SendMessageResp) {
	if req == nil || req.Conversation == nil || req.Conversation.ConversationId == "" {
		return utils.ClientErrorCode_Unknown, nil
	}
	resp := &SendMessageResp{}
	code := c.doJSON(http.MethodPost, "messages/send", req, resp)
	if code != utils.ClientErrorCode_Success {
		return code, nil
	}
	return code, resp
}

// QryUserInfo 查询用户资料；userId 为空时查询 Bot 自身，并把结果写入 c.UserId。
func (c *Client) QryUserInfo(userId string) (utils.ClientErrorCode, *UserInfo) {
	path := "bots/info"
	if userId != "" {
		path = "bots/info?user_id=" + userId
	}
	resp := &UserInfo{}
	code := c.doJSON(http.MethodGet, path, nil, resp)
	if code != utils.ClientErrorCode_Success {
		return code, nil
	}
	if userId == "" && resp.UserId != "" {
		c.UserId = resp.UserId
	}
	return code, resp
}

// SetWebhook 配置当前 Bot 的回调地址（POST /bots/webhook/set）。
// 调用后会把 apiKey 记到 c.ApiKey，用于校验后续回调请求的 Authorization。
func (c *Client) SetWebhook(url, apiKey string, isStream bool) utils.ClientErrorCode {
	code := c.doJSON(http.MethodPost, "bots/webhook/set", &SetWebhookReq{
		Url:      url,
		ApiKey:   apiKey,
		IsStream: isStream,
	}, nil)
	if code == utils.ClientErrorCode_Success {
		c.ApiKey = apiKey
	}
	return code
}

// GetWebhook 查询当前 Bot 的回调配置（GET /bots/webhook/get）。
func (c *Client) GetWebhook() (utils.ClientErrorCode, *Webhook) {
	resp := &Webhook{}
	code := c.doJSON(http.MethodGet, "bots/webhook/get", nil, resp)
	if code != utils.ClientErrorCode_Success {
		return code, nil
	}
	return code, resp
}

// DelWebhook 删除当前 Bot 的回调配置（POST /bots/webhook/del）。
func (c *Client) DelWebhook() utils.ClientErrorCode {
	return c.doJSON(http.MethodPost, "bots/webhook/del", nil, nil)
}
