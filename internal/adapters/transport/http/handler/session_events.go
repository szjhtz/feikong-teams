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
			if item.ResultRef != "" {
				payload["result_ref"] = item.ResultRef
			}
			if item.Summary != "" {
				payload["summary"] = item.Summary
			}
			if item.Truncated {
				payload["truncated"] = true
				payload["original_chars"] = item.OriginalChars
			}
			if len(item.ContentParts) > 0 {
				payload["content_parts"] = append([]domainmessage.ContentPart(nil), item.ContentParts...)
			}
			result = append(result, payload)
		}
	}
	return result
}

func transcriptEventToRuntimeEvents(sessionID string, item eventlog.TranscriptEvent) []events.Event {
	base := events.Event{
		EventID:          item.ID,
		CreatedAt:        item.At,
		TurnID:           transcriptRuntimeTurnID(sessionID, item.Turn),
		MessageID:        item.ID,
		AgentName:        item.Agent,
		ToolCallID:       item.ToolCallID,
		ParentToolCallID: item.ParentToolCallID,
	}
	switch item.Type {
	case eventlog.TranscriptUserMessage:
		base.Type = events.EventUserMessage
		base.Role = domainmessage.RoleUser
		base.Content = item.Content
		return []events.Event{base}
	case eventlog.TranscriptAssistantMessage:
		base.Type = events.EventAssistantCompleted
		base.Role = domainmessage.RoleAssistant
		base.Content = item.Content
		base.ReasoningContent = item.Reasoning
		base.Message = &domainmessage.Message{
			Role:             domainmessage.RoleAssistant,
			Content:          item.Content,
			ReasoningContent: item.Reasoning,
			ContentParts:     append([]domainmessage.ContentPart(nil), item.ContentParts...),
		}
		if item.Usage == nil {
			return []events.Event{base}
		}
		base.PromptTokens = item.Usage.PromptTokens
		base.CompletionTokens = item.Usage.CompletionTokens
		base.TotalTokens = item.Usage.TotalTokens
		base.Usage = &events.UsagePayload{
			PromptTokens:     item.Usage.PromptTokens,
			CompletionTokens: item.Usage.CompletionTokens,
			TotalTokens:      item.Usage.TotalTokens,
		}
		usage := events.Usage(item.Agent, "", item.Usage.PromptTokens, item.Usage.CompletionTokens, item.Usage.TotalTokens)
		usage.EventID = item.ID + ":usage"
		usage.CreatedAt = item.At
		usage.TurnID = base.TurnID
		usage.MessageID = item.ID
		usage.AgentName = item.Agent
		usage.ParentToolCallID = item.ParentToolCallID
		return []events.Event{base, usage}
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
	case eventlog.TranscriptAskRequested, eventlog.TranscriptAskAnswered:
		base.Type = events.EventAskRequested
		if item.Type == eventlog.TranscriptAskAnswered {
			base.Type = events.EventAskAnswered
		}
		if item.Ask != nil {
			base.Ask = &events.AskPayload{
				ID:          item.Ask.ID,
				Question:    item.Ask.Question,
				Options:     append([]string(nil), item.Ask.Options...),
				MultiSelect: item.Ask.MultiSelect,
				Selected:    append([]string(nil), item.Ask.Selected...),
				FreeText:    item.Ask.FreeText,
			}
		}
		base.Content = item.Content
		return []events.Event{base}
	case eventlog.TranscriptSystemNotice:
		base.Type = events.EventSystemNotice
		base.Content = item.Content
		base.Detail = item.Detail
		base.Notice = &events.NoticePayload{Message: item.Content}
		return []events.Event{base}
	case eventlog.TranscriptError:
		base.Type = events.EventError
		base.Content = item.Content
		if item.Error != nil {
			base.Error = item.Error.Message
		}
		return []events.Event{base}
	case eventlog.TranscriptCancelled:
		base.Type = events.EventCancelled
		base.Content = item.Content
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
	name := item.ToolName
	args := item.ToolArgs
	if item.ToolCall != nil {
		if id == "" {
			id = item.ToolCall.ID
		}
		if name == "" {
			name = item.ToolCall.Name
		}
		if args == "" {
			args = item.ToolCall.Arguments
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
	if item.Result != "" {
		return item.Result
	}
	if item.Summary != "" {
		return item.Summary
	}
	if item.ResultRef != "" {
		return fmt.Sprintf("[tool result stored at %s]", item.ResultRef)
	}
	return ""
}

func transcriptRuntimeTurnID(sessionID string, turn int) string {
	if turn <= 0 {
		turn = 1
	}
	if sessionID == "" {
		return fmt.Sprintf("history:turn:%d", turn)
	}
	return fmt.Sprintf("%s:history:turn:%d", sessionID, turn)
}
