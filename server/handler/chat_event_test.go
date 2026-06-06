package handler

import (
	"testing"
	"time"

	"fkteams/agentcore"
	"fkteams/agents/toolmeta"
	"fkteams/events"
)

func TestConvertEventToMapKeepsFrontendStreamAndMemberMetadata(t *testing.T) {
	toolIndex := 0
	toolName := "ask_fkagent_event_flow_member"
	toolmeta.RegisterAgentToolDisplay(toolName, "Event Flow Member")
	event := events.NormalizeEvent(events.Event{
		Type:             events.EventMessageDelta,
		Sequence:         42,
		CreatedAt:        time.Unix(100, 0).UTC(),
		MessageID:        "msg_member_1",
		AgentName:        "event_flow_member",
		Role:             agentcore.RoleAssistant,
		DeltaKind:        events.DeltaToolArgs,
		Content:          `{"request":`,
		Delta:            `{"request":`,
		ToolCallID:       "tool-call-1",
		ToolCallRef:      "ref-tool-call-1",
		ToolName:         toolName,
		ToolCallIndex:    &toolIndex,
		IsMemberEvent:    true,
		MemberCallID:     "member-call-1",
		MemberToolName:   "ask_fkagent_parent_member",
		MemberName:       "Event Flow Member",
		MemberOrder:      &toolIndex,
		ParentToolCallID: "member-call-1",
		ParentToolName:   "ask_fkagent_parent_member",
		ToolCalls: []agentcore.ToolCall{{
			ID:    "tool-call-1",
			Index: &toolIndex,
			Type:  "function",
			Function: agentcore.FunctionCall{
				Name:      toolName,
				Arguments: `{"request":"hello"}`,
			},
		}},
		ToolCallRefs:    map[int]string{0: "ref-tool-call-1"},
		ToolCallSpanIDs: map[int]string{0: "span-tool-call-1"},
	})

	got := convertEventToMap(event)
	requireMapValue(t, got, "type", events.EventMessageDelta)
	requireMapValue(t, got, "sequence", int64(42))
	requireMapValue(t, got, "message_id", "msg_member_1")
	requireMapValue(t, got, "stream_id", "msg_member_1:tool_args")
	requireMapValue(t, got, "chunk_index", int64(42))
	requireMapValue(t, got, "delta_kind", events.DeltaToolArgs)
	requireMapValue(t, got, "delta", `{"request":`)
	requireMapValue(t, got, "content", `{"request":`)
	requireMapValue(t, got, "tool_call_id", "tool-call-1")
	requireMapValue(t, got, "tool_call_ref", "ref-tool-call-1")
	requireMapValue(t, got, "tool_name", toolName)
	requireMapValue(t, got, "tool_display_name", "指派给 Event Flow Member")
	requireMapValue(t, got, "tool_kind", toolmeta.ToolKindAgent)
	requireMapValue(t, got, "tool_target", "Event Flow Member")
	requireMapValue(t, got, "tool_call_index", 0)
	requireMapValue(t, got, "is_member_event", true)
	requireMapValue(t, got, "member_call_id", "member-call-1")
	requireMapValue(t, got, "member_tool_name", "ask_fkagent_parent_member")
	requireMapValue(t, got, "member_name", "Event Flow Member")
	requireMapValue(t, got, "member_order", 0)
	requireMapValue(t, got, "parent_tool_call_id", "member-call-1")
	requireMapValue(t, got, "parent_tool_name", "ask_fkagent_parent_member")

	toolCalls, ok := got["tool_calls"].([]map[string]any)
	if !ok || len(toolCalls) != 1 {
		t.Fatalf("expected one tool call map, got %#v", got["tool_calls"])
	}
	requireMapValue(t, toolCalls[0], "id", "tool-call-1")
	requireMapValue(t, toolCalls[0], "index", 0)
	requireMapValue(t, toolCalls[0], "ref", "ref-tool-call-1")
	requireMapValue(t, toolCalls[0], "span_id", "span-tool-call-1")
	requireMapValue(t, toolCalls[0], "name", toolName)
	requireMapValue(t, toolCalls[0], "display_name", "指派给 Event Flow Member")
	requireMapValue(t, toolCalls[0], "kind", toolmeta.ToolKindAgent)
	requireMapValue(t, toolCalls[0], "target", "Event Flow Member")
	requireMapValue(t, toolCalls[0], "arguments", `{"request":"hello"}`)
}

func TestConvertEventToMapOmitsStreamMetadataForNonDeltaEvents(t *testing.T) {
	got := convertEventToMap(events.Event{
		Type:      events.EventMessageEnd,
		Sequence:  7,
		MessageID: "msg_1",
		Role:      agentcore.RoleAssistant,
		Content:   "done",
	})

	if _, ok := got["stream_id"]; ok {
		t.Fatalf("stream_id should be omitted for non-delta event: %#v", got)
	}
	if _, ok := got["chunk_index"]; ok {
		t.Fatalf("chunk_index should be omitted for non-delta event: %#v", got)
	}
}

func requireMapValue(t *testing.T, got map[string]any, key string, want any) {
	t.Helper()
	if got[key] != want {
		t.Fatalf("unexpected %s: got %#v, want %#v; map=%#v", key, got[key], want, got)
	}
}
