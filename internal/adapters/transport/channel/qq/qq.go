package qq

import (
	"container/list"
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	channel "fkteams/internal/adapters/transport/channel"
	"fkteams/internal/runtime/log"

	"github.com/tencent-connect/botgo"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/event"
	"github.com/tencent-connect/botgo/openapi"
	"github.com/tencent-connect/botgo/token"
)

// Register 注册 QQ 通道工厂。
func Register(registry *channel.FactoryRegistry) {
	registry.Register("qq", NewChannel)
}

// chatState 保存每个会话的最近消息 ID 和序号（用于被动回复）
type chatState struct {
	mu       sync.Mutex
	msgID    string
	msgSeq   atomic.Uint32
	lastSeen time.Time
	order    *list.Element
}

// Channel QQ 机器人通道
type Channel struct {
	appID     string
	appSecret string
	sandbox   bool

	api     openapi.OpenAPI
	handler channel.MessageHandler
	running atomic.Bool

	lifecycleMu sync.Mutex
	cancel      context.CancelFunc
	sessionDone <-chan error

	// 消息去重（TTL 5 分钟）
	seen      map[string]time.Time
	seenOrder *list.List
	seenMu    sync.Mutex

	// 每个会话的状态
	states     map[string]*chatState
	stateOrder *list.List
	statesMu   sync.Mutex
}

const (
	seenMessageTTL  = 5 * time.Minute
	chatStateTTL    = 30 * time.Minute
	maxSeenMessages = 20_000
	maxChatStates   = 4_096
	cleanupInterval = 2 * time.Minute
)

type seenMessage struct {
	id     string
	seenAt time.Time
}

// NewChannel 创建 QQ 通道实例
func NewChannel(cfg channel.ChannelConfig, handler channel.MessageHandler) (channel.Channel, error) {
	return &Channel{
		appID:      cfg.Extra["app_id"],
		appSecret:  cfg.Extra["app_secret"],
		sandbox:    cfg.Extra["sandbox"] == "true",
		handler:    handler,
		seen:       make(map[string]time.Time),
		seenOrder:  list.New(),
		states:     make(map[string]*chatState),
		stateOrder: list.New(),
	}, nil
}

func (c *Channel) Name() string    { return "qq" }
func (c *Channel) IsRunning() bool { return c.running.Load() }

// Start 启动 QQ 机器人 WebSocket 连接
func (c *Channel) Start(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if !c.running.CompareAndSwap(false, true) {
		return fmt.Errorf("QQ channel is already running")
	}
	runCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	c.lifecycleMu.Lock()
	c.cancel = cancel
	c.sessionDone = done
	c.lifecycleMu.Unlock()
	started := false
	defer func() {
		if !started {
			cancel()
			close(done)
			c.running.Store(false)
		}
	}()

	tokenSource := token.NewQQBotTokenSource(
		&token.QQBotCredentials{
			AppID:     c.appID,
			AppSecret: c.appSecret,
		},
	)

	if c.sandbox {
		c.api = botgo.NewSandboxOpenAPI(c.appID, tokenSource).WithTimeout(10 * time.Second)
	} else {
		c.api = botgo.NewOpenAPI(c.appID, tokenSource).WithTimeout(10 * time.Second)
	}

	intent := event.RegisterHandlers(
		c.c2cMessageHandler(),
		c.groupATMessageHandler(),
	)

	wsInfo, err := c.api.WS(runCtx, nil, "")
	if err != nil {
		return err
	}

	started = true

	go func() {
		cleanupDone := make(chan struct{})
		go func() {
			defer close(cleanupDone)
			c.cleanupState(runCtx)
		}()

		err := newSessionRunner().Run(runCtx, wsInfo, tokenSource, &intent)
		if err != nil && runCtx.Err() == nil {
			log.Printf("[qq] session manager exited: %v", err)
		}
		cancel()
		<-cleanupDone
		c.resetStateCaches()
		done <- err
		close(done)
		c.running.Store(false)
	}()

	log.Printf("[qq] QQ bot started (appID=%s, sandbox=%v)", c.appID, c.sandbox)
	return nil
}

// Stop 停止 QQ 机器人
func (c *Channel) Stop(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	c.lifecycleMu.Lock()
	cancel := c.cancel
	done := c.sessionDone
	c.lifecycleMu.Unlock()
	if cancel == nil {
		c.running.Store(false)
		return nil
	}
	cancel()
	if done != nil {
		select {
		case <-done:
		case <-ctx.Done():
			return fmt.Errorf("stop QQ channel: %w", ctx.Err())
		}
	}
	c.lifecycleMu.Lock()
	if c.sessionDone == done {
		c.cancel = nil
		c.sessionDone = nil
	}
	c.lifecycleMu.Unlock()
	c.running.Store(false)
	log.Printf("[qq] QQ bot stopped")
	return nil
}

// Send 向指定会话发送消息（支持文本和多媒体）
func (c *Channel) Send(ctx context.Context, chatID string, msg channel.Message) error {
	isGroup := strings.HasPrefix(chatID, "group:")
	targetID := strings.TrimPrefix(chatID, "group:")
	targetID = strings.TrimPrefix(targetID, "c2c:")

	state := c.getState(chatID)
	seq := state.msgSeq.Add(1)

	// 富媒体消息（图片、语音、视频、文件）
	if msg.Type != channel.MsgText && len(msg.Attachments) > 0 {
		for _, att := range msg.Attachments {
			richMsg := &dto.RichMediaMessage{
				FileType:   qqFileType(att.Type),
				URL:        att.URL,
				SrvSendMsg: true,
				MsgSeq:     int64(seq),
			}
			if isGroup {
				_, err := c.api.PostGroupMessage(ctx, targetID, richMsg)
				if err != nil {
					return err
				}
			} else {
				_, err := c.api.PostC2CMessage(ctx, targetID, richMsg)
				if err != nil {
					return err
				}
			}
			seq = state.msgSeq.Add(1)
		}
		// 如果富媒体消息同时附带文本，继续发送文本
		if msg.Content == "" {
			return nil
		}
	}

	// 文本消息
	state.mu.Lock()
	msgID := state.msgID
	state.mu.Unlock()
	textMsg := &dto.MessageToCreate{
		Content: msg.Content,
		MsgType: dto.TextMsg,
		MsgID:   msgID,
		MsgSeq:  uint32(seq),
	}

	if isGroup {
		_, err := c.api.PostGroupMessage(ctx, targetID, textMsg)
		return err
	}
	_, err := c.api.PostC2CMessage(ctx, targetID, textMsg)
	return err
}

// qqFileType 将通用消息类型映射为 QQ 富媒体文件类型
func qqFileType(t channel.MessageType) uint64 {
	switch t {
	case channel.MsgImage:
		return 1
	case channel.MsgVideo:
		return 2
	case channel.MsgAudio:
		return 3
	default:
		return 1
	}
}

// c2cMessageHandler 处理 C2C（私聊）消息
func (c *Channel) c2cMessageHandler() event.C2CMessageEventHandler {
	return func(ev *dto.WSPayload, data *dto.WSC2CMessageData) error {
		msg := (*dto.Message)(data)
		if c.isDuplicate(msg.ID) {
			return nil
		}

		chatID := "c2c:" + msg.Author.ID
		c.updateState(chatID, msg.ID)

		content := strings.TrimSpace(msg.Content)
		attachments := extractAttachments(msg)
		if content == "" && len(attachments) == 0 {
			return nil
		}

		inMsg := channel.Message{Content: content, Attachments: attachments}
		if len(attachments) > 0 {
			inMsg.Type = attachments[0].Type
		}

		ctx := channel.WithChannelName(context.Background(), "qq")
		c.handler(ctx, chatID, msg.Author.ID, inMsg, false)
		return nil
	}
}

// groupATMessageHandler 处理群 @机器人 消息
func (c *Channel) groupATMessageHandler() event.GroupATMessageEventHandler {
	return func(ev *dto.WSPayload, data *dto.WSGroupATMessageData) error {
		msg := (*dto.Message)(data)
		if c.isDuplicate(msg.ID) {
			return nil
		}

		chatID := "group:" + msg.GroupID
		c.updateState(chatID, msg.ID)

		content := strings.TrimSpace(msg.Content)
		attachments := extractAttachments(msg)
		if content == "" && len(attachments) == 0 {
			return nil
		}

		inMsg := channel.Message{Content: content, Attachments: attachments}
		if len(attachments) > 0 {
			inMsg.Type = attachments[0].Type
		}

		ctx := channel.WithChannelName(context.Background(), "qq")
		c.handler(ctx, chatID, msg.Author.ID, inMsg, true)
		return nil
	}
}

// extractAttachments 从 QQ 消息中提取附件
func extractAttachments(msg *dto.Message) []channel.Attachment {
	if len(msg.Attachments) == 0 {
		return nil
	}
	var atts []channel.Attachment
	for _, a := range msg.Attachments {
		t := guessAttachmentType(a.URL, a.FileName)
		atts = append(atts, channel.Attachment{
			Type:     t,
			URL:      a.URL,
			FileName: a.FileName,
		})
	}
	return atts
}

// guessAttachmentType 根据文件名或 URL 推断附件类型
func guessAttachmentType(url, fileName string) channel.MessageType {
	name := strings.ToLower(fileName)
	if name == "" {
		name = strings.ToLower(url)
	}
	switch {
	case strings.HasSuffix(name, ".jpg"), strings.HasSuffix(name, ".jpeg"),
		strings.HasSuffix(name, ".png"), strings.HasSuffix(name, ".gif"),
		strings.HasSuffix(name, ".webp"), strings.HasSuffix(name, ".bmp"):
		return channel.MsgImage
	case strings.HasSuffix(name, ".mp4"), strings.HasSuffix(name, ".avi"),
		strings.HasSuffix(name, ".mov"), strings.HasSuffix(name, ".mkv"):
		return channel.MsgVideo
	case strings.HasSuffix(name, ".mp3"), strings.HasSuffix(name, ".wav"),
		strings.HasSuffix(name, ".silk"), strings.HasSuffix(name, ".amr"),
		strings.HasSuffix(name, ".ogg"):
		return channel.MsgAudio
	default:
		return channel.MsgFile
	}
}

// isDuplicate 检查消息是否重复（5 分钟内的重复 msgID）
func (c *Channel) isDuplicate(msgID string) bool {
	if msgID == "" {
		return false
	}
	now := time.Now()
	c.seenMu.Lock()
	defer c.seenMu.Unlock()
	c.pruneSeenLocked(now)
	if seenAt, ok := c.seen[msgID]; ok && now.Sub(seenAt) < seenMessageTTL {
		return true
	}
	for len(c.seen) >= maxSeenMessages {
		c.removeOldestSeenLocked()
	}
	c.seen[msgID] = now
	c.seenOrder.PushBack(seenMessage{id: msgID, seenAt: now})
	return false
}

// cleanupState 定期清理过期的去重记录和空闲会话状态。
func (c *Channel) cleanupState(ctx context.Context) {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			c.seenMu.Lock()
			c.pruneSeenLocked(now)
			c.seenMu.Unlock()

			c.statesMu.Lock()
			c.pruneStatesLocked(now)
			c.statesMu.Unlock()
		}
	}
}

// getState 获取会话状态，不存在则创建
func (c *Channel) getState(chatID string) *chatState {
	c.statesMu.Lock()
	defer c.statesMu.Unlock()
	c.ensureStateCacheLocked()
	now := time.Now()
	if s, ok := c.states[chatID]; ok {
		s.mu.Lock()
		s.lastSeen = now
		s.mu.Unlock()
		c.stateOrder.MoveToFront(s.order)
		return s
	}
	for len(c.states) >= maxChatStates {
		c.removeOldestStateLocked()
	}
	s := &chatState{lastSeen: now}
	s.order = c.stateOrder.PushFront(chatID)
	c.states[chatID] = s
	return s
}

// updateState 更新会话的最近消息 ID
func (c *Channel) updateState(chatID, msgID string) {
	s := c.getState(chatID)
	s.mu.Lock()
	s.msgID = msgID
	s.lastSeen = time.Now()
	s.mu.Unlock()
	s.msgSeq.Store(0)
}

func (c *Channel) ensureStateCacheLocked() {
	if c.stateOrder == nil {
		c.stateOrder = list.New()
	}
}

func (c *Channel) pruneSeenLocked(now time.Time) {
	if c.seenOrder == nil {
		c.seenOrder = list.New()
	}
	for element := c.seenOrder.Front(); element != nil; element = c.seenOrder.Front() {
		entry := element.Value.(seenMessage)
		if now.Sub(entry.seenAt) < seenMessageTTL {
			return
		}
		c.removeOldestSeenLocked()
	}
}

func (c *Channel) removeOldestSeenLocked() {
	if c.seenOrder == nil {
		return
	}
	element := c.seenOrder.Front()
	if element == nil {
		return
	}
	entry := element.Value.(seenMessage)
	if c.seen[entry.id].Equal(entry.seenAt) {
		delete(c.seen, entry.id)
	}
	c.seenOrder.Remove(element)
}

func (c *Channel) pruneStatesLocked(now time.Time) {
	c.ensureStateCacheLocked()
	for element := c.stateOrder.Back(); element != nil; element = c.stateOrder.Back() {
		chatID := element.Value.(string)
		state := c.states[chatID]
		state.mu.Lock()
		idle := now.Sub(state.lastSeen) >= chatStateTTL
		state.mu.Unlock()
		if !idle {
			return
		}
		c.removeOldestStateLocked()
	}
}

func (c *Channel) removeOldestStateLocked() {
	c.ensureStateCacheLocked()
	element := c.stateOrder.Back()
	if element == nil {
		return
	}
	delete(c.states, element.Value.(string))
	c.stateOrder.Remove(element)
}

func (c *Channel) resetStateCaches() {
	c.seenMu.Lock()
	c.seen = make(map[string]time.Time)
	c.seenOrder = list.New()
	c.seenMu.Unlock()
	c.statesMu.Lock()
	c.states = make(map[string]*chatState)
	c.stateOrder = list.New()
	c.statesMu.Unlock()
}
