package eventlog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fkteams/internal/domain/message"
)

func (h *HistoryRecorder) SaveToFile(filePath string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.sessionDir == "" {
		h.sessionDir = filepath.Dir(filePath)
	}
	if err := os.MkdirAll(h.sessionDir, 0755); err != nil {
		return fmt.Errorf("create session dir: %w", err)
	}
	return nil
}

func (h *HistoryRecorder) LoadFromFile(filePath string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.sessionDir = filepath.Dir(filePath)
	events, err := loadTranscript(filePath)
	if err != nil {
		return err
	}
	h.messages = projectTranscriptEvents(h.sessionDir, events)
	h.reconstructSummaryFromEvents()
	h.activeMessages = make(map[string]*activeMessageContext)
	h.activeOrder = nil
	h.subagents = make(map[string]*subagentRun)
	h.agentToolCalls = make(map[string]pendingToolCall)
	return nil
}

func LoadTranscriptFromFile(filePath string) ([]TranscriptEvent, error) {
	return loadTranscript(filePath)
}

func projectTranscriptEvents(sessionDir string, events []TranscriptEvent) []AgentMessage {
	var messages []AgentMessage
	var current *AgentMessage
	toolEventIndex := make(map[string]int)
	sessionID := filepath.Base(sessionDir)
	flush := func() {
		if current == nil || len(current.Events) == 0 {
			current = nil
			return
		}
		if current.EndTime.IsZero() {
			current.EndTime = time.Now()
		}
		messages = append(messages, *current)
		current = nil
		toolEventIndex = make(map[string]int)
	}
	ensure := func(agent string, ts time.Time) *AgentMessage {
		if agent == "" {
			agent = "coordinator"
		}
		if current == nil || current.AgentName != agent {
			flush()
			current = &AgentMessage{AgentName: agent, StartTime: ts, Events: make([]MessageEvent, 0)}
		}
		if current.StartTime.IsZero() {
			current.StartTime = ts
		}
		current.EndTime = ts
		return current
	}

	for _, event := range events {
		ts := event.At
		if ts.IsZero() {
			ts = time.Now()
		}
		turnID := transcriptTurnID(sessionID, event.Turn)
		messageID := event.ID
		switch event.Type {
		case TranscriptUserMessage:
			flush()
			messages = append(messages, AgentMessage{
				AgentName: "user",
				StartTime: ts,
				EndTime:   ts,
				Events: []MessageEvent{{
					Type:         MsgTypeText,
					CreatedAt:    ts,
					TurnID:       turnID,
					MessageID:    messageID,
					Content:      event.Content,
					ContentParts: append([]message.ContentPart(nil), event.ContentParts...),
				}},
			})
		case TranscriptAssistantMessage:
			msg := ensure(event.Agent, ts)
			if event.Reasoning != "" {
				msg.Events = append(msg.Events, MessageEvent{Type: MsgTypeReasoning, CreatedAt: ts, TurnID: turnID, MessageID: messageID, Content: event.Reasoning})
			}
			if event.Content != "" || len(event.ContentParts) > 0 {
				msg.Events = append(msg.Events, MessageEvent{Type: MsgTypeText, CreatedAt: ts, TurnID: turnID, MessageID: messageID, Content: event.Content, ContentParts: append([]message.ContentPart(nil), event.ContentParts...)})
			}
			if event.Usage != nil {
				msg.Events = append(msg.Events, MessageEvent{Type: MsgTypeUsageReported, CreatedAt: ts, TurnID: turnID, MessageID: messageID, Usage: event.Usage})
			}
		case TranscriptToolCallStart:
			msg := ensure(event.Agent, ts)
			record := transcriptToolCallRecord(event, "")
			msg.Events = append(msg.Events, MessageEvent{Type: MsgTypeToolCall, CreatedAt: ts, TurnID: turnID, MessageID: messageID, ToolCall: &record})
			if event.ToolCallID != "" {
				toolEventIndex[event.ToolCallID] = len(msg.Events) - 1
			}
		case TranscriptToolCallEnd:
			msg := ensure(event.Agent, ts)
			result := transcriptToolResult(event)
			if idx, ok := toolEventIndex[event.ToolCallID]; ok && idx >= 0 && idx < len(msg.Events) && msg.Events[idx].ToolCall != nil {
				msg.Events[idx].ToolCall.Result = result
				continue
			}
			record := transcriptToolCallRecord(event, result)
			msg.Events = append(msg.Events, MessageEvent{Type: MsgTypeToolCall, CreatedAt: ts, TurnID: turnID, MessageID: messageID, ToolCall: &record})
		case TranscriptAskRequested, TranscriptAskAnswered:
			msg := ensure(event.Agent, ts)
			msg.Events = append(msg.Events, MessageEvent{Type: MsgTypeAsk, CreatedAt: ts, TurnID: turnID, MessageID: messageID, Content: event.Content, Ask: event.Ask})
		case TranscriptSystemNotice:
			msg := ensure(event.Agent, ts)
			msg.Events = append(msg.Events, MessageEvent{Type: MsgTypeNotice, CreatedAt: ts, TurnID: turnID, MessageID: messageID, Content: event.Content, Detail: event.Detail})
		case TranscriptError:
			msg := ensure(event.Agent, ts)
			content := event.Content
			if content == "" && event.Error != nil {
				content = event.Error.Message
			}
			msg.Events = append(msg.Events, MessageEvent{Type: MsgTypeError, CreatedAt: ts, TurnID: turnID, MessageID: messageID, Content: content, Error: event.Error})
		case TranscriptCancelled:
			msg := ensure(event.Agent, ts)
			msg.Events = append(msg.Events, MessageEvent{Type: MsgTypeCancelled, CreatedAt: ts, TurnID: turnID, MessageID: messageID, Content: event.Content})
		}
	}
	flush()
	messages = append(messages, projectSubagentTranscriptFiles(sessionDir)...)
	return messages
}

func projectSubagentTranscriptFiles(sessionDir string) []AgentMessage {
	matches, err := filepath.Glob(filepath.Join(sessionDir, subagentsDirName, "*.jsonl"))
	if err != nil || len(matches) == 0 {
		return nil
	}
	messages := make([]AgentMessage, 0, len(matches))
	for _, filePath := range matches {
		events, err := loadTranscript(filePath)
		if err != nil {
			continue
		}
		msg := projectSubagentTranscript(events)
		if len(msg.Events) > 0 {
			messages = append(messages, msg)
		}
	}
	return messages
}

func projectSubagentTranscript(events []TranscriptEvent) AgentMessage {
	agent := ""
	parentToolCallID := ""
	parentToolName := ""
	for _, event := range events {
		if agent == "" {
			agent = event.Agent
		}
		if parentToolCallID == "" {
			parentToolCallID = event.ParentToolCallID
		}
		if parentToolName == "" {
			parentToolName = event.ToolName
		}
	}
	if agent == "" {
		agent = "member"
	}
	msg := AgentMessage{
		AgentName:      agent,
		MemberCallID:   parentToolCallID,
		MemberToolName: parentToolName,
		MemberName:     agent,
		Events:         make([]MessageEvent, 0),
	}
	toolEventIndex := make(map[string]int)
	for _, event := range events {
		ts := event.At
		turnID := transcriptTurnID("", event.Turn)
		messageID := event.ID
		if msg.StartTime.IsZero() {
			msg.StartTime = ts
		}
		msg.EndTime = ts
		switch event.Type {
		case TranscriptAssistantMessage:
			if event.Reasoning != "" {
				msg.Events = append(msg.Events, MessageEvent{Type: MsgTypeReasoning, CreatedAt: ts, TurnID: turnID, MessageID: messageID, Content: event.Reasoning})
			}
			if event.Content != "" || len(event.ContentParts) > 0 {
				msg.Events = append(msg.Events, MessageEvent{Type: MsgTypeText, CreatedAt: ts, TurnID: turnID, MessageID: messageID, Content: event.Content, ContentParts: append([]message.ContentPart(nil), event.ContentParts...)})
			}
			if event.Usage != nil {
				msg.Events = append(msg.Events, MessageEvent{Type: MsgTypeUsageReported, CreatedAt: ts, TurnID: turnID, MessageID: messageID, Usage: event.Usage})
			}
		case TranscriptToolCallStart:
			record := transcriptToolCallRecord(event, "")
			msg.Events = append(msg.Events, MessageEvent{Type: MsgTypeToolCall, CreatedAt: ts, TurnID: turnID, MessageID: messageID, ToolCall: &record})
			if event.ToolCallID != "" {
				toolEventIndex[event.ToolCallID] = len(msg.Events) - 1
			}
		case TranscriptToolCallEnd:
			result := transcriptToolResult(event)
			if idx, ok := toolEventIndex[event.ToolCallID]; ok && idx >= 0 && idx < len(msg.Events) && msg.Events[idx].ToolCall != nil {
				msg.Events[idx].ToolCall.Result = result
				continue
			}
			record := transcriptToolCallRecord(event, result)
			msg.Events = append(msg.Events, MessageEvent{Type: MsgTypeToolCall, CreatedAt: ts, TurnID: turnID, MessageID: messageID, ToolCall: &record})
		case TranscriptSystemNotice:
			msg.Events = append(msg.Events, MessageEvent{Type: MsgTypeNotice, CreatedAt: ts, TurnID: turnID, MessageID: messageID, Content: event.Content, Detail: event.Detail})
		case TranscriptError:
			msg.Events = append(msg.Events, MessageEvent{Type: MsgTypeError, CreatedAt: ts, TurnID: turnID, MessageID: messageID, Content: event.Content, Error: event.Error})
		}
	}
	return msg
}

func transcriptToolCallRecord(event TranscriptEvent, result string) ToolCallRecord {
	if event.ToolCall != nil {
		record := *event.ToolCall
		record.Result = result
		if record.Result == "" {
			record.Result = transcriptToolResult(event)
		}
		return record
	}
	return ToolCallRecord{
		Ref:         toolCallRef(event.ToolCallID),
		ID:          event.ToolCallID,
		Name:        event.ToolName,
		DisplayName: event.DisplayName,
		Kind:        event.Kind,
		Target:      event.Target,
		Arguments:   event.ToolArgs,
		Result:      result,
	}
}

func transcriptToolResult(event TranscriptEvent) string {
	if event.Result != "" {
		return event.Result
	}
	if event.Summary != "" {
		return event.Summary
	}
	if event.ResultRef != "" {
		return fmt.Sprintf("[tool result stored at %s]", event.ResultRef)
	}
	return ""
}

func toolCallRef(id string) string {
	id = strings.TrimSpace(id)
	if id == "" || strings.HasPrefix(id, "tool_call:") {
		return id
	}
	return "tool_call:" + id
}
