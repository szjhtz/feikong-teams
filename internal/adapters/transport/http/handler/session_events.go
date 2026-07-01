package handler

import (
	"fmt"

	eventlog "fkteams/internal/adapters/storage/file/history"
	domainmessage "fkteams/internal/domain/message"
	"fkteams/internal/runtime/events"
)

func (rt *Runtime) transcriptToChatEvents(sessionID string, transcript []eventlog.TranscriptEvent) []map[string]any {
	result := make([]map[string]any, 0, len(transcript))
	for index, item := range transcript {
		for _, event := range transcriptEventToRuntimeEvents(sessionID, item) {
			event.Sequence = int64(len(result) + 1)
			payload := rt.convertEventToMap(event)
			payload["session_id"] = sessionID
			payload["transcript_index"] = index
			if item.AgentRunID != "" {
				payload["agent_run_id"] = item.AgentRunID
			}
			if item.Payload.ResultRef != "" {
				payload["result_ref"] = item.Payload.ResultRef
			}
			if item.Payload.Summary != "" {
				payload["summary"] = item.Payload.Summary
			}
			if item.Payload.Truncated {
				payload["truncated"] = true
				payload["original_chars"] = item.Payload.OriginalChars
			}
			if len(item.Payload.ContentParts) > 0 {
				payload["content_parts"] = append([]domainmessage.ContentPart(nil), item.Payload.ContentParts...)
			}
			result = append(result, payload)
		}
	}
	return result
}

func transcriptEventToRuntimeEvents(sessionID string, item eventlog.TranscriptEvent) []events.Event {
	base := events.Event{
		EventID:          item.ID,
		CreatedAt:        item.TS,
		TurnID:           item.TurnID,
		MessageID:        item.MessageID,
		AgentName:        item.Agent,
		ToolCallID:       item.ToolCallID,
		ParentToolCallID: item.ParentToolCallID,
	}
	if base.TurnID == "" {
		base.TurnID = fmt.Sprintf("%s:history:turn:1", sessionID)
	}
	switch item.Type {
	case eventlog.TranscriptUserMessage:
		base.Type = events.EventUserMessage
		base.Role = domainmessage.RoleUser
		base.Content = item.Payload.Content
		return []events.Event{base}
	case eventlog.TranscriptAssistantReasoning:
		base.Type = events.EventAssistantReasoning
		base.Role = domainmessage.RoleAssistant
		base.DeltaKind = events.DeltaReasoning
		base.Content = item.Payload.Content
		base.ReasoningContent = item.Payload.ReasoningContent
		if base.ReasoningContent == "" {
			base.ReasoningContent = item.Payload.Content
		}
		return []events.Event{base}
	case eventlog.TranscriptAssistantTextDelta:
		base.Type = events.EventAssistantText
		base.Role = domainmessage.RoleAssistant
		base.DeltaKind = events.DeltaOutput
		base.Content = item.Payload.Content
		return []events.Event{base}
	case eventlog.TranscriptToolCallStart:
		base.Type = events.EventToolCallStarted
		attachTranscriptToolCall(&base, item)
		return []events.Event{base}
	case eventlog.TranscriptToolCallEnd:
		base.Type = events.EventToolCallCompleted
		attachTranscriptToolCall(&base, item)
		base.Content = transcriptResultContent(item)
		base.ToolResult = base.Content
		return []events.Event{base}
	case eventlog.TranscriptUsageReported:
		base.Type = events.EventUsageReported
		if item.Payload.Usage != nil {
			base.PromptTokens = item.Payload.Usage.PromptTokens
			base.CompletionTokens = item.Payload.Usage.CompletionTokens
			base.TotalTokens = item.Payload.Usage.TotalTokens
			base.Usage = &events.UsagePayload{
				PromptTokens:     item.Payload.Usage.PromptTokens,
				CompletionTokens: item.Payload.Usage.CompletionTokens,
				TotalTokens:      item.Payload.Usage.TotalTokens,
			}
		}
		return []events.Event{base}
	case eventlog.TranscriptAskRequested, eventlog.TranscriptAskAnswered:
		base.Type = events.EventAskRequested
		if item.Type == eventlog.TranscriptAskAnswered {
			base.Type = events.EventAskAnswered
		}
		if item.Payload.Ask != nil {
			base.Ask = &events.AskPayload{
				ID:          item.Payload.Ask.ID,
				Question:    item.Payload.Ask.Question,
				Options:     append([]string(nil), item.Payload.Ask.Options...),
				MultiSelect: item.Payload.Ask.MultiSelect,
				Selected:    append([]string(nil), item.Payload.Ask.Selected...),
				FreeText:    item.Payload.Ask.FreeText,
			}
		}
		base.Content = item.Payload.Content
		return []events.Event{base}
	case eventlog.TranscriptSystemNotice:
		base.Type = events.EventSystemNotice
		base.Content = item.Payload.Content
		base.Detail = item.Payload.Detail
		base.Notice = &events.NoticePayload{Message: item.Payload.Content}
		return []events.Event{base}
	case eventlog.TranscriptError:
		base.Type = events.EventError
		base.Content = item.Payload.Content
		if item.Payload.Error != nil {
			base.Error = item.Payload.Error.Message
		}
		return []events.Event{base}
	case eventlog.TranscriptCancelled:
		base.Type = events.EventCancelled
		base.Content = item.Payload.Content
		return []events.Event{base}
	default:
		return nil
	}
}

func attachTranscriptToolCall(event *events.Event, item eventlog.TranscriptEvent) {
	if event == nil {
		return
	}
	id := item.ToolCallID
	name := item.Payload.ToolName
	args := item.Payload.ToolArgs
	if item.Payload.ToolCall != nil {
		if id == "" {
			id = item.Payload.ToolCall.ID
		}
		if name == "" {
			name = item.Payload.ToolCall.Name
		}
		if args == "" {
			args = item.Payload.ToolCall.Arguments
		}
	}
	event.ToolCallID = id
	if id != "" {
		event.ToolCallRef = "tool_call:" + id
	}
	event.ToolName = name
	event.ToolArgs = args
	event.ToolCall = &domainmessage.ToolCall{
		ID: id,
		Function: domainmessage.FunctionCall{
			Name:      name,
			Arguments: args,
		},
	}
	event.ToolCalls = []domainmessage.ToolCall{*event.ToolCall}
}

func transcriptResultContent(item eventlog.TranscriptEvent) string {
	if item.Payload.Result != "" {
		return item.Payload.Result
	}
	if item.Payload.Summary != "" {
		return item.Payload.Summary
	}
	if item.Payload.ResultRef != "" {
		return fmt.Sprintf("[tool result stored at %s]", item.Payload.ResultRef)
	}
	return ""
}
