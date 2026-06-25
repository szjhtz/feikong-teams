package hooks

import hookport "fkteams/internal/ports/hooks"

type HookPoint = hookport.HookPoint

const (
	HookBeforeRun          = hookport.HookBeforeRun
	HookAfterRun           = hookport.HookAfterRun
	HookOnEvent            = hookport.HookOnEvent
	HookBeforeToolCall     = hookport.HookBeforeToolCall
	HookAfterToolCall      = hookport.HookAfterToolCall
	HookBeforeModelRequest = hookport.HookBeforeModelRequest
	HookAfterModelResponse = hookport.HookAfterModelResponse
)

type Action = hookport.Action

const (
	ActionContinue = hookport.ActionContinue
	ActionSkip     = hookport.ActionSkip
	ActionReject   = hookport.ActionReject
)

type ErrorPolicy = hookport.ErrorPolicy

const (
	ErrorIgnore = hookport.ErrorIgnore
	ErrorWarn   = hookport.ErrorWarn
	ErrorFail   = hookport.ErrorFail
)

type Invocation = hookport.Invocation
type Result = hookport.Result
type Handler = hookport.Handler
type HandlerFunc = hookport.HandlerFunc
type BeforeRunPayload = hookport.BeforeRunPayload
type AfterRunPayload = hookport.AfterRunPayload
type EventPayload = hookport.EventPayload
type BeforeToolCallPayload = hookport.BeforeToolCallPayload
type AfterToolCallPayload = hookport.AfterToolCallPayload
type BeforeModelRequestPayload = hookport.BeforeModelRequestPayload
type AfterModelResponsePayload = hookport.AfterModelResponsePayload

func NewHandler(name string, points []HookPoint, fn HandlerFunc) Handler {
	return hookport.NewHandler(name, points, fn)
}
