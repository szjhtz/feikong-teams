package handler

import (
	"fmt"
	"sort"
	"strings"

	eventlog "fkteams/internal/adapters/storage/file/history"
	domainevent "fkteams/internal/domain/event"
	domainmessage "fkteams/internal/domain/message"
	"fkteams/internal/runtime/events"
)

func (rt *Runtime) historyLinesToChatEvents(sessionID string, lines []eventlog.HistoryLine) []map[string]any {
	emitted := make([]orderedHistoryEvent, 0)
	memberCompletionKeys := historyMemberCompletionKeys(lines)
	emissionIndex := 0
	for _, item := range orderedHistoryLines(lines) {
		line := item.line
		msg := historyMessageFromLine(line)
		historyEvent := line.Event
		eventsForLine := historyEventToChatEvents(sessionID, line.MessageID, item.index, line.EventIndex, item.turn, msg, historyEvent)
		for eventIndex, event := range eventsForLine {
			key := item.key + float64(eventIndex)/10_000
			if event.Type == events.EventToolCallCompleted {
				if completionKey, ok := historyToolCompletionKey(event, memberCompletionKeys); ok && completionKey > key {
					key = completionKey + 0.0001
				}
			}
			emitted = append(emitted, orderedHistoryEvent{
				event:        event,
				historyEvent: historyEvent,
				key:          key,
				index:        emissionIndex,
			})
			emissionIndex++
		}
	}
	sort.SliceStable(emitted, func(i, j int) bool {
		if emitted[i].key == emitted[j].key {
			return emitted[i].index < emitted[j].index
		}
		return emitted[i].key < emitted[j].key
	})

	result := make([]map[string]any, 0, len(emitted))
	for order, item := range emitted {
		if item.event.Sequence == 0 {
			item.event.Sequence = int64(order + 1)
		}
		payload := rt.convertEventToMap(item.event)
		payload["session_id"] = sessionID
		if len(item.historyEvent.ContentParts) > 0 {
			payload["content_parts"] = append([]domainmessage.ContentPart(nil), item.historyEvent.ContentParts...)
		}
		applyHistoryToolDisplay(payload, item.historyEvent.ToolCall)
		result = append(result, payload)
	}
	return result
}

type orderedHistoryEvent struct {
	event        events.Event
	historyEvent eventlog.MessageEvent
	key          float64
	index        int
}

type orderedHistoryLine struct {
	line  eventlog.HistoryLine
	key   float64
	turn  int
	index int
}

func orderedHistoryLines(lines []eventlog.HistoryLine) []orderedHistoryLine {
	items := make([]orderedHistoryLine, 0, len(lines))
	turn := -1
	lastSequencedKey := make(map[int]float64)
	for index, line := range lines {
		if isHistoryUserAgent(line.AgentName) {
			turn++
		}
		if turn < 0 {
			turn = 0
		}
		turnBase := float64(turn) * 1_000_000_000
		key := turnBase + float64(index+1)/1_000_000
		if line.Event.Sequence > 0 {
			key = turnBase + float64(line.Event.Sequence)
			lastSequencedKey[turn] = key
		} else if lastKey, ok := lastSequencedKey[turn]; ok && !isHistoryUserAgent(line.AgentName) {
			key = lastKey + float64(index+1)/1_000_000
		}
		items = append(items, orderedHistoryLine{line: line, key: key, turn: turn, index: index})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].key == items[j].key {
			return items[i].index < items[j].index
		}
		return items[i].key < items[j].key
	})
	return items
}

func historyMessageFromLine(line eventlog.HistoryLine) eventlog.AgentMessage {
	return eventlog.AgentMessage{
		AgentName:      line.AgentName,
		RunPath:        line.RunPath,
		MemberCallID:   line.MemberCallID,
		MemberToolName: line.MemberToolName,
		MemberName:     line.MemberName,
		StartTime:      line.StartTime,
		EndTime:        line.EndTime,
	}
}

func historyMemberCompletionKeys(lines []eventlog.HistoryLine) map[string]float64 {
	result := make(map[string]float64)
	for _, item := range orderedHistoryLines(lines) {
		if item.line.MemberCallID == "" || item.line.Event.Sequence <= 0 {
			continue
		}
		for _, key := range historyMemberKeys(item.line) {
			if item.key > result[key] {
				result[key] = item.key
			}
		}
	}
	return result
}

func historyMemberKeys(line eventlog.HistoryLine) []string {
	keys := []string{line.MemberCallID}
	if line.MemberCallID != "" && !strings.HasPrefix(line.MemberCallID, "tool_call:") {
		keys = append(keys, fmt.Sprintf("tool_call:%s", line.MemberCallID))
	}
	if line.MemberToolName != "" {
		keys = append(keys, line.MemberToolName)
	}
	return keys
}

func historyToolCompletionKey(event events.Event, memberCompletionKeys map[string]float64) (float64, bool) {
	for _, key := range []string{event.ToolCallID, event.ToolCallRef, event.ToolName} {
		if key == "" {
			continue
		}
		if value, ok := memberCompletionKeys[key]; ok {
			return value, true
		}
	}
	return 0, false
}

func historyEventToChatEvents(sessionID, messageID string, msgIndex, eventIndex, turn int, msg eventlog.AgentMessage, historyEvent eventlog.MessageEvent) []events.Event {
	base := historyEventBase(sessionID, messageID, msgIndex, eventIndex, turn, msg, historyEvent)
	switch historyEvent.Type {
	case eventlog.MsgTypeText:
		base.Content = historyEvent.Content
		if isHistoryUserAgent(msg.AgentName) {
			base.Type = events.EventUserMessage
			base.Role = domainmessage.RoleUser
		} else {
			base.Type = events.EventAssistantText
			base.Role = domainmessage.RoleAssistant
			base.DeltaKind = events.DeltaOutput
		}
		return []events.Event{base}
	case eventlog.MsgTypeReasoning:
		base.Type = events.EventAssistantReasoning
		base.Role = domainmessage.RoleAssistant
		base.DeltaKind = events.DeltaReasoning
		base.Content = historyEvent.Content
		base.ReasoningContent = historyEvent.Content
		return []events.Event{base}
	case eventlog.MsgTypeToolCall:
		if historyEvent.ToolCall == nil {
			base.Type = events.EventToolCallCompleted
			base.Content = historyEvent.Content
			return []events.Event{base}
		}
		started := base
		completed := base
		attachHistoryToolCall(&started, historyEvent.ToolCall)
		attachHistoryToolCall(&completed, historyEvent.ToolCall)
		started.Type = events.EventToolCallStarted
		completed.Type = events.EventToolCallCompleted
		completed.ToolResult = historyEvent.ToolCall.Result
		completed.Content = historyEvent.ToolCall.Result
		completed.EventID = fmt.Sprintf("%s:end", base.EventID)
		return []events.Event{started, completed}
	case eventlog.MsgTypeAsk:
		if historyEvent.Ask == nil {
			base.Type = events.EventAskRequested
			base.Content = historyEvent.Content
			return []events.Event{base}
		}
		base.Type = events.EventAskRequested
		if historyEvent.Ask.Answered {
			base.Type = events.EventAskAnswered
		}
		base.Ask = &events.AskPayload{
			ID:          historyEvent.Ask.ID,
			Question:    historyEvent.Ask.Question,
			Options:     append([]string(nil), historyEvent.Ask.Options...),
			MultiSelect: historyEvent.Ask.MultiSelect,
			Selected:    append([]string(nil), historyEvent.Ask.Selected...),
			FreeText:    historyEvent.Ask.FreeText,
		}
		base.Content = historyAskContent(historyEvent.Ask)
		return []events.Event{base}
	case eventlog.MsgTypeNotice:
		base.Type = events.EventSystemNotice
		base.Content = historyEvent.Content
		base.Detail = historyEvent.Detail
		base.Notice = &events.NoticePayload{Message: historyEvent.Content}
		return []events.Event{base}
	case eventlog.MsgTypeUsageReported:
		base.Type = events.EventUsageReported
		if historyEvent.Usage != nil {
			base.PromptTokens = historyEvent.Usage.PromptTokens
			base.CompletionTokens = historyEvent.Usage.CompletionTokens
			base.TotalTokens = historyEvent.Usage.TotalTokens
			base.Usage = &events.UsagePayload{
				PromptTokens:     historyEvent.Usage.PromptTokens,
				CompletionTokens: historyEvent.Usage.CompletionTokens,
				TotalTokens:      historyEvent.Usage.TotalTokens,
			}
		}
		return []events.Event{base}
	case eventlog.MsgTypeError:
		base.Type = events.EventError
		base.Content = historyEvent.Content
		base.Error = historyErrorText(historyEvent)
		return []events.Event{base}
	case eventlog.MsgTypeCancelled:
		base.Type = domainevent.Type(domainevent.NotifyCancelled)
		base.Content = historyEvent.Content
		return []events.Event{base}
	default:
		base.Type = events.EventSystemNotice
		base.Content = historyEvent.Content
		base.Detail = historyEvent.Detail
		return []events.Event{base}
	}
}

func historyEventBase(sessionID, messageID string, msgIndex, eventIndex, turn int, msg eventlog.AgentMessage, historyEvent eventlog.MessageEvent) events.Event {
	createdAt := msg.StartTime
	if !historyEvent.CreatedAt.IsZero() {
		createdAt = historyEvent.CreatedAt
	}
	if createdAt.IsZero() {
		createdAt = msg.EndTime
	}
	runID := fmt.Sprintf("%s:history:turn:%d", sessionID, turn+1)
	if historyEvent.RunID != "" {
		runID = historyEvent.RunID
	}
	turnID := historyEvent.TurnID
	if turnID == "" {
		turnID = events.TurnID(runID, 1)
	}
	resolvedMessageID := messageID
	if historyEvent.MessageID != "" {
		resolvedMessageID = historyEvent.MessageID
	}
	eventID := historyEvent.EventID
	if eventID == "" {
		eventID = fmt.Sprintf("history:%s:%06d:%04d:%s", sessionID, msgIndex, eventIndex, historyEvent.Type)
	}
	event := events.Event{
		EventID:          eventID,
		Sequence:         historyEvent.Sequence,
		CreatedAt:        createdAt,
		RunID:            runID,
		TurnID:           turnID,
		MessageID:        resolvedMessageID,
		AgentName:        msg.AgentName,
		RunPath:          msg.RunPath,
		MemberCallID:     msg.MemberCallID,
		MemberToolName:   msg.MemberToolName,
		MemberName:       msg.MemberName,
		ParentToolCallID: msg.MemberCallID,
		ParentToolName:   msg.MemberToolName,
		Detail:           historyEvent.Detail,
	}
	if isHistorySystemAgent(msg.AgentName) {
		event.Role = domainmessage.RoleSystem
	}
	return event
}

func attachHistoryToolCall(event *events.Event, record *eventlog.ToolCallRecord) {
	if event == nil || record == nil {
		return
	}
	ref := record.Ref
	if ref == "" && record.ID != "" {
		ref = fmt.Sprintf("tool_call:%s", record.ID)
	}
	event.ToolCallID = record.ID
	event.ToolCallRef = ref
	event.ToolCallIndex = record.Index
	event.ToolName = record.Name
	event.ToolArgs = record.Arguments
	event.ToolCall = &domainmessage.ToolCall{
		ID:    record.ID,
		Index: record.Index,
		Type:  "function",
		Function: domainmessage.FunctionCall{
			Name:      record.Name,
			Arguments: record.Arguments,
		},
	}
	event.ToolCalls = []domainmessage.ToolCall{*event.ToolCall}
}

func applyHistoryToolDisplay(payload map[string]any, record *eventlog.ToolCallRecord) {
	if record == nil {
		return
	}
	if record.DisplayName != "" {
		payload["tool_display_name"] = record.DisplayName
	}
	if record.Kind != "" {
		payload["tool_kind"] = record.Kind
	}
	if record.Target != "" {
		payload["tool_target"] = record.Target
	}
	if call, ok := payload["tool_call"].(map[string]any); ok {
		applyHistoryToolCallDisplay(call, record)
	}
	if calls, ok := payload["tool_calls"].([]map[string]any); ok {
		for _, call := range calls {
			applyHistoryToolCallDisplay(call, record)
		}
	}
}

func applyHistoryToolCallDisplay(call map[string]any, record *eventlog.ToolCallRecord) {
	if record.DisplayName != "" {
		call["display_name"] = record.DisplayName
	}
	if record.Kind != "" {
		call["kind"] = record.Kind
	}
	if record.Target != "" {
		call["target"] = record.Target
	}
	if record.Result != "" {
		call["result"] = record.Result
	}
}

func isHistoryUserAgent(agentName string) bool {
	name := strings.TrimSpace(strings.ToLower(agentName))
	return name == "user" || name == "用户"
}

func isHistorySystemAgent(agentName string) bool {
	name := strings.TrimSpace(strings.ToLower(agentName))
	return name == "system" || name == "系统"
}

func historyAskContent(record *eventlog.AskRecord) string {
	if record == nil {
		return ""
	}
	if !record.Answered {
		return record.Question
	}
	parts := append([]string(nil), record.Selected...)
	if strings.TrimSpace(record.FreeText) != "" {
		parts = append(parts, record.FreeText)
	}
	return strings.Join(parts, "；")
}

func historyErrorText(historyEvent eventlog.MessageEvent) string {
	if historyEvent.Error == nil {
		return historyEvent.Content
	}
	if historyEvent.Error.TechnicalDetail != "" {
		return historyEvent.Error.TechnicalDetail
	}
	if historyEvent.Error.Message != "" {
		return historyEvent.Error.Message
	}
	if historyEvent.Error.Title != "" {
		return historyEvent.Error.Title
	}
	return historyEvent.Content
}
