package events

import (
	"fmt"

	"fkteams/internal/domain/message"
)

func ToolCallsFromEvent(event Event) []message.ToolCall {
	if event.ToolCall == nil {
		return event.ToolCalls
	}
	toolCalls := make([]message.ToolCall, 0, len(event.ToolCalls)+1)
	toolCalls = append(toolCalls, *event.ToolCall)
	toolCalls = append(toolCalls, event.ToolCalls...)
	return toolCalls
}

func ToolCallRefAt(event Event, tool message.ToolCall, position int) string {
	if tool.Index != nil && event.ToolCallRefs != nil {
		if ref := event.ToolCallRefs[*tool.Index]; ref != "" {
			return ref
		}
	}
	if event.ToolCallRefs != nil {
		if ref := event.ToolCallRefs[position]; ref != "" {
			return ref
		}
	}
	if event.ToolCall != nil && position == 0 && event.ToolCallRef != "" {
		return event.ToolCallRef
	}
	return ""
}

func ValidateEventContract(event Event) error {
	if event.Type == "" {
		return fmt.Errorf("event type is required")
	}
	switch event.Type {
	case EventTurnStart, EventTurnEnd:
		if event.RunID == "" || event.TurnID == "" {
			return fmt.Errorf("%s missing run or turn identity", event.Type)
		}
	case EventMessageStart, EventMessageDelta, EventMessageEnd:
		if event.Role == "" {
			return fmt.Errorf("%s missing message role", event.Type)
		}
		if event.Type == EventMessageDelta &&
			(event.DeltaKind == DeltaToolArgs || event.DeltaKind == DeltaToolResult) &&
			(event.ToolCallRef == "" || event.ToolCallID == "") {
			return fmt.Errorf("message_delta %s missing stable tool identity", event.DeltaKind)
		}
		if event.Type == EventMessageEnd && event.Role == message.RoleTool && (event.ToolCallRef == "" || event.ToolCallID == "") {
			return fmt.Errorf("tool message_end missing stable tool identity")
		}
		if event.Type == EventMessageEnd {
			for i, tool := range event.ToolCalls {
				if IsInternalToolName(tool.Function.Name) {
					continue
				}
				if tool.ID == "" {
					return fmt.Errorf("message_end tool call missing id at position %d", i)
				}
				if ToolCallRefAt(event, tool, i) == "" {
					return fmt.Errorf("message_end tool call missing ref at position %d", i)
				}
			}
		}
	case EventToolStart:
		if event.ToolName == "" {
			return fmt.Errorf("tool_start missing tool name")
		}
		if event.ToolCallRef == "" || event.ToolCallID == "" {
			return fmt.Errorf("tool_start missing stable tool identity")
		}
	case EventToolUpdate, EventToolEnd:
		if event.ToolName == "" {
			return fmt.Errorf("%s missing tool name", event.Type)
		}
		if event.ToolCallRef == "" || event.ToolCallID == "" {
			return fmt.Errorf("%s missing stable tool identity", event.Type)
		}
	case EventAction:
		if event.ActionType == "" {
			return fmt.Errorf("action missing action type")
		}
	case EventError:
		if event.Error == "" && event.Content == "" {
			return fmt.Errorf("error event missing error content")
		}
	case EventUsage:
		if event.PromptTokens == 0 && event.CompletionTokens == 0 && event.TotalTokens == 0 {
			return fmt.Errorf("usage event missing token counts")
		}
	}
	return nil
}
