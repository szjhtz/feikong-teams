package events

import (
	"fkteams/agentcore"
	"fmt"
)

type Emitter struct {
	runID     string
	turnID    string
	sink      agentcore.EventSink
	lastEvent agentcore.Event
}

func NewEmitter(runID, turnID string, sink agentcore.EventSink) *Emitter {
	if sink == nil {
		sink = agentcore.NoopEventSink
	}
	return &Emitter{
		runID:  runID,
		turnID: turnID,
		sink:   sink,
	}
}

func (e *Emitter) Emit(event agentcore.Event) error {
	event.RunID = firstNonEmpty(event.RunID, e.runID)
	event.TurnID = firstNonEmpty(event.TurnID, e.turnID)
	e.lastEvent = NormalizeEvent(event)
	if err := ValidateEventContract(e.lastEvent); err != nil {
		return err
	}
	return e.sink(e.lastEvent)
}

func (e *Emitter) LastEvent() agentcore.Event {
	if e == nil {
		return agentcore.Event{}
	}
	return e.lastEvent
}

func AgentStart(runID string) agentcore.Event {
	return agentcore.Event{Type: agentcore.EventAgentStart, RunID: runID}
}

func AgentEnd(runID string) agentcore.Event {
	return agentcore.Event{Type: agentcore.EventAgentEnd, RunID: runID}
}

func AgentError(runID string, err error) agentcore.Event {
	event := AgentEnd(runID)
	if err != nil {
		event.Error = err.Error()
	}
	return event
}

func TurnStart(runID, turnID string) agentcore.Event {
	return agentcore.Event{Type: agentcore.EventTurnStart, RunID: runID, TurnID: turnID}
}

func TurnEnd(runID, turnID string) agentcore.Event {
	return agentcore.Event{Type: agentcore.EventTurnEnd, RunID: runID, TurnID: turnID}
}

type MessageEvent struct {
	MessageID        string
	Role             agentcore.MessageRole
	AgentName        string
	RunPath          string
	Content          string
	DeltaKind        agentcore.DeltaKind
	Message          *agentcore.Message
	ToolCallID       string
	ToolCallRef      string
	ToolName         string
	ToolCalls        []agentcore.ToolCall
	ToolCallRefs     map[int]string
	ReasoningContent string
}

func MessageStart(meta MessageEvent) agentcore.Event {
	return agentcore.Event{
		Type:        agentcore.EventMessageStart,
		MessageID:   meta.MessageID,
		Role:        meta.Role,
		AgentName:   meta.AgentName,
		RunPath:     meta.RunPath,
		Content:     meta.Content,
		Message:     meta.Message,
		ToolCallID:  meta.ToolCallID,
		ToolCallRef: meta.ToolCallRef,
		ToolName:    meta.ToolName,
	}
}

func MessageDelta(meta MessageEvent, delta string) agentcore.Event {
	return agentcore.Event{
		Type:        agentcore.EventMessageDelta,
		MessageID:   meta.MessageID,
		Role:        meta.Role,
		AgentName:   meta.AgentName,
		RunPath:     meta.RunPath,
		DeltaKind:   meta.DeltaKind,
		Content:     delta,
		ToolCallID:  meta.ToolCallID,
		ToolCallRef: meta.ToolCallRef,
		ToolName:    meta.ToolName,
	}
}

func MessageEnd(meta MessageEvent) agentcore.Event {
	return agentcore.Event{
		Type:             agentcore.EventMessageEnd,
		MessageID:        meta.MessageID,
		Role:             meta.Role,
		AgentName:        meta.AgentName,
		RunPath:          meta.RunPath,
		Content:          meta.Content,
		ReasoningContent: meta.ReasoningContent,
		Message:          meta.Message,
		ToolCallID:       meta.ToolCallID,
		ToolCallRef:      meta.ToolCallRef,
		ToolName:         meta.ToolName,
		ToolCalls:        meta.ToolCalls,
		ToolCallRefs:     meta.ToolCallRefs,
	}
}

type ToolEvent struct {
	AgentName     string
	RunPath       string
	ToolCallID    string
	ToolCallRef   string
	ToolName      string
	ToolArgs      string
	ToolResult    string
	Content       string
	ToolCall      *agentcore.ToolCall
	ToolCallIndex *int
}

func ToolStart(meta ToolEvent) agentcore.Event {
	content := firstNonEmpty(meta.Content, meta.ToolArgs)
	return agentcore.Event{
		Type:          agentcore.EventToolStart,
		AgentName:     meta.AgentName,
		RunPath:       meta.RunPath,
		ToolCallID:    meta.ToolCallID,
		ToolCallRef:   meta.ToolCallRef,
		ToolName:      meta.ToolName,
		ToolArgs:      meta.ToolArgs,
		Content:       content,
		ToolCall:      meta.ToolCall,
		ToolCallIndex: meta.ToolCallIndex,
	}
}

func ToolUpdate(meta ToolEvent) agentcore.Event {
	return agentcore.Event{
		Type:        agentcore.EventToolUpdate,
		AgentName:   meta.AgentName,
		RunPath:     meta.RunPath,
		ToolCallID:  meta.ToolCallID,
		ToolCallRef: meta.ToolCallRef,
		ToolName:    meta.ToolName,
		Content:     meta.Content,
		DeltaKind:   agentcore.DeltaToolResult,
	}
}

func ToolEnd(meta ToolEvent) agentcore.Event {
	content := firstNonEmpty(meta.Content, meta.ToolResult)
	return agentcore.Event{
		Type:        agentcore.EventToolEnd,
		AgentName:   meta.AgentName,
		RunPath:     meta.RunPath,
		ToolCallID:  meta.ToolCallID,
		ToolCallRef: meta.ToolCallRef,
		ToolName:    meta.ToolName,
		Content:     content,
		ToolResult:  firstNonEmpty(meta.ToolResult, content),
	}
}

func Action(agentName, runPath string, actionType agentcore.ActionType, content string) agentcore.Event {
	return agentcore.Event{
		Type:       agentcore.EventAction,
		AgentName:  agentName,
		RunPath:    runPath,
		ActionType: actionType,
		Content:    content,
	}
}

func Error(agentName, runPath string, err error) agentcore.Event {
	event := agentcore.Event{
		Type:      agentcore.EventError,
		AgentName: agentName,
		RunPath:   runPath,
	}
	if err != nil {
		event.Error = err.Error()
	}
	return event
}

func Usage(agentName, runPath string, promptTokens, completionTokens, totalTokens int) agentcore.Event {
	return agentcore.Event{
		Type:             agentcore.EventUsage,
		AgentName:        agentName,
		RunPath:          runPath,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
	}
}

func UserMessagePair(runID, turnID, messageID string, message agentcore.Message) (agentcore.Event, agentcore.Event) {
	content := message.DisplayText()
	meta := MessageEvent{
		MessageID: messageID,
		Role:      agentcore.RoleUser,
		Content:   content,
		Message:   &message,
	}
	start := MessageStart(meta)
	end := MessageEnd(meta)
	start.RunID = runID
	start.TurnID = turnID
	end.RunID = runID
	end.TurnID = turnID
	return start, end
}

func TurnID(runID string, index int) string {
	return fmt.Sprintf("%s:turn:%d", runID, index)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
