package agentcore

import (
	domainevent "fkteams/internal/domain/event"
	domainmessage "fkteams/internal/domain/message"
	runtimeport "fkteams/internal/ports/runtime"
	checkpointstore "fkteams/internal/runtime/checkpoint"
)

type MessageRole = domainmessage.Role

const (
	RoleSystem    = domainmessage.RoleSystem
	RoleUser      = domainmessage.RoleUser
	RoleAssistant = domainmessage.RoleAssistant
	RoleTool      = domainmessage.RoleTool
)

type ContentPartType = domainmessage.ContentPartType

const (
	ContentPartText     = domainmessage.ContentPartText
	ContentPartImageURL = domainmessage.ContentPartImageURL
	ContentPartAudioURL = domainmessage.ContentPartAudioURL
	ContentPartVideoURL = domainmessage.ContentPartVideoURL
	ContentPartFileURL  = domainmessage.ContentPartFileURL
)

type ContentPart = domainmessage.ContentPart
type FunctionCall = domainmessage.FunctionCall
type ToolCall = domainmessage.ToolCall
type Message = domainmessage.Message
type TurnInput = domainmessage.TurnInput

type EventType = domainevent.Type

const (
	EventAgentStart   = domainevent.TypeAgentStart
	EventAgentEnd     = domainevent.TypeAgentEnd
	EventTurnStart    = domainevent.TypeTurnStart
	EventTurnEnd      = domainevent.TypeTurnEnd
	EventMessageStart = domainevent.TypeMessageStart
	EventMessageDelta = domainevent.TypeMessageDelta
	EventMessageEnd   = domainevent.TypeMessageEnd
	EventToolStart    = domainevent.TypeToolStart
	EventToolUpdate   = domainevent.TypeToolUpdate
	EventToolEnd      = domainevent.TypeToolEnd
	EventAction       = domainevent.TypeAction
	EventUsage        = domainevent.TypeUsage
	EventError        = domainevent.TypeError
	EventMemberUpdate = domainevent.TypeMemberUpdate
)

type DeltaKind = domainevent.DeltaKind

const (
	DeltaOutput     = domainevent.DeltaOutput
	DeltaReasoning  = domainevent.DeltaReasoning
	DeltaToolArgs   = domainevent.DeltaToolArgs
	DeltaToolResult = domainevent.DeltaToolResult
)

type ActionType = domainevent.ActionType

const (
	ActionTransfer             = domainevent.ActionTransfer
	ActionInterrupted          = domainevent.ActionInterrupted
	ActionExit                 = domainevent.ActionExit
	ActionAskQuestions         = domainevent.ActionAskQuestions
	ActionAskResponse          = domainevent.ActionAskResponse
	ActionApprovalRequired     = domainevent.ActionApprovalRequired
	ActionApprovalDecision     = domainevent.ActionApprovalDecision
	ActionContextCompressStart = domainevent.ActionContextCompressStart
	ActionContextCompress      = domainevent.ActionContextCompress
)

type NotifyType = domainevent.NotifyType

const (
	NotifyProcessingStart  = domainevent.NotifyProcessingStart
	NotifyProcessingEnd    = domainevent.NotifyProcessingEnd
	NotifyUserMessage      = domainevent.NotifyUserMessage
	NotifyQueueUpdated     = domainevent.NotifyQueueUpdated
	NotifyCancelled        = domainevent.NotifyCancelled
	NotifyError            = domainevent.NotifyError
	NotifyAskQuestions     = domainevent.NotifyAskQuestions
	NotifyApprovalRequired = domainevent.NotifyApprovalRequired
	NotifyConnected        = domainevent.NotifyConnected
	NotifyPong             = domainevent.NotifyPong
	NotifyInvalidAPIKey    = domainevent.NotifyInvalidAPIKey
)

type Event = domainevent.Event
type Interrupt = runtimeport.Interrupt
type InterruptHandler = runtimeport.InterruptHandler
type EventSink = runtimeport.EventSink
type RunOptions = runtimeport.RunOptions
type RunResult = runtimeport.RunResult
type Runner = runtimeport.Runner
type CheckPointStore = checkpointstore.Store

var NoopEventSink = runtimeport.NoopEventSink
