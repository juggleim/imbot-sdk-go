package messages

import "github.com/juggleim/imbot-sdk-go/models"

// DecodeContent 根据消息类型把原始 byte 解码为对应的 MessageContentInterface。
// 无法识别的类型或解码失败时回退为 UnknownMessage，保留原始数据。
// WebSocket 与 Webhook 两种传输方式共用此解码逻辑，保证收到的消息结构一致。
func DecodeContent(msgType string, data []byte) models.MessageContentInterface {
	var content models.MessageContentInterface
	switch msgType {
	case MessageContentTypeText:
		content = NewTextMessage("")
	case MessageContentTypeImage:
		content = NewImageMessage()
	case MessageContentTypeFile:
		content = NewFileMessage()
	case MessageContentTypeVideo:
		content = NewVideoMessage()
	case MessageContentTypeVoice:
		content = NewVoiceMessage()
	case MessageContentTypeStreamText:
		content = NewStreamTextMessage()
	case MessageContentTypeRecallInfo:
		content = NewRecallInfoMessage()
	case MessageContentTypeMerge:
		content = NewMergeMessage("", nil, nil, nil)
	case MessageContentTypeThumbnailPackedImage:
		content = NewThumbnailPackedImageMessage()
	case MessageContentTypeSnapshotPackedVideo:
		content = NewSnapshotPackedVideoMessage()
	default:
		unknown := NewUnknownMessage(msgType)
		unknown.Content = string(data)
		return unknown
	}
	if err := content.Decode(data); err != nil {
		unknown := NewUnknownMessage(msgType)
		unknown.Content = string(data)
		return unknown
	}
	return content
}
