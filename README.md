# imbot-sdk-go

`imbot-sdk-go` 是 JuggleIM 的 Go SDK，基于 WebSocket 与 IM 服务建立长连接，适合 Bot、服务端消息代理或需要主动收发消息的 Go 程序。

当前代码已经覆盖了连接管理、消息收发、会话查询、历史消息、聊天室、用户状态、RTC 房间和文件凭证查询等常用能力。

## 安装

```bash
go get github.com/juggleim/imbot-sdk-go
```

## 能力概览

- 连接管理：`Connect`、`Disconnect`、`Logout`、心跳保活、断线自动重连
- 消息收发：单聊、群聊、聊天室、公共频道消息发送与接收
- 消息管理：历史消息查询、撤回、修改、已读标记、搜索、置顶消息
- 会话管理：会话列表、未读数、置顶、免打扰、标签
- 用户与群信息：用户资料、好友资料、群信息、在线状态订阅
- 聊天室：加入/退出聊天室、聊天室消息、聊天室属性同步
- RTC：创建/加入/查询/退出 RTC 房间
- 文件：文件凭证查询 `GetFileCred`

## 快速说明

### 1. 创建客户端

```go
client := imbotclients.NewImBotClient("ws://127.0.0.1:9002", "your-appkey")
```

说明：

- `address` 传基础地址即可，SDK 会自动拼接为 `ws://host/imbot` 或 `wss://host/imbot`
- `Platform` 默认是 `Bot`
- `AutoReconnect` 默认是 `true`

### 2. 建立连接

```go
code, ack := client.Connect("your-token")
```

说明：

- 连接成功时 `code == utils.ClientErrorCode_Success`
- 成功后 `ack.UserId` 会写入 `client.UserId`
- `Connect` 只能在断开状态下调用；重复连接会返回 `ClientErrorCode_ConnectExisted`
- 主动调用 `Disconnect()` 或 `Logout()` 后，不会继续自动重连

### 3. 接收消息的两种方式

推荐优先使用高层监听：

- `AddMessageListener(listener)`：收到的是已经解码后的 `*models.Message`

如果你想拿到底层 protobuf 数据：

- `client.OnMessageCallBack = func(msg *pbobjs.DownMsg) {}`
- `client.OnStreamMsgCallBack = func(msg *pbobjs.StreamDownMsg) {}`

### 4. 发送消息

发送消息的核心方法是：

```go
code, ack := client.SendMessage(conversation, upMsg)
```

其中：

- `conversation` 用来指定目标会话和会话类型
- `upMsg` 是 protobuf 上行消息体
- 成功后可以从 `ack.MsgId`、`ack.MsgSeqNo` 中拿到服务端确认结果

## 收发消息示例

下面是一个完整的示例：连接 IM，监听文本消息，并向单聊会话发送一条文本消息。

```go
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/juggleim/imbot-sdk-go/imbotclients"
	"github.com/juggleim/imbot-sdk-go/imbotclients/pbdefines/pbobjs"
	"github.com/juggleim/imbot-sdk-go/models"
	"github.com/juggleim/imbot-sdk-go/models/messages"
	"github.com/juggleim/imbot-sdk-go/utils"
)

type connListener struct{}

func (connListener) OnStatusChange(status utils.ConnectState, code utils.ClientErrorCode) {
	log.Printf("connection status changed: status=%d code=%d", status, code)
}

type messageListener struct{}

func (messageListener) OnMessageReceive(msg *models.Message) {
	switch content := msg.MsgContent.(type) {
	case *messages.TextMessage:
		log.Printf(
			"received text message: from=%s target=%s msgId=%s content=%s",
			msg.SenderId,
			msg.Conversation.Conversation,
			msg.MsgId,
			content.Content,
		)
	default:
		log.Printf(
			"received message: from=%s target=%s msgId=%s type=%s",
			msg.SenderId,
			msg.Conversation.Conversation,
			msg.MsgId,
			msg.MsgType,
		)
	}
}

func (messageListener) OnMessageRecall(msg *models.Message) {}

func (messageListener) OnMessageUpdate(msg *models.Message) {}

func (messageListener) OnMessageDelete(conver *models.Conversation, msgIds []string) {}

func (messageListener) OnMessageClear(conver *models.Conversation, t int64, senderId string) {}

func (messageListener) OnMessageReactionAdd(conver *models.Conversation, reaction *models.MessageReaction) {}

func (messageListener) OnMessageReactionRemove(conver *models.Conversation, reaction *models.MessageReaction) {}

func (messageListener) OnMessageSetTop(message *models.Message, operatorId string, isTop bool) {}

func main() {
	address := "ws://127.0.0.1:9002"
	appKey := "your-appkey"
	token := "your-token"
	targetUserID := "target-user-id"

	client := imbotclients.NewImBotClient(address, appKey)
	client.AddConnectionStatusChangeListener(connListener{})
	client.AddMessageListener(messageListener{})

	client.DisconnectCallback = func(code utils.ClientErrorCode, disMsg *pbobjs.DisconnectMsgBody) {
		log.Printf("disconnect: code=%d ext=%s", code, disMsg.GetExt())
	}

	code, ack := client.Connect(token)
	if code != utils.ClientErrorCode_Success {
		log.Fatalf("connect failed: code=%d", code)
	}
	log.Printf("connected: userId=%s session=%s", ack.UserId, ack.Session)
	defer client.Disconnect()

	text := messages.NewTextMessage("hello from imbot-sdk-go")
	payload, err := text.Encode()
	if err != nil {
		log.Fatalf("encode message failed: %v", err)
	}

	conversation := &models.Conversation{
		ConversationType: pbobjs.ChannelType_Private,
		Conversation:     targetUserID,
	}

	upMsg := &pbobjs.UpMsg{
		MsgType:    text.GetContentType(),
		MsgContent: payload,
		Flags:      int32(text.GetFlags()),
		ClientUid:  fmt.Sprintf("bot-%d", time.Now().UnixNano()),
	}

	sendCode, sendAck := client.SendMessage(conversation, upMsg)
	if sendCode != utils.ClientErrorCode_Success {
		log.Fatalf("send message failed: code=%d", sendCode)
	}
	log.Printf("send success: msgId=%s seq=%d", sendAck.MsgId, sendAck.MsgSeqNo)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
```

补充说明：

- 单聊发送时使用 `pbobjs.ChannelType_Private`
- 群聊发送时改为 `pbobjs.ChannelType_Group`
- 聊天室消息可以直接调用 `SendChatroomMsg(chatroomId, upMsg)`
- 接收到的 `msg.MsgContent` 已经按消息类型解码，可以直接做类型断言

## 消息类型

SDK 当前内置了这些消息内容模型，位于 `models/messages`：

- 文本：`jg:text`
- 图片：`jg:img`
- 文件：`jg:file`
- 视频：`jg:video`
- 语音：`jg:voice`
- 流式文本：`jg:streamtext`
- 撤回通知：`jg:recallinfo`
- 合并消息：`jg:merge`
- 缩略图打包图片：`jg:tpimg`
- 快照打包视频：`jg:spvideo`

如果收到未内置支持的消息类型，SDK 会回落为 `*messages.UnknownMessage`。

## 常用 API

### 连接与状态

- `Connect(token string)`
- `Reconnect()`
- `Disconnect()`
- `Logout()`
- `Ping()`
- `AddConnectionStatusChangeListener(listener)`

### 消息

- `SendMessage(conversation, upMsg)`
- `QryHistoryMsgs(req)`
- `RecallMsg(req)`
- `ModifyMsg(req)`
- `MarkReadMsg(req)`
- `MsgSearch(req)`
- `MsgGlobalSearch(req)`
- `SetTopMsg(req)`
- `DelTopMsg(req)`

### 会话

- `GetConversation(req)`
- `GetConversations(req)`
- `SyncConversations(req)`
- `ClearUnreadCount(req)`
- `SetConversationTop(req)`
- `SetMute(req)`
- `DeleteConversations(req)`

### 用户、群组、状态

- `FetchUserInfo(userId)`
- `FetchGroupInfo(groupId)`
- `FetchFriendInfo(friendUserId)`
- `GetUserStatus(req)`
- `SubscribeUserStatus(req)`
- `UnsubscribeUserStatus(req)`

### 聊天室

- `JoinChatroom(chatroomId)`
- `QuitChatroom(chatroomId)`
- `SendChatroomMsg(chatroomId, upMsg)`
- `SetAttributes(chatroomId, attributes)`
- `RemoveAttributes(chatroomId, keys)`

### RTC

- `CreateRtcRoom(req)`
- `JoinRtcRoom(req)`
- `QryRtcRoom(roomId)`
- `QuitRtcRoom(roomId)`
- `RtcInvite(req)`

### 文件

- `GetFileCred(req)`

说明：当前 SDK 提供的是文件凭证查询能力，文件上传/下载流程需要由业务侧结合存储服务自行处理。

## 使用注意事项

- `Publish` 和 `Query` 都要求当前连接状态为 `connected`，否则会返回 `ClientErrorCode_ConnectClosed`
- 发送和查询默认等待 10 秒 ACK，超时分别返回 `ClientErrorCode_SendTimeout`、`ClientErrorCode_QueryTimeout`
- SDK 内部会自动处理心跳；如果连续超过两个心跳周期未收到下行数据，会断开并尝试自动重连
- 消息监听回调里尽量不要执行长时间阻塞操作，必要时请自行起 goroutine 或投递到工作队列
- 图片、文件、语音、视频等消息体只负责内容编解码，不负责上传文件本身

## 许可证

[LICENSE](LICENSE)
