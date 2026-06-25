package hooks

import (
	"context"
	"fkteams/internal/domain/event"
	"fkteams/internal/domain/message"
	runtimeport "fkteams/internal/ports/runtime"
)

// HookPoint 表示运行期可扩展的稳定边界。
type HookPoint string

const (
	HookBeforeRun          HookPoint = "before_run"
	HookAfterRun           HookPoint = "after_run"
	HookOnEvent            HookPoint = "on_event"
	HookBeforeToolCall     HookPoint = "before_tool_call"
	HookAfterToolCall      HookPoint = "after_tool_call"
	HookBeforeModelRequest HookPoint = "before_model_request"
	HookAfterModelResponse HookPoint = "after_model_response"
)

// Action 表示 hook 对当前扩展点的处理建议。
type Action string

const (
	ActionContinue Action = "continue"
	ActionSkip     Action = "skip"
	ActionReject   Action = "reject"
)

type ErrorPolicy string

const (
	ErrorIgnore ErrorPolicy = "ignore"
	ErrorWarn   ErrorPolicy = "warn"
	ErrorFail   ErrorPolicy = "fail"
)

type Invocation struct {
	HookPoint HookPoint
	SessionID string
	RunID     string
	TurnID    string
	Payload   Payload
}

type Result struct {
	Payload Payload
	Action  Action
	Message string
}

// Payload 是所有 hook 载荷必须实现的显式契约。
type Payload interface {
	HookPoint() HookPoint
}

type Handler interface {
	Name() string
	Points() []HookPoint
	Handle(ctx context.Context, inv Invocation) (Result, error)
}

type HandlerFunc func(ctx context.Context, inv Invocation) (Result, error)

type funcHandler struct {
	name   string
	points []HookPoint
	fn     HandlerFunc
}

func NewHandler(name string, points []HookPoint, fn HandlerFunc) Handler {
	return &funcHandler{name: name, points: append([]HookPoint(nil), points...), fn: fn}
}

func (h *funcHandler) Name() string {
	return h.name
}

func (h *funcHandler) Points() []HookPoint {
	return append([]HookPoint(nil), h.points...)
}

func (h *funcHandler) Handle(ctx context.Context, inv Invocation) (Result, error) {
	if h.fn == nil {
		return Result{}, nil
	}
	return h.fn(ctx, inv)
}

type BeforeRunPayload struct {
	Input message.TurnInput
}

func (BeforeRunPayload) HookPoint() HookPoint { return HookBeforeRun }

type AfterRunPayload struct {
	Input  message.TurnInput
	Result *runtimeport.RunResult
	Error  error
}

func (AfterRunPayload) HookPoint() HookPoint { return HookAfterRun }

type EventPayload struct {
	Event event.Event
}

func (EventPayload) HookPoint() HookPoint { return HookOnEvent }

type BeforeToolCallPayload struct {
	ToolName string
	Args     string
	Meta     map[string]any
}

func (BeforeToolCallPayload) HookPoint() HookPoint { return HookBeforeToolCall }

type AfterToolCallPayload struct {
	ToolName string
	Args     string
	Result   string
	Error    error
	Meta     map[string]any
}

func (AfterToolCallPayload) HookPoint() HookPoint { return HookAfterToolCall }

type BeforeModelRequestPayload struct {
	Messages []message.Message
	Meta     map[string]any
}

func (BeforeModelRequestPayload) HookPoint() HookPoint { return HookBeforeModelRequest }

type AfterModelResponsePayload struct {
	Message message.Message
	Usage   *event.Event
	Error   error
	Meta    map[string]any
}

func (AfterModelResponsePayload) HookPoint() HookPoint { return HookAfterModelResponse }
