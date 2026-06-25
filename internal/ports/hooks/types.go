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
	Payload   any
}

type Result struct {
	Payload any
	Action  Action
	Message string
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

type AfterRunPayload struct {
	Input  message.TurnInput
	Result *runtimeport.RunResult
	Error  error
}

type EventPayload struct {
	Event event.Event
}

type BeforeToolCallPayload struct {
	ToolName string
	Args     string
	Meta     map[string]any
}

type AfterToolCallPayload struct {
	ToolName string
	Args     string
	Result   string
	Error    error
	Meta     map[string]any
}

type BeforeModelRequestPayload struct {
	Messages []message.Message
	Meta     map[string]any
}

type AfterModelResponsePayload struct {
	Message message.Message
	Usage   *event.Event
	Error   error
	Meta    map[string]any
}
