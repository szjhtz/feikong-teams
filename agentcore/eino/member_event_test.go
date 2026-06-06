package eino

import (
	"context"
	"strings"
	"testing"

	"fkteams/agentcore"
	"fkteams/internal/testmodel"
)

func TestAgentToolMemberEventsKeepScopeForReasoningAndTools(t *testing.T) {
	ctx := context.Background()
	memberTool, err := agentcore.InferTool("member_echo", "member echo", func(_ context.Context, req *memberEchoRequest) (*memberEchoResponse, error) {
		return &memberEchoResponse{Text: "tool:" + req.Text}, nil
	})
	if err != nil {
		t.Fatalf("create member tool: %v", err)
	}

	memberModel := testmodel.New().
		EnqueueStream(
			agentcore.Message{Role: agentcore.RoleAssistant, ReasoningContent: "member-thinking"},
			agentcore.Message{Role: agentcore.RoleAssistant, ToolCalls: []agentcore.ToolCall{{
				ID:    "member-tool-call",
				Index: intPtr(0),
				Type:  "function",
				Function: agentcore.FunctionCall{
					Name:      "member_echo",
					Arguments: `{"text":"hello"}`,
				},
			}}},
		).
		EnqueueStream(testmodel.AssistantMessage("member-done"))
	memberAgent, err := NewChatModelAgent(ctx, &agentcore.ChatAgentConfig{
		Name:               "member",
		Description:        "member",
		Model:              memberModel,
		Tools:              []agentcore.Tool{memberTool},
		MaxIterations:      4,
		EmitInternalEvents: true,
	})
	if err != nil {
		t.Fatalf("create member agent: %v", err)
	}

	agentTools, err := NewAgentTools(ctx, []agentcore.Agent{memberAgent}, agentcore.AgentToolConfig{
		ToolName: func(string, int) string { return "ask_fkagent_member" },
	})
	if err != nil {
		t.Fatalf("create agent tools: %v", err)
	}

	parentModel := testmodel.New().
		EnqueueStream(agentcore.Message{Role: agentcore.RoleAssistant, ToolCalls: []agentcore.ToolCall{{
			ID:    "parent-member-call",
			Index: intPtr(0),
			Type:  "function",
			Function: agentcore.FunctionCall{
				Name:      "ask_fkagent_member",
				Arguments: `{"request":"do member task"}`,
			},
		}}}).
		EnqueueStream(testmodel.AssistantMessage("parent-done"))
	parentAgent, err := NewChatModelAgent(ctx, &agentcore.ChatAgentConfig{
		Name:               "parent",
		Description:        "parent",
		Model:              parentModel,
		Tools:              agentTools,
		MaxIterations:      4,
		EmitInternalEvents: true,
	})
	if err != nil {
		t.Fatalf("create parent agent: %v", err)
	}

	got := runAgentForTest(t, ctx, parentAgent, true)

	parentStartIdx := requireEventIndex(t, got, func(event agentcore.Event) bool {
		return event.Type == agentcore.EventToolStart &&
			event.ToolCallID == "parent-member-call" &&
			event.ToolName == "ask_fkagent_member" &&
			event.ToolCallRef != ""
	}, "parent member tool start")
	memberReasoningIdx := requireEventIndex(t, got, func(event agentcore.Event) bool {
		return event.MemberCallID == "parent-member-call" &&
			event.ParentToolCallID == "parent-member-call" &&
			event.MemberToolName == "ask_fkagent_member" &&
			event.MemberName == "member" &&
			event.DeltaKind == agentcore.DeltaReasoning &&
			strings.Contains(event.Content, "member-thinking")
	}, "member-scoped reasoning")
	memberToolStartIdx := requireEventIndex(t, got, func(event agentcore.Event) bool {
		return event.MemberCallID == "parent-member-call" &&
			event.Type == agentcore.EventToolStart &&
			event.ToolName == "member_echo" &&
			event.ToolCallRef != "" &&
			event.ToolCallIndex != nil &&
			*event.ToolCallIndex == 0
	}, "member-scoped tool start")
	memberToolResultIdx := requireEventIndex(t, got, func(event agentcore.Event) bool {
		return event.MemberCallID == "parent-member-call" &&
			(event.Type == agentcore.EventToolUpdate || event.Type == agentcore.EventToolEnd) &&
			event.ToolName == "member_echo" &&
			event.ToolCallRef != ""
	}, "member-scoped tool result")

	requireBefore(t, got, parentStartIdx, memberReasoningIdx, "parent member tool start", "member reasoning")
	requireBefore(t, got, memberReasoningIdx, memberToolStartIdx, "member reasoning", "member tool start")
	requireBefore(t, got, memberToolStartIdx, memberToolResultIdx, "member tool start", "member tool result")
}

type memberEchoRequest struct {
	Text string `json:"text"`
}

type memberEchoResponse struct {
	Text string `json:"text"`
}
