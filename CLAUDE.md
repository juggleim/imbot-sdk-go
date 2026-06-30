# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`imbot-sdk-go` is the Go SDK for [JuggleIM](https://github.com/juggleim). It lets bots, server-side message proxies, or any Go program actively send and receive IM messages. It offers two **independent** transports, each its own package — they do not depend on each other:

- **WebSocket** (`imbotclients/`) — a long-lived WebSocket connection. `imbotclients.NewImBotClient(address, appkey)` then `client.Connect(token)`.
- **Webhook** (`webhookclients/`) — HTTP calls to the server's `botapigateway` for outbound (send / query / webhook config), plus a local HTTP server that receives inbound message callbacks. `webhookclients.NewImBotWebhookClient(baseURL, appKey, token)`; no Connect step.

Code comments and the README are in Chinese; match that when editing existing files.

## Commands

```bash
go build ./...        # compile everything
go vet ./...          # static checks
go run .              # run the demo bot in main.go (edit address/appkey/token first)
```

There are no tests in this repo yet. The default IM address in `main.go` is `ws://127.0.0.1:9002`; the SDK appends the `/imbot` path automatically.

## Architecture

### Transport core (`imbotclients/imbotclient.go`)
`ImBotClient` wraps a `gorilla/websocket` connection and a custom binary protocol. All wire messages are `pbobjs.ImWebsocketMsg` envelopes distinguished by `Cmd` (connect/publish/query/ping + their acks). The four building blocks every higher-level method goes through:

- **`Publish(topic, targetId, data)`** — fire a message that needs a `PublishAckMsgBody` (used for sending messages).
- **`Query(method, targetId, data)`** — request/response that returns a `QueryAckMsgBody` (used for all queries).
- **`Ping` / heartbeat** — liveness.

Request/response correlation is done with a monotonic `myIndex` and an `accssorCache` (`sync.Map` of index → `utils.DataAccessor`). A call stores a `DataAccessor`, writes the frame, then blocks on `GetWithTimeout`; the read loop's `OnPublishAck`/`OnQueryAck` looks the index up and unblocks it. `utils.DataAccessor` (`utils/dataaccessor.go`) is the one-shot future used for this.

The single read loop is `startListener` → `handleMsg`; each inbound frame is dispatched to a `handleMsg` go routine. Inbound `cmdPublish` frames carry a `topic` (`"msg"`, `"ntf"`, `"stream_msg"`): `"ntf"` of type `NotifyType_Msg` triggers a `SyncMsgs` pull loop driven by `inboxTime`/`sendboxTime` watermarks.

### Connection lifecycle (`imbotclients/connect.go`, `heartbeat_reconnect.go`)
`Connect` → dials, sends `ConnectMsgBody`, waits on `connAckAccessor`. `Disconnect`/`Logout` set `suppressAutoReconnect` so the read-loop teardown won't reconnect. Heartbeat and exponential reconnect backoff are deliberately ported to match the JuggleIM Android client (`HeartbeatManager` / `IntervalGenerator` — see the Chinese comments); preserve that parity when touching timing constants.

### Feature methods (one file per domain in `imbotclients/`)
`message.go`, `historymsg.go`, `conversation.go`, `chatroom.go`, `user.go`, `group.go`, `rtcroom.go`, `file.go`. These are thin: build a `pbobjs` request, call `Publish`/`Query`, unmarshal the ack `.Data` into the response proto, and translate codes via `trans2ClientErrorCode` / `ackCode`. `SendMessage` picks the topic (`p_msg`/`g_msg`/`c_msg`/`pc_msg`) from `Conversation.ConversationType`.

### Listeners / callbacks
Two layers for receiving data:
- **High-level listeners** (preferred): `AddMessageListener`, `AddConnectionStatusChangeListener`, `AddConversationChangeListener`. Message listeners receive decoded `*models.Message`.
- **Low-level callbacks**: `OnMessageCallBack`/`OnStreamMsgCallBack` give raw `pbobjs.DownMsg`; `DisconnectCallback` for disconnect reasons.

### Message content model (`models/`, `models/messages/`)
`models.Message` carries a `MsgContent models.MessageContentInterface`. Each content type (text, image, file, video, voice, streamtext, merge, recallinfo, thumbnail/snapshot packed media) lives in `models/messages/` and implements `Encode`/`Decode` (JSON) plus a `MessageContentType*` string constant defined in `models/messages/common.go`.

**Adding a new message content type requires two edits:** add the type in `models/messages/` (with its constant), AND register it in the `messages.DecodeContent` switch in `models/messages/decode.go` — otherwise inbound messages of that type fall back to `UnknownMessage`. (Both transports share this decoder.) Sending is manual: call `content.Encode()`, then build a `pbobjs.UpMsg` with `MsgType`, `MsgContent`, `Flags`, and a unique `ClientUid`.

### Webhook transport (`webhookclients/`)
A standalone package with no dependency on `imbotclients`. `Client` has two halves:
- **Outbound HTTP** (`api.go`, `client.go`): `SendMessage`, `QryUserInfo`, `SetWebhook`/`GetWebhook`/`DelWebhook`. All go through `doJSON`, which targets `<BaseURL>/<Prefix>/...` (default prefix `botgateway`, matching the server's `botApiRouters.Route(engine, "botgateway")`) and sends `appkey`/`token` headers. The server wraps responses as `{code,msg,data}`; `code != 0` maps to `utils.ClientErrorCode`.
- **Inbound HTTP** (`receiver.go`): an HTTP server (`StartReceiver`, or embed `Handler()`) that receives the server's `botmsg` callback (`InboundMessage`, mirroring `CustomChatMsgReq`), verifies `Authorization: Bearer <ApiKey>`, decodes into a `*models.Message`, and fans out to `IInboundMessageListener`s.

Request/response DTOs in `webhookclients/models.go` mirror `services/botapigateway/models` field-for-field. `SendMessage` accepts the same `(conver, *pbobjs.UpMsg)` signature as the WebSocket client; `upMsgToSendReq` splits `UpMsg.Flags` into the gateway's `is_storage`/`is_count`/`is_cmd`/`is_state` booleans (bit values must stay in sync with the server's `msgdefines/msgflag.go`: 1=cmd, 2=count, 4=state, 8=store). The callback payload carries no group id (`receiver` is the bot itself), so the decoded `Conversation` uses the sender — fine for 1:1 replies; group replies must set the conversation explicitly when calling `SendMessage`.

### Generated protobuf (`imbotclients/pbdefines/`)
`.proto` sources live in `imbotclients/pbdefines/`; generated Go lands in `imbotclients/pbdefines/pbobjs/` (package `pbobjs`, `// DO NOT EDIT`). Regenerate from the repo root, e.g.:

```bash
protoc --go_out=. --go_opt=paths=source_relative imbotclients/pbdefines/*.proto
```

Never hand-edit `*.pb.go`.

### Error codes
All public methods return a `utils.ClientErrorCode` (`utils/consts.go`). `ClientErrorCode_Success == 0`. SDK-level failures use the `2xxxx` range; server-side codes pass through unchanged via `trans2ClientErrorCode`.
