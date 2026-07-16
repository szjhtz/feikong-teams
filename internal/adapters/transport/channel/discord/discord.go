package discord

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	channel "fkteams/internal/adapters/transport/channel"
	"fkteams/internal/runtime/env"
	"fkteams/internal/runtime/log"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/websocket"
)

// Register 注册 Discord 通道工厂。
func Register(registry *channel.FactoryRegistry) {
	registry.Register("discord", NewChannel)
}

// Channel Discord 机器人通道
type Channel struct {
	token     string
	allowFrom map[string]bool

	session   *discordgo.Session
	handler   channel.MessageHandler
	running   atomic.Bool
	botID     string
	accepting bool
	cancel    context.CancelFunc
	runCtx    context.Context
	runDone   <-chan struct{}
	mu        sync.Mutex

	typingMu      sync.Mutex
	typingCancels map[string]typingIndicator
	typingSeq     uint64
}

const (
	maxTypingIndicators = 256
	typingRefresh       = 8 * time.Second
	typingMaxDuration   = 2 * time.Minute
	discordHTTPTimeout  = 20 * time.Second
	discordDialTimeout  = 15 * time.Second
)

type typingIndicator struct {
	id        uint64
	cancel    context.CancelFunc
	startedAt time.Time
}

// NewChannel 创建 Discord 通道实例
func NewChannel(cfg channel.ChannelConfig, handler channel.MessageHandler) (channel.Channel, error) {
	c := &Channel{
		token:         cfg.Extra["token"],
		handler:       handler,
		typingCancels: make(map[string]typingIndicator),
	}
	if ids := cfg.Extra["allow_from"]; ids != "" {
		c.allowFrom = make(map[string]bool)
		for _, id := range strings.Split(ids, ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				c.allowFrom[id] = true
			}
		}
	}
	return c, nil
}

func (c *Channel) Name() string    { return "discord" }
func (c *Channel) IsRunning() bool { return c.running.Load() }

// Start 启动 Discord Bot WebSocket 连接
func (c *Channel) Start(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(c.token) == "" {
		return fmt.Errorf("Discord bot token is required")
	}
	if !c.running.CompareAndSwap(false, true) {
		return fmt.Errorf("Discord channel is already running")
	}
	runCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	c.mu.Lock()
	c.cancel = cancel
	c.runCtx = runCtx
	c.runDone = done
	c.accepting = false
	c.mu.Unlock()
	started := false
	var openedSession *discordgo.Session
	defer func() {
		if !started {
			cancel()
			if openedSession != nil {
				_ = openedSession.Close()
			}
			close(done)
			c.mu.Lock()
			if c.runDone == done {
				c.cancel = nil
				c.runCtx = nil
				c.runDone = nil
				c.accepting = false
			}
			c.mu.Unlock()
			c.running.Store(false)
		}
	}()

	session, err := discordgo.New("Bot " + strings.TrimPrefix(c.token, "Bot "))
	if err != nil {
		return err
	}
	dialer := *websocket.DefaultDialer
	dialer.HandshakeTimeout = discordDialTimeout
	session.Dialer = &dialer

	// 配置代理（FEIKONG_PROXY_URL）
	if proxyStr := env.Get(env.ProxyURL); proxyStr != "" {
		proxyURL, err := url.Parse(proxyStr)
		if err != nil {
			return fmt.Errorf("parse Discord proxy URL: %w", err)
		}
		if proxyURL.Scheme == "" || proxyURL.Host == "" {
			return fmt.Errorf("Discord proxy URL must include scheme and host")
		}
		var transport *http.Transport
		if defaultTransport, ok := http.DefaultTransport.(*http.Transport); ok {
			transport = defaultTransport.Clone()
		} else {
			transport = &http.Transport{}
		}
		transport.Proxy = http.ProxyURL(proxyURL)
		session.Client = &http.Client{Transport: transport, Timeout: discordHTTPTimeout}
		dialer.Proxy = http.ProxyURL(proxyURL)
		log.Printf("[discord] using proxy: %s", proxyURL.Redacted())
	}

	session.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsDirectMessages |
		discordgo.IntentsMessageContent

	session.AddHandler(c.messageCreate)

	if err := session.Open(); err != nil {
		return err
	}
	openedSession = session

	c.mu.Lock()
	if err := runCtx.Err(); err != nil {
		c.mu.Unlock()
		return fmt.Errorf("start Discord channel: %w", err)
	}
	c.session = session
	if session.State != nil && session.State.User != nil {
		c.botID = session.State.User.ID
	}
	c.accepting = true
	c.mu.Unlock()
	started = true
	openedSession = nil
	username := "unknown"
	if session.State != nil && session.State.User != nil && session.State.User.Username != "" {
		username = session.State.User.Username
	}
	log.Printf("[discord] Discord bot started (user=%s)", username)

	go func() {
		<-runCtx.Done()
		c.mu.Lock()
		shouldClose := c.session == session
		if shouldClose {
			c.session = nil
			c.accepting = false
		}
		c.mu.Unlock()
		c.cancelAllTyping()
		if shouldClose {
			if err := session.Close(); err != nil {
				log.Printf("[discord] close session: %v", err)
			}
		}
		close(done)
		c.running.Store(false)
	}()

	return nil
}

// Stop 停止 Discord Bot
func (c *Channel) Stop(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	c.mu.Lock()
	session := c.session
	cancel := c.cancel
	done := c.runDone
	c.session = nil
	c.accepting = false
	if cancel != nil {
		cancel()
	}
	c.mu.Unlock()

	c.cancelAllTyping()
	var result error
	if session != nil {
		if err := session.Close(); err != nil {
			result = fmt.Errorf("close Discord session: %w", err)
		}
	}
	if done != nil {
		select {
		case <-done:
		case <-ctx.Done():
			result = errors.Join(result, fmt.Errorf("stop Discord channel: %w", ctx.Err()))
		}
	}
	c.mu.Lock()
	if c.runDone == done {
		c.cancel = nil
		c.runCtx = nil
		c.runDone = nil
		c.botID = ""
	}
	c.mu.Unlock()
	c.running.Store(false)
	log.Printf("[discord] Discord bot stopped")
	return result
}

// Send 向指定频道发送消息
func (c *Channel) Send(ctx context.Context, chatID string, msg channel.Message) error {
	if ctx == nil {
		ctx = context.Background()
	}
	c.mu.Lock()
	session := c.session
	accepting := c.accepting
	c.mu.Unlock()
	if session == nil || !accepting {
		return fmt.Errorf("Discord channel is not running")
	}

	c.stopTyping(chatID)
	channelID := extractChannelID(chatID)
	parts := make([]string, 0, len(msg.Attachments)+1)
	if strings.TrimSpace(msg.Content) != "" {
		parts = append(parts, msg.Content)
	}
	for _, attachment := range msg.Attachments {
		if strings.TrimSpace(attachment.URL) != "" {
			parts = append(parts, attachment.URL)
		}
	}
	if len(parts) == 0 {
		return fmt.Errorf("Discord message has no deliverable content")
	}
	for _, chunk := range splitDiscordMessage(strings.Join(parts, "\n"), 2_000) {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("send Discord message: %w", err)
		}
		_, err := session.ChannelMessageSend(channelID, chunk)
		if err != nil {
			return err
		}
	}

	return nil
}

func splitDiscordMessage(content string, limit int) []string {
	if content == "" || limit <= 0 {
		return nil
	}
	runes := []rune(content)
	chunks := make([]string, 0, (len(runes)+limit-1)/limit)
	for len(runes) > 0 {
		end := min(limit, len(runes))
		chunks = append(chunks, string(runes[:end]))
		runes = runes[end:]
	}
	return chunks
}

// messageCreate 处理收到的消息
func (c *Channel) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m == nil || m.Author == nil {
		return
	}
	botID, runCtx, accepting := c.messageContext(s)
	if !accepting {
		return
	}
	if botID != "" && m.Author.ID == botID {
		return
	}
	if m.Author.Bot {
		return
	}

	if len(c.allowFrom) > 0 && !c.allowFrom[m.Author.ID] {
		return
	}

	isDM := isDMChannel(s, m)
	isGroup := !isDM

	if isGroup {
		if botID == "" {
			return
		}
		mentioned := false
		for _, u := range m.Mentions {
			if u.ID == c.botID {
				mentioned = true
				break
			}
		}
		if !mentioned {
			return
		}
	}

	var chatID string
	if isDM {
		chatID = "dm:" + m.ChannelID
	} else {
		chatID = "guild:" + m.ChannelID
	}

	content := cleanMentions(m.Content, botID)
	content = strings.TrimSpace(content)

	attachments := extractAttachments(m.Attachments)

	if content == "" && len(attachments) == 0 {
		return
	}

	inMsg := channel.Message{Content: content, Attachments: attachments}
	if len(attachments) > 0 {
		inMsg.Type = attachments[0].Type
	}

	ctx := channel.WithChannelName(runCtx, "discord")
	c.startTyping(ctx, s, chatID, m.ChannelID)
	c.handler(ctx, chatID, m.Author.ID, inMsg, isGroup)
}

func (c *Channel) messageContext(session *discordgo.Session) (string, context.Context, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.session != session || !c.accepting || c.runCtx == nil {
		return "", nil, false
	}
	botID := c.botID
	if session.State != nil && session.State.User != nil && session.State.User.ID != "" {
		botID = session.State.User.ID
		c.botID = botID
	}
	return botID, c.runCtx, true
}

func (c *Channel) startTyping(parent context.Context, session *discordgo.Session, chatID, channelID string) {
	if parent == nil || session == nil || chatID == "" || channelID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(parent, typingMaxDuration)
	id := c.registerTyping(chatID, cancel)
	go func() {
		defer cancel()
		defer c.removeTyping(chatID, id)
		if err := session.ChannelTyping(channelID); err != nil {
			return
		}
		ticker := time.NewTicker(typingRefresh)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := session.ChannelTyping(channelID); err != nil {
					return
				}
			}
		}
	}()
}

func (c *Channel) registerTyping(chatID string, cancel context.CancelFunc) uint64 {
	c.typingMu.Lock()
	defer c.typingMu.Unlock()
	if previous, ok := c.typingCancels[chatID]; ok {
		previous.cancel()
		delete(c.typingCancels, chatID)
	}
	for len(c.typingCancels) >= maxTypingIndicators {
		var oldestChat string
		var oldestTime time.Time
		for candidate, indicator := range c.typingCancels {
			if oldestChat == "" || indicator.startedAt.Before(oldestTime) {
				oldestChat = candidate
				oldestTime = indicator.startedAt
			}
		}
		oldest := c.typingCancels[oldestChat]
		oldest.cancel()
		delete(c.typingCancels, oldestChat)
	}
	c.typingSeq++
	id := c.typingSeq
	c.typingCancels[chatID] = typingIndicator{id: id, cancel: cancel, startedAt: time.Now()}
	return id
}

func (c *Channel) stopTyping(chatID string) {
	c.typingMu.Lock()
	if indicator, ok := c.typingCancels[chatID]; ok {
		indicator.cancel()
		delete(c.typingCancels, chatID)
	}
	c.typingMu.Unlock()
}

func (c *Channel) removeTyping(chatID string, id uint64) {
	c.typingMu.Lock()
	if indicator, ok := c.typingCancels[chatID]; ok && indicator.id == id {
		delete(c.typingCancels, chatID)
	}
	c.typingMu.Unlock()
}

func (c *Channel) cancelAllTyping() {
	c.typingMu.Lock()
	for chatID, indicator := range c.typingCancels {
		indicator.cancel()
		delete(c.typingCancels, chatID)
	}
	c.typingMu.Unlock()
}

// isDMChannel 判断消息是否来自私聊
func isDMChannel(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	ch, err := s.State.Channel(m.ChannelID)
	if err != nil {
		ch, err = s.Channel(m.ChannelID)
		if err != nil {
			return false
		}
	}
	return ch.Type == discordgo.ChannelTypeDM
}

// cleanMentions 移除文本中对 bot 的 @mention
func cleanMentions(content, botID string) string {
	content = strings.ReplaceAll(content, "<@"+botID+">", "")
	content = strings.ReplaceAll(content, "<@!"+botID+">", "")
	return content
}

// extractAttachments 从 Discord 消息附件中提取 Attachment 列表
func extractAttachments(atts []*discordgo.MessageAttachment) []channel.Attachment {
	if len(atts) == 0 {
		return nil
	}
	var result []channel.Attachment
	for _, a := range atts {
		if a == nil {
			continue
		}
		t := guessAttachmentType(a.Filename, a.ContentType)
		result = append(result, channel.Attachment{
			Type:     t,
			URL:      a.URL,
			FileName: a.Filename,
		})
	}
	return result
}

// guessAttachmentType 根据文件名和 Content-Type 推断附件类型
func guessAttachmentType(fileName, contentType string) channel.MessageType {
	ct := strings.ToLower(contentType)
	switch {
	case strings.HasPrefix(ct, "image/"):
		return channel.MsgImage
	case strings.HasPrefix(ct, "video/"):
		return channel.MsgVideo
	case strings.HasPrefix(ct, "audio/"):
		return channel.MsgAudio
	}
	name := strings.ToLower(fileName)
	switch {
	case strings.HasSuffix(name, ".jpg"), strings.HasSuffix(name, ".jpeg"),
		strings.HasSuffix(name, ".png"), strings.HasSuffix(name, ".gif"),
		strings.HasSuffix(name, ".webp"):
		return channel.MsgImage
	case strings.HasSuffix(name, ".mp4"), strings.HasSuffix(name, ".avi"),
		strings.HasSuffix(name, ".mov"), strings.HasSuffix(name, ".mkv"):
		return channel.MsgVideo
	case strings.HasSuffix(name, ".mp3"), strings.HasSuffix(name, ".wav"),
		strings.HasSuffix(name, ".ogg"), strings.HasSuffix(name, ".flac"):
		return channel.MsgAudio
	default:
		return channel.MsgFile
	}
}

// extractChannelID 从 chatID 中提取 Discord 频道 ID
func extractChannelID(chatID string) string {
	if strings.HasPrefix(chatID, "dm:") {
		return strings.TrimPrefix(chatID, "dm:")
	}
	return strings.TrimPrefix(chatID, "guild:")
}
