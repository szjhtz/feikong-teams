package protocol

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	DefaultBaseURL = "https://ilinkai.weixin.qq.com"
	CDNBaseURL     = "https://novac2c.cdn.weixin.qq.com/c2c"
	ChannelVersion = "2.0.0"

	maxAPIResponseBytes = 16 << 20
	maxAPIErrorBytes    = 4 << 10
)

// APIError 表示 iLink API 返回的业务或 HTTP 错误。
type APIError struct {
	Message    string
	HTTPStatus int
	ErrCode    int
}

func (e *APIError) Error() string {
	return fmt.Sprintf("ilink api: %s (http=%d, errcode=%d)", e.Message, e.HTTPStatus, e.ErrCode)
}

// IsSessionExpired 判断错误是否为会话过期。
func (e *APIError) IsSessionExpired() bool {
	return e.ErrCode == -14
}

// RandomWechatUIN 生成 X-WECHAT-UIN 请求头。
func RandomWechatUIN() string {
	var buf [4]byte
	rand.Read(buf[:])
	val := binary.BigEndian.Uint32(buf[:])
	return base64.StdEncoding.EncodeToString([]byte(strconv.FormatUint(uint64(val), 10)))
}

// AuthHeaders 返回 iLink POST 请求头。
func AuthHeaders(token string) http.Header {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	h.Set("AuthorizationType", "ilink_bot_token")
	h.Set("Authorization", "Bearer "+token)
	h.Set("X-WECHAT-UIN", RandomWechatUIN())
	return h
}

func baseInfo() map[string]string {
	return map[string]string{"channel_version": ChannelVersion}
}

// Client 封装 iLink API 请求。
type Client struct {
	HTTP *http.Client
}

// NewClient 创建协议客户端。
func NewClient() *Client {
	return &Client{
		HTTP: &http.Client{Timeout: 45 * time.Second},
	}
}

// QRCodeResponse 是 get_bot_qrcode 响应。
type QRCodeResponse struct {
	QRCode       string `json:"qrcode"`
	QRCodeImgURL string `json:"qrcode_img_content"`
}

// QRStatusResponse 是 get_qrcode_status 响应。
type QRStatusResponse struct {
	Status   string `json:"status"` // wait/scaned/confirmed/expired
	BotToken string `json:"bot_token,omitempty"`
	BotID    string `json:"ilink_bot_id,omitempty"`
	UserID   string `json:"ilink_user_id,omitempty"`
	BaseURL  string `json:"baseurl,omitempty"`
}

// GetUpdatesResponse 是 getupdates 响应。
type GetUpdatesResponse struct {
	Ret           int               `json:"ret"`
	Msgs          []json.RawMessage `json:"msgs"`
	GetUpdatesBuf string            `json:"get_updates_buf"`
	ErrCode       int               `json:"errcode,omitempty"`
	ErrMsg        string            `json:"errmsg,omitempty"`
}

// GetConfigResponse 是 getconfig 响应。
type GetConfigResponse struct {
	TypingTicket string `json:"typing_ticket,omitempty"`
	Ret          int    `json:"ret,omitempty"`
}

// GetQRCode 请求登录二维码。
func (c *Client) GetQRCode(ctx context.Context, baseURL string) (*QRCodeResponse, error) {
	u := baseURL + "/ilink/bot/get_bot_qrcode?bot_type=3"
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("get_bot_qrcode request: %w", err)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get_bot_qrcode: %w", err)
	}
	defer resp.Body.Close()
	var result QRCodeResponse
	if err := decodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("get_bot_qrcode: %w", err)
	}
	return &result, nil
}

// PollQRStatus 轮询二维码扫码状态。
func (c *Client) PollQRStatus(ctx context.Context, baseURL, qrcode string) (*QRStatusResponse, error) {
	u := baseURL + "/ilink/bot/get_qrcode_status?qrcode=" + url.QueryEscape(qrcode)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("get_qrcode_status request: %w", err)
	}
	req.Header.Set("iLink-App-ClientVersion", "1")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get_qrcode_status: %w", err)
	}
	defer resp.Body.Close()
	var result QRStatusResponse
	if err := decodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("get_qrcode_status: %w", err)
	}
	return &result, nil
}

// apiPost 发送 iLink POST 请求并解析响应。
func (c *Client) apiPost(ctx context.Context, baseURL, endpoint, token string, body any, timeout time.Duration) (json.RawMessage, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("%s encode request: %w", endpoint, err)
	}
	u := baseURL + endpoint
	httpCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(httpCtx, "POST", u, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("%s create request: %w", endpoint, err)
	}
	for k, v := range AuthHeaders(token) {
		req.Header[k] = v
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	raw, err := readLimitedBody(resp.Body, maxAPIResponseBytes)
	if err != nil {
		return nil, fmt.Errorf("%s read response: %w", endpoint, err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, responseAPIError(resp.StatusCode, raw)
	}

	var check struct {
		Ret     int    `json:"ret"`
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(raw, &check); err != nil {
		return nil, fmt.Errorf("%s decode response: %w", endpoint, err)
	}
	if check.Ret != 0 {
		code := check.ErrCode
		if code == 0 {
			code = check.Ret
		}
		msg := check.ErrMsg
		if msg == "" {
			msg = fmt.Sprintf("ret=%d", check.Ret)
		}
		return nil, &APIError{Message: msg, HTTPStatus: resp.StatusCode, ErrCode: code}
	}

	return json.RawMessage(raw), nil
}

func decodeResponse(resp *http.Response, target any) error {
	raw, err := readLimitedBody(resp.Body, maxAPIResponseBytes)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return responseAPIError(resp.StatusCode, raw)
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func readLimitedBody(reader io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("response exceeds %d bytes", limit)
	}
	return data, nil
}

func responseAPIError(status int, raw []byte) *APIError {
	var payload struct {
		Ret     int    `json:"ret"`
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	_ = json.Unmarshal(raw, &payload)
	code := payload.ErrCode
	if code == 0 {
		code = payload.Ret
	}
	message := payload.ErrMsg
	if message == "" {
		message = string(raw)
	}
	if len(message) > maxAPIErrorBytes {
		message = message[:maxAPIErrorBytes] + "..."
	}
	if message == "" {
		message = http.StatusText(status)
	}
	return &APIError{Message: message, HTTPStatus: status, ErrCode: code}
}

// GetUpdates 长轮询获取新消息。
func (c *Client) GetUpdates(ctx context.Context, baseURL, token, cursor string) (*GetUpdatesResponse, error) {
	body := map[string]any{
		"get_updates_buf": cursor,
		"base_info":       baseInfo(),
	}
	raw, err := c.apiPost(ctx, baseURL, "/ilink/bot/getupdates", token, body, 45*time.Second)
	if err != nil {
		return nil, err
	}
	var result GetUpdatesResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("getupdates decode: %w", err)
	}
	return &result, nil
}

// SendMessage 通过 iLink API 发送消息。
func (c *Client) SendMessage(ctx context.Context, baseURL, token string, msg any) error {
	body := map[string]any{
		"msg":       msg,
		"base_info": baseInfo(),
	}
	_, err := c.apiPost(ctx, baseURL, "/ilink/bot/sendmessage", token, body, 15*time.Second)
	return err
}

// GetConfig 获取用户的 typing ticket。
func (c *Client) GetConfig(ctx context.Context, baseURL, token, userID, contextToken string) (*GetConfigResponse, error) {
	body := map[string]any{
		"ilink_user_id": userID,
		"context_token": contextToken,
		"base_info":     baseInfo(),
	}
	raw, err := c.apiPost(ctx, baseURL, "/ilink/bot/getconfig", token, body, 15*time.Second)
	if err != nil {
		return nil, err
	}
	var result GetConfigResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("getconfig decode: %w", err)
	}
	return &result, nil
}

// SendTyping 发送或取消输入中状态。
func (c *Client) SendTyping(ctx context.Context, baseURL, token, userID, ticket string, status int) error {
	body := map[string]any{
		"ilink_user_id": userID,
		"typing_ticket": ticket,
		"status":        status,
		"base_info":     baseInfo(),
	}
	_, err := c.apiPost(ctx, baseURL, "/ilink/bot/sendtyping", token, body, 15*time.Second)
	return err
}

// GetUploadURLRequest 是 getuploadurl 请求参数。
type GetUploadURLRequest struct {
	FileKey     string `json:"filekey"`
	MediaType   int    `json:"media_type"`
	ToUserID    string `json:"to_user_id"`
	RawSize     int    `json:"rawsize"`
	RawFileMD5  string `json:"rawfilemd5"`
	FileSize    int    `json:"filesize"`
	NoNeedThumb bool   `json:"no_need_thumb"`
	AESKey      string `json:"aeskey,omitempty"`
}

// GetUploadURLResponse 是 getuploadurl 响应。
type GetUploadURLResponse struct {
	UploadParam string `json:"upload_param"`
}

// GetUploadURL 请求 CDN 上传地址。
func (c *Client) GetUploadURL(ctx context.Context, baseURL, token string, req GetUploadURLRequest) (*GetUploadURLResponse, error) {
	body := map[string]any{
		"filekey":       req.FileKey,
		"media_type":    req.MediaType,
		"to_user_id":    req.ToUserID,
		"rawsize":       req.RawSize,
		"rawfilemd5":    req.RawFileMD5,
		"filesize":      req.FileSize,
		"no_need_thumb": req.NoNeedThumb,
		"aeskey":        req.AESKey,
		"base_info":     baseInfo(),
	}
	raw, err := c.apiPost(ctx, baseURL, "/ilink/bot/getuploadurl", token, body, 15*time.Second)
	if err != nil {
		return nil, err
	}
	var result GetUploadURLResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("getuploadurl decode: %w", err)
	}
	return &result, nil
}

// BuildMediaMessage 构建媒体消息载荷。
func BuildMediaMessage(userID, contextToken string, itemList []map[string]any) map[string]any {
	return map[string]any{
		"from_user_id":  "",
		"to_user_id":    userID,
		"client_id":     newUUID(),
		"message_type":  2,
		"message_state": 2,
		"context_token": contextToken,
		"item_list":     itemList,
	}
}

// BuildTextMessage 构建文本消息载荷。
func BuildTextMessage(userID, contextToken, text string) map[string]any {
	return map[string]any{
		"from_user_id":  "",
		"to_user_id":    userID,
		"client_id":     newUUID(),
		"message_type":  2,
		"message_state": 2,
		"context_token": contextToken,
		"item_list": []map[string]any{
			{"type": 1, "text_item": map[string]string{"text": text}},
		},
	}
}

func newUUID() string {
	var buf [16]byte
	rand.Read(buf[:])
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:16])
}
