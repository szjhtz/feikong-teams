// Package taskstream 提供统一的任务事件流管理，支持 Push（WebSocket）和 Pull（SSE）两种消费模式。
//
// 设计原则：
//   - 事件日志：所有事件带递增 ID 持久保存，支持任意 offset 重连
//   - Push 订阅：多个 Subscriber（如多端 WS 连接），事件实时推送
//   - Pull 监听：多个 Watcher（如 SSE 连接），通过通知+轮询获取事件
//   - 连接分离：断连只解绑订阅者，不影响后台任务
//   - TTL 清理：任务完成后保留一段时间供重连客户端拉取残余事件
package taskstream

import (
	"fmt"
	"sync"
	"time"
)

// Event 是传输无关的任务事件
type Event = map[string]any

// IndexedEvent 带递增 ID 的事件，支持 offset 断点续传
type IndexedEvent struct {
	ID   uint64 `json:"id"`
	Data Event  `json:"data"`
}

// SubscriptionID 标识一个 Push 订阅者。
type SubscriptionID uint64

// InterruptKind 描述当前等待的人工输入类型。
type InterruptKind string

const (
	InterruptNone     InterruptKind = ""
	InterruptApproval InterruptKind = "approval"
	InterruptAsk      InterruptKind = "ask"
)

// Subscriber 接收推送事件（Push 模式，如 WebSocket 连接）
type Subscriber interface {
	WriteEvent(event Event) error
}

// FuncSubscriber 将函数适配为 Subscriber 接口
type FuncSubscriber func(any) error

func (f FuncSubscriber) WriteEvent(event Event) error { return f(event) }

// StreamConfig 创建 Stream 时的配置
type StreamConfig struct {
	SessionID  string
	Cancel     func()        // 取消任务的函数
	CleanupTTL time.Duration // 任务完成后保留数据的时间（0=立即清理）

	// 元数据（可选）
	Mode      string // 协作模式
	AgentName string // 智能体名称
}

// Stream 代表单个任务的事件流，是事件投递的核心抽象。
// 同时支持 Push（Subscriber）和 Pull（Watcher + EventsSince）两种消费方式。
type Stream struct {
	mu     sync.Mutex
	config StreamConfig

	// 事件日志（带递增 ID，支持断点续传）
	events []IndexedEvent
	nextID uint64

	// Push 订阅者（多个，如多端 WS 连接）
	subs    map[SubscriptionID]Subscriber
	subNext SubscriptionID

	// Pull 监听者（多个，如 SSE 连接）
	watchers    map[uint64]chan struct{}
	watcherNext uint64

	// 生命周期
	done        bool
	status      string // "processing", "completed", "error", "cancelled"
	createdAt   time.Time
	doneAt      time.Time
	interruptCh chan any // 审批/ask 通道
	pendingKind InterruptKind
	submitted   bool

	// 所属 Manager 引用（用于 grace timer 自动移除）
	manager *Manager
}

// Publish 发布事件到流。有订阅者时推送，同时写入日志，通知所有监听者。
func (s *Stream) Publish(event Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 写入事件日志
	id := s.nextID
	s.nextID++
	event["stream_event_id"] = id
	s.events = append(s.events, IndexedEvent{ID: id, Data: event})

	// 推送给所有 Push 订阅者
	for subID, sub := range s.subs {
		if err := sub.WriteEvent(event); err != nil {
			delete(s.subs, subID)
		}
	}

	// 通知所有 Pull 监听者
	for _, ch := range s.watchers {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// Subscribe 绑定 Push 订阅者并回放错过的事件。
// 返回 (false, 0) 表示流已结束/过期，调用方需自行通知客户端。
// 返回 (true, id) 表示绑定成功，调用方应保存 id 用于后续 Unsubscribe。
func (s *Stream) Subscribe(sub Subscriber, offset uint64) (bool, SubscriptionID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.done {
		// 流已结束，不回放事件（避免重连时重复渲染）
		// 已完成的任务数据应通过历史 API 获取
		return false, 0
	}

	for _, e := range s.eventsSinceLocked(offset) {
		if err := sub.WriteEvent(e.Data); err != nil {
			return false, 0
		}
	}
	if s.subs == nil {
		s.subs = make(map[SubscriptionID]Subscriber)
	}
	s.subNext++
	subID := s.subNext
	s.subs[subID] = sub
	return true, subID
}

// Unsubscribe 解绑 Push 订阅者。断连只影响当前订阅，不取消后台任务。
func (s *Stream) Unsubscribe(id SubscriptionID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.subs, id)
}

// SubscriptionCount 返回当前 Push 订阅者数量。
func (s *Stream) SubscriptionCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.subs)
}

// Watch 返回事件通知通道和取消函数（Pull 模式，用于 SSE）。
// 当有新事件发布时，通知通道会收到信号。
func (s *Stream) Watch() (<-chan struct{}, func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan struct{}, 1)
	id := s.watcherNext
	s.watcherNext++
	if s.watchers == nil {
		s.watchers = make(map[uint64]chan struct{})
	}
	s.watchers[id] = ch

	unwatch := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.watchers, id)
	}
	return ch, unwatch
}

// EventsSince 返回从 offset 开始的所有事件（Pull 模式，用于 SSE 和 HTTP 轮询）。
func (s *Stream) EventsSince(offset uint64) []IndexedEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.eventsSinceLocked(offset)
}

func (s *Stream) eventsSinceLocked(offset uint64) []IndexedEvent {
	// 二分查找起始位置
	start := 0
	for start < len(s.events) && s.events[start].ID < offset {
		start++
	}
	if start >= len(s.events) {
		return nil
	}
	result := make([]IndexedEvent, len(s.events)-start)
	copy(result, s.events[start:])
	return result
}

// Done 标记流已完成。通知所有监听者。
func (s *Stream) Done() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.done = true
	s.doneAt = time.Now()

	// 通知所有 Pull 监听者（使其退出等待循环）
	for _, ch := range s.watchers {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// Cancel 取消底层任务
func (s *Stream) Cancel() {
	s.config.Cancel()
}

// BeginInterrupt 标记当前流正在等待指定类型的人工输入。
func (s *Stream) BeginInterrupt(kind InterruptKind) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.drainInterruptLocked()
	s.pendingKind = kind
	s.submitted = false
}

// CompleteInterrupt 清除当前人工输入等待状态。
func (s *Stream) CompleteInterrupt(kind InterruptKind) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pendingKind == kind {
		s.pendingKind = InterruptNone
		s.submitted = false
	}
}

// SubmitInterrupt 提交人工输入。仅当前确实存在同类型 pending 时才接受。
func (s *Stream) SubmitInterrupt(kind InterruptKind, value any) error {
	s.mu.Lock()
	if s.done || s.status != "processing" {
		s.mu.Unlock()
		return fmt.Errorf("task is not processing")
	}
	if s.pendingKind != kind {
		s.mu.Unlock()
		return fmt.Errorf("no pending %s request", kind)
	}
	if s.submitted {
		s.mu.Unlock()
		return fmt.Errorf("%s request already submitted", kind)
	}
	s.submitted = true
	ch := s.interruptCh
	s.mu.Unlock()

	select {
	case ch <- value:
		return nil
	default:
		s.mu.Lock()
		if s.pendingKind == kind {
			s.submitted = false
		}
		s.mu.Unlock()
		return fmt.Errorf("%s request is not ready", kind)
	}
}

func (s *Stream) drainInterruptLocked() {
	for {
		select {
		case <-s.interruptCh:
		default:
			return
		}
	}
}

// IsDone 检查流是否已完成
func (s *Stream) IsDone() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.done
}

// --- 状态与元数据 ---

// SetStatus 设置任务状态
func (s *Stream) SetStatus(status string) {
	s.mu.Lock()
	s.status = status
	s.mu.Unlock()
}

// Status 返回任务状态
func (s *Stream) Status() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status
}

// Mode 返回协作模式
func (s *Stream) Mode() string {
	return s.config.Mode
}

// AgentName 返回智能体名称
func (s *Stream) AgentName() string {
	return s.config.AgentName
}

// CreatedAt 返回流创建时间
func (s *Stream) CreatedAt() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.createdAt
}

// DoneAt 返回流完成的时间
func (s *Stream) DoneAt() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.doneAt
}

// SessionID 返回关联的会话 ID
func (s *Stream) SessionID() string {
	return s.config.SessionID
}

// InterruptCh 返回审批/ask 中断通道
func (s *Stream) InterruptCh() chan any {
	return s.interruptCh
}

// EventCount 返回事件日志中的事件数量
func (s *Stream) EventCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.events)
}
