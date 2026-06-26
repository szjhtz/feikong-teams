// Package wechatbot 提供微信 iLink Bot SDK。
package wechatbot

import "time"

// MessageType 表示消息发送方。
type MessageType int

const (
	MessageTypeUser MessageType = 1
	MessageTypeBot  MessageType = 2
)

// MessageState 表示消息投递状态。
type MessageState int

const (
	MessageStateNew        MessageState = 0
	MessageStateGenerating MessageState = 1
	MessageStateFinish     MessageState = 2
)

// MessageItemType 表示消息条目类型。
type MessageItemType int

const (
	ItemText  MessageItemType = 1
	ItemImage MessageItemType = 2
	ItemVoice MessageItemType = 3
	ItemFile  MessageItemType = 4
	ItemVideo MessageItemType = 5
)

// MediaType 表示上传媒体类型。
type MediaType int

const (
	MediaImage MediaType = 1
	MediaVideo MediaType = 2
	MediaFile  MediaType = 3
	MediaVoice MediaType = 4
)

type BaseInfo struct {
	ChannelVersion string `json:"channel_version"`
}

// CDNMedia 表示微信 CDN 上的加密媒体。
type CDNMedia struct {
	EncryptQueryParam string `json:"encrypt_query_param"`
	AESKey            string `json:"aes_key"`
	EncryptType       int    `json:"encrypt_type,omitempty"`
}

type TextItem struct {
	Text string `json:"text"`
}

// ImageItem 表示图片内容。
type ImageItem struct {
	Media       *CDNMedia `json:"media,omitempty"`
	ThumbMedia  *CDNMedia `json:"thumb_media,omitempty"`
	AESKey      string    `json:"aeskey,omitempty"`
	URL         string    `json:"url,omitempty"`
	MidSize     int64     `json:"mid_size,omitempty"`
	ThumbSize   int64     `json:"thumb_size,omitempty"`
	ThumbWidth  int       `json:"thumb_width,omitempty"`
	ThumbHeight int       `json:"thumb_height,omitempty"`
	HDSize      int64     `json:"hd_size,omitempty"`
}

// VoiceItem 表示语音内容。
type VoiceItem struct {
	Media      *CDNMedia `json:"media,omitempty"`
	EncodeType int       `json:"encode_type,omitempty"`
	Text       string    `json:"text,omitempty"`
	Playtime   int       `json:"playtime,omitempty"`
}

// FileItem 表示文件内容。
type FileItem struct {
	Media    *CDNMedia `json:"media,omitempty"`
	FileName string    `json:"file_name,omitempty"`
	MD5      string    `json:"md5,omitempty"`
	Len      string    `json:"len,omitempty"`
}

// VideoItem 表示视频内容。
type VideoItem struct {
	Media      *CDNMedia `json:"media,omitempty"`
	VideoSize  int64     `json:"video_size,omitempty"`
	PlayLength int       `json:"play_length,omitempty"`
	ThumbMedia *CDNMedia `json:"thumb_media,omitempty"`
}

// RefMessage 表示引用消息。
type RefMessage struct {
	Title       string       `json:"title,omitempty"`
	MessageItem *MessageItem `json:"message_item,omitempty"`
}

// MessageItem 表示单个消息条目。
type MessageItem struct {
	Type      MessageItemType `json:"type"`
	TextItem  *TextItem       `json:"text_item,omitempty"`
	ImageItem *ImageItem      `json:"image_item,omitempty"`
	VoiceItem *VoiceItem      `json:"voice_item,omitempty"`
	FileItem  *FileItem       `json:"file_item,omitempty"`
	VideoItem *VideoItem      `json:"video_item,omitempty"`
	RefMsg    *RefMessage     `json:"ref_msg,omitempty"`
}

// WireMessage 是 iLink API 返回的原始消息。
type WireMessage struct {
	Seq          int64         `json:"seq,omitempty"`
	MessageID    int64         `json:"message_id,omitempty"`
	FromUserID   string        `json:"from_user_id"`
	ToUserID     string        `json:"to_user_id"`
	ClientID     string        `json:"client_id"`
	CreateTimeMs int64         `json:"create_time_ms"`
	MessageType  MessageType   `json:"message_type"`
	MessageState MessageState  `json:"message_state"`
	ContextToken string        `json:"context_token"`
	ItemList     []MessageItem `json:"item_list"`
}

// ContentType 表示收到消息的主类型。
type ContentType string

const (
	ContentText  ContentType = "text"
	ContentImage ContentType = "image"
	ContentVoice ContentType = "voice"
	ContentFile  ContentType = "file"
	ContentVideo ContentType = "video"
)

// IncomingMessage 是解析后的消息。
type IncomingMessage struct {
	UserID        string
	Text          string
	Type          ContentType
	Timestamp     time.Time
	Images        []ImageContent
	Voices        []VoiceContent
	Files         []FileContent
	Videos        []VideoContent
	QuotedMessage *QuotedMessage
	Raw           *WireMessage
	ContextToken  string // SDK 内部维护
}

// ImageContent 表示解析后的图片内容。
type ImageContent struct {
	Media      *CDNMedia
	ThumbMedia *CDNMedia
	AESKey     string
	URL        string
	Width      int
	Height     int
}

// VoiceContent 表示解析后的语音内容。
type VoiceContent struct {
	Media      *CDNMedia
	Text       string
	DurationMs int
	EncodeType int
}

// FileContent 表示解析后的文件内容。
type FileContent struct {
	Media    *CDNMedia
	FileName string
	MD5      string
	Size     int64
}

// VideoContent 表示解析后的视频内容。
type VideoContent struct {
	Media      *CDNMedia
	ThumbMedia *CDNMedia
	DurationMs int
	Width      int
	Height     int
}

// QuotedMessage 表示被引用的消息。
type QuotedMessage struct {
	Title string
	Text  string
	Type  ContentType
}

// DownloadedMedia 是媒体下载结果。
type DownloadedMedia struct {
	Data     []byte
	Type     string // image/file/video/voice
	FileName string
	Format   string // 语音格式，如 silk
}

// UploadResult 是媒体上传结果。
type UploadResult struct {
	Media             CDNMedia
	AESKey            []byte
	EncryptedFileSize int
}

// Credentials 表示登录凭证。
type Credentials struct {
	Token     string `json:"token"`
	BaseURL   string `json:"baseUrl"`
	AccountID string `json:"accountId"`
	UserID    string `json:"userId"`
	SavedAt   string `json:"savedAt,omitempty"`
}
