package messages

import "github.com/juggleim/imbot-sdk-go/models"

const (
	MessageContentTypeText                 = "jg:text"
	MessageContentTypeImage                = "jg:img"
	MessageContentTypeFile                 = "jg:file"
	MessageContentTypeVideo                = "jg:video"
	MessageContentTypeVoice                = "jg:voice"
	MessageContentTypeStreamText           = "jg:streamtext"
	MessageContentTypeRecallInfo           = "jg:recallinfo"
	MessageContentTypeMerge                = "jg:merge"
	MessageContentTypeThumbnailPackedImage = "jg:tpimg"
	MessageContentTypeSnapshotPackedVideo  = "jg:spvideo"
)

func newMediaMessageContent(msgType string) models.MediaMessageContent {
	return models.MediaMessageContent{MessageContent: models.NewMessageContent(msgType)}
}
