package webhookclients

// 本文件中的结构体与服务端 botapigateway 的 HTTP 请求/响应体一一对应，
// 字段顺序、json tag 均对齐 services/botapigateway/models 下的定义。

// apiResp 是 botapigateway 统一的响应外层结构（tools.SuccessHttpResp / ErrorHttpResp）。
type apiResp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// Conversation 标识一个会话；ConversationType 即 channelType。
type Conversation struct {
	ConversationType int    `json:"conversation_type"`
	ConversationId   string `json:"conversation_id"`
}

// MentionInfo 对应 POST /messages/send 中的 @ 信息。
type MentionInfo struct {
	MentionType   string   `json:"mention_type"`
	TargetUserIds []string `json:"target_user_ids"`
}

// ReferMsg 引用消息。
type ReferMsg struct {
	MsgId       string `json:"msg_id"`
	SenderId    string `json:"sender_id"`
	TargetId    string `json:"target_id"`
	ChannelType int    `json:"channel_type"`
	MsgType     string `json:"msg_type"`
	MsgTime     int64  `json:"msg_time"`
	MsgContent  string `json:"msg_content"`
}

// PushData 推送设置。
type PushData struct {
	PushTitle string `json:"push_title"`
	PushText  string `json:"push_text"`
	PushExtra string `json:"push_extra"`
	PushLevel int    `json:"push_level"`
}

// SendMessageReq 对应 POST /messages/send 的请求体。
type SendMessageReq struct {
	Conversation *Conversation `json:"conversation"`
	MsgType      string        `json:"msg_type"`
	MsgContent   string        `json:"msg_content"`

	IsStorage *bool `json:"is_storage,omitempty"`
	IsCount   *bool `json:"is_count,omitempty"`
	IsState   *bool `json:"is_state,omitempty"`
	IsCmd     *bool `json:"is_cmd,omitempty"`

	MentionInfo *MentionInfo `json:"mention_info,omitempty"`
	ReferMsg    *ReferMsg    `json:"refer_msg,omitempty"`
	PushData    *PushData    `json:"push_data,omitempty"`
	ToUserIds   []string     `json:"to_user_ids,omitempty"`

	LifeTime          int64 `json:"life_time,omitempty"`
	LifeTimeAfterRead int64 `json:"life_time_after_read,omitempty"`

	MsgId *string `json:"msg_id,omitempty"`
}

// SendMessageResp 对应 POST /messages/send 的响应 data。
type SendMessageResp struct {
	MsgId string `json:"msg_id"`
}

// UserInfo 对应 GET /bots/info 的响应 data。
type UserInfo struct {
	UserId       string            `json:"user_id"`
	Nickname     string            `json:"nickname"`
	UserPortrait string            `json:"user_portrait"`
	ExtFields    map[string]string `json:"ext_fields"`
	UpdatedTime  int64             `json:"updated_time"`
}

// SetWebhookReq 对应 POST /bots/webhook/set 的请求体。
type SetWebhookReq struct {
	Url      string `json:"url"`
	ApiKey   string `json:"api_key,omitempty"`
	IsStream bool   `json:"is_stream,omitempty"`
}

// Webhook 对应 GET /bots/webhook/get 的响应 data。
type Webhook struct {
	Url      string `json:"url"`
	ApiKey   string `json:"api_key"`
	IsStream bool   `json:"is_stream"`
}

// InboundMessage 是服务端 botmsg CustomBotEngine 回调到 Bot webhook 的请求体，
// 字段对齐 services/botmsg/services/botengines.CustomChatMsgReq。
type InboundMessage struct {
	AppKey      string       `json:"app_key"`
	Sender      string       `json:"sender"`
	Receiver    string       `json:"receiver"`
	ConverType  int          `json:"conver_type"`
	MsgType     string       `json:"msg_type"`
	MsgContent  string       `json:"msg_content"`
	MsgId       string       `json:"msg_id"`
	MsgTime     int64        `json:"msg_time"`
	MentionInfo *MentionInfo `json:"mention_info,omitempty"`
}
