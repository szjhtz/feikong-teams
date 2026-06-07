package events

import "fkteams/agentcore"

type EventType = agentcore.EventType

const (
	EventAgentStart   = agentcore.EventAgentStart
	EventAgentEnd     = agentcore.EventAgentEnd
	EventTurnStart    = agentcore.EventTurnStart
	EventTurnEnd      = agentcore.EventTurnEnd
	EventMessageStart = agentcore.EventMessageStart
	EventMessageDelta = agentcore.EventMessageDelta
	EventMessageEnd   = agentcore.EventMessageEnd
	EventToolStart    = agentcore.EventToolStart
	EventToolUpdate   = agentcore.EventToolUpdate
	EventToolEnd      = agentcore.EventToolEnd
	EventAction       = agentcore.EventAction
	EventUsage        = agentcore.EventUsage
	EventError        = agentcore.EventError
	EventMemberUpdate = agentcore.EventMemberUpdate
)

type DeltaKind = agentcore.DeltaKind

const (
	DeltaOutput     = agentcore.DeltaOutput
	DeltaReasoning  = agentcore.DeltaReasoning
	DeltaToolArgs   = agentcore.DeltaToolArgs
	DeltaToolResult = agentcore.DeltaToolResult
)

type ActionType = agentcore.ActionType

const (
	ActionTransfer             = agentcore.ActionTransfer
	ActionInterrupted          = agentcore.ActionInterrupted
	ActionExit                 = agentcore.ActionExit
	ActionAskQuestions         = agentcore.ActionAskQuestions
	ActionAskResponse          = agentcore.ActionAskResponse
	ActionApprovalRequired     = agentcore.ActionApprovalRequired
	ActionApprovalDecision     = agentcore.ActionApprovalDecision
	ActionContextCompressStart = agentcore.ActionContextCompressStart
	ActionContextCompress      = agentcore.ActionContextCompress
)

type NotifyType = agentcore.NotifyType

const (
	NotifyProcessingStart  = agentcore.NotifyProcessingStart
	NotifyProcessingEnd    = agentcore.NotifyProcessingEnd
	NotifyUserMessage      = agentcore.NotifyUserMessage
	NotifyQueueUpdated     = agentcore.NotifyQueueUpdated
	NotifyCancelled        = agentcore.NotifyCancelled
	NotifyError            = agentcore.NotifyError
	NotifyAskQuestions     = agentcore.NotifyAskQuestions
	NotifyApprovalRequired = agentcore.NotifyApprovalRequired
	NotifyConnected        = agentcore.NotifyConnected
	NotifyPong             = agentcore.NotifyPong
	NotifyInvalidAPIKey    = agentcore.NotifyInvalidAPIKey
)

type Event = agentcore.Event
