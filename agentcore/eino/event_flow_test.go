package eino

import (
	"context"
	"strings"
	"testing"

	"fkteams/agentcore"
	"fkteams/common"
	"fkteams/internal/testmodel"
)

func TestStreamingRunEmitsOrderedToolFlowEvents(t *testing.T) {
	ctx := context.Background()
	toolCallIndex := 0
	echoTool, err := agentcore.InferTool("flow_echo", "echo text", func(_ context.Context, req *flowEchoRequest) (*flowEchoResponse, error) {
		return &flowEchoResponse{Text: "echo:" + req.Text}, nil
	})
	if err != nil {
		t.Fatalf("create tool: %v", err)
	}

	model := testmodel.New().
		EnqueueStream(
			agentcore.Message{Role: agentcore.RoleAssistant, ReasoningContent: "think "},
			agentcore.Message{Role: agentcore.RoleAssistant, Content: "draft "},
			agentcore.Message{Role: agentcore.RoleAssistant, ToolCalls: []agentcore.ToolCall{{
				ID:    "flow-tool-call",
				Index: &toolCallIndex,
				Type:  "function",
				Function: agentcore.FunctionCall{
					Name:      "flow_echo",
					Arguments: `{"text":`,
				},
			}}},
			agentcore.Message{Role: agentcore.RoleAssistant, ToolCalls: []agentcore.ToolCall{{
				Index: &toolCallIndex,
				Type:  "function",
				Function: agentcore.FunctionCall{
					Arguments: `"hello"}`,
				},
			}}},
		).
		EnqueueStream(testmodel.AssistantMessage("final"))
	agent, err := NewChatModelAgent(ctx, &agentcore.ChatAgentConfig{
		Name:               "flow",
		Description:        "flow",
		Model:              model,
		Tools:              []agentcore.Tool{echoTool},
		MaxIterations:      4,
		EmitInternalEvents: true,
	})
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}

	events := runAgentForTest(t, ctx, agent, true)
	reasoningIdx := requireEventIndex(t, events, func(event agentcore.Event) bool {
		return event.Type == agentcore.EventMessageDelta &&
			event.Role == agentcore.RoleAssistant &&
			event.DeltaKind == agentcore.DeltaReasoning &&
			event.Content == "think "
	}, "reasoning delta")
	outputIdx := requireEventIndex(t, events, func(event agentcore.Event) bool {
		return event.Type == agentcore.EventMessageDelta &&
			event.Role == agentcore.RoleAssistant &&
			event.DeltaKind == agentcore.DeltaOutput &&
			event.Content == "draft "
	}, "output delta")
	firstArgsIdx := requireEventIndex(t, events, func(event agentcore.Event) bool {
		return event.Type == agentcore.EventMessageDelta &&
			event.DeltaKind == agentcore.DeltaToolArgs &&
			event.ToolCallID == "flow-tool-call" &&
			event.ToolName == "flow_echo" &&
			event.Content == `{"text":`
	}, "first tool args delta")
	secondArgsIdx := requireEventIndex(t, events, func(event agentcore.Event) bool {
		return event.Type == agentcore.EventMessageDelta &&
			event.DeltaKind == agentcore.DeltaToolArgs &&
			event.ToolCallID == "flow-tool-call" &&
			event.ToolName == "flow_echo" &&
			event.Content == `"hello"}`
	}, "second tool args delta")
	messageEndIdx := requireEventIndex(t, events, func(event agentcore.Event) bool {
		return event.Type == agentcore.EventMessageEnd &&
			event.Role == agentcore.RoleAssistant &&
			event.Content == "draft " &&
			event.ReasoningContent == "think " &&
			len(event.ToolCalls) == 1 &&
			event.ToolCalls[0].ID == "flow-tool-call" &&
			event.ToolCalls[0].Function.Name == "flow_echo" &&
			event.ToolCalls[0].Function.Arguments == `{"text":"hello"}`
	}, "assistant message end with aggregated tool call")
	toolStartIdx := requireEventIndex(t, events, func(event agentcore.Event) bool {
		return event.Type == agentcore.EventToolStart &&
			event.ToolCallID == "flow-tool-call" &&
			event.ToolCallRef != "" &&
			event.ToolName == "flow_echo" &&
			event.ToolArgs == `{"text":"hello"}` &&
			event.ToolCallIndex != nil &&
			*event.ToolCallIndex == 0
	}, "tool start")
	toolStart := events[toolStartIdx]
	toolEndIdx := requireEventIndex(t, events, func(event agentcore.Event) bool {
		return event.Type == agentcore.EventToolEnd &&
			event.ToolCallID == "flow-tool-call" &&
			event.ToolCallRef == toolStart.ToolCallRef &&
			event.ToolName == "flow_echo" &&
			strings.Contains(event.ToolResult, "echo:hello")
	}, "tool end")
	toolMessageIdx := requireEventIndex(t, events, func(event agentcore.Event) bool {
		return event.Type == agentcore.EventMessageEnd &&
			event.Role == agentcore.RoleTool &&
			event.ToolCallID == "flow-tool-call" &&
			event.ToolName == "flow_echo" &&
			strings.Contains(event.Content, "echo:hello")
	}, "tool message end")
	finalIdx := requireEventIndex(t, events, func(event agentcore.Event) bool {
		return event.Type == agentcore.EventMessageDelta &&
			event.Role == agentcore.RoleAssistant &&
			event.DeltaKind == agentcore.DeltaOutput &&
			event.Content == "final"
	}, "final assistant delta")

	requireBefore(t, events, reasoningIdx, outputIdx, "reasoning", "output")
	requireBefore(t, events, outputIdx, firstArgsIdx, "output", "first tool args")
	requireBefore(t, events, firstArgsIdx, secondArgsIdx, "first tool args", "second tool args")
	requireBefore(t, events, secondArgsIdx, messageEndIdx, "second tool args", "message end")
	requireBefore(t, events, messageEndIdx, toolStartIdx, "message end", "tool start")
	requireBefore(t, events, toolStartIdx, toolEndIdx, "tool start", "tool end")
	requireBefore(t, events, toolEndIdx, toolMessageIdx, "tool end", "tool message")
	requireBefore(t, events, toolMessageIdx, finalIdx, "tool message", "final output")
}

func TestGenerateRunEmitsRegularMessageAndToolEvents(t *testing.T) {
	ctx := context.Background()
	toolCallIndex := 0
	echoTool, err := agentcore.InferTool("generate_echo", "echo text", func(_ context.Context, req *flowEchoRequest) (*flowEchoResponse, error) {
		return &flowEchoResponse{Text: "echo:" + req.Text}, nil
	})
	if err != nil {
		t.Fatalf("create tool: %v", err)
	}

	model := testmodel.New(
		agentcore.Message{
			Role:             agentcore.RoleAssistant,
			Content:          "regular-draft",
			ReasoningContent: "regular-thinking",
			ToolCalls: []agentcore.ToolCall{{
				ID:    "generate-tool-call",
				Index: &toolCallIndex,
				Type:  "function",
				Function: agentcore.FunctionCall{
					Name:      "generate_echo",
					Arguments: `{"text":"hello"}`,
				},
			}},
		},
		testmodel.AssistantMessage("regular-final"),
	)
	agent, err := NewChatModelAgent(ctx, &agentcore.ChatAgentConfig{
		Name:               "generate-flow",
		Description:        "generate flow",
		Model:              model,
		Tools:              []agentcore.Tool{echoTool},
		MaxIterations:      4,
		EmitInternalEvents: true,
	})
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}

	events := runAgentForTest(t, ctx, agent, false)
	reasoningIdx := requireEventIndex(t, events, func(event agentcore.Event) bool {
		return event.Type == agentcore.EventMessageDelta &&
			event.Role == agentcore.RoleAssistant &&
			event.DeltaKind == agentcore.DeltaReasoning &&
			event.Content == "regular-thinking"
	}, "regular reasoning delta")
	outputIdx := requireEventIndex(t, events, func(event agentcore.Event) bool {
		return event.Type == agentcore.EventMessageDelta &&
			event.Role == agentcore.RoleAssistant &&
			event.DeltaKind == agentcore.DeltaOutput &&
			event.Content == "regular-draft"
	}, "regular output delta")
	messageEndIdx := requireEventIndex(t, events, func(event agentcore.Event) bool {
		return event.Type == agentcore.EventMessageEnd &&
			event.Role == agentcore.RoleAssistant &&
			event.Content == "regular-draft" &&
			event.ReasoningContent == "regular-thinking" &&
			len(event.ToolCalls) == 1 &&
			event.ToolCalls[0].ID == "generate-tool-call" &&
			event.ToolCalls[0].Function.Name == "generate_echo" &&
			event.ToolCalls[0].Function.Arguments == `{"text":"hello"}`
	}, "regular message end with tool call")
	toolStartIdx := requireEventIndex(t, events, func(event agentcore.Event) bool {
		return event.Type == agentcore.EventToolStart &&
			event.ToolCallID == "generate-tool-call" &&
			event.ToolName == "generate_echo" &&
			event.ToolCallRef != "" &&
			event.ToolArgs == `{"text":"hello"}`
	}, "regular tool start")
	toolStart := events[toolStartIdx]
	toolEndIdx := requireEventIndex(t, events, func(event agentcore.Event) bool {
		return event.Type == agentcore.EventToolEnd &&
			event.ToolCallID == "generate-tool-call" &&
			event.ToolCallRef == toolStart.ToolCallRef &&
			event.ToolName == "generate_echo" &&
			strings.Contains(event.ToolResult, "echo:hello")
	}, "regular tool end")
	finalIdx := requireEventIndex(t, events, func(event agentcore.Event) bool {
		return event.Type == agentcore.EventMessageDelta &&
			event.Role == agentcore.RoleAssistant &&
			event.DeltaKind == agentcore.DeltaOutput &&
			event.Content == "regular-final"
	}, "regular final delta")

	requireBefore(t, events, reasoningIdx, outputIdx, "regular reasoning", "regular output")
	requireBefore(t, events, outputIdx, messageEndIdx, "regular output", "regular message end")
	requireBefore(t, events, messageEndIdx, toolStartIdx, "regular message end", "regular tool start")
	requireBefore(t, events, toolStartIdx, toolEndIdx, "regular tool start", "regular tool end")
	requireBefore(t, events, toolEndIdx, finalIdx, "regular tool end", "regular final output")
}

func runAgentForTest(t *testing.T, ctx context.Context, agent agentcore.Agent, streaming bool) []agentcore.Event {
	t.Helper()

	runner, err := NewRunnerFromConfig(ctx, agentcore.RunnerConfig{
		Agent:           agent,
		EnableStreaming: streaming,
		CheckPointStore: common.NewInMemoryStore(),
	})
	if err != nil {
		t.Fatalf("create runner: %v", err)
	}

	var events []agentcore.Event
	_, err = runner.Run(ctx, agentcore.TurnInput{
		Message: agentcore.Message{Role: agentcore.RoleUser, Content: "start"},
	}, agentcore.RunOptions{
		RunID:        "event-flow-test",
		CheckpointID: "event-flow-test",
		Sink: func(event agentcore.Event) error {
			events = append(events, event)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("run agent: %v", err)
	}
	return events
}

func requireEventIndex(t *testing.T, events []agentcore.Event, match func(agentcore.Event) bool, name string) int {
	t.Helper()
	for i, event := range events {
		if match(event) {
			return i
		}
	}
	t.Fatalf("missing %s event; events=%#v", name, events)
	return -1
}

func requireBefore(t *testing.T, events []agentcore.Event, before, after int, beforeName, afterName string) {
	t.Helper()
	if before >= after {
		t.Fatalf("expected %s before %s; before=%d after=%d events=%#v", beforeName, afterName, before, after, events)
	}
}

type flowEchoRequest struct {
	Text string `json:"text"`
}

type flowEchoResponse struct {
	Text string `json:"text"`
}
