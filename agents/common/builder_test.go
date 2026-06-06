package common_test

import (
	"context"
	"strings"
	"testing"

	"fkteams/agentcore"
	einoruntime "fkteams/agentcore/eino"
	agentscommon "fkteams/agents/common"
	"fkteams/internal/testmodel"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

func TestAgentBuilderRunsWithInjectedTestModel(t *testing.T) {
	ctx := context.Background()
	cm := testmodel.New(schema.AssistantMessage("builder-ok", nil))

	agent, err := agentscommon.NewAgentBuilder("builder_test", "builder test agent").
		WithModel(agentcore.WrapRuntimeChatModel(cm)).
		WithInstruction("you are a {role}").
		WithTemplateVar("role", "tester").
		Build(ctx)
	if err != nil {
		t.Fatalf("build agent: %v", err)
	}

	runnerAgent, err := einoruntime.AdaptAgentForRunner(agent)
	if err != nil {
		t.Fatalf("adapt agent: %v", err)
	}
	events := drainAgent(t, runnerAgent, schema.UserMessage("ping"))
	if len(events) == 0 {
		t.Fatal("expected at least one event")
	}

	calls := cm.GenerateCalls()
	if len(calls) != 1 {
		t.Fatalf("expected one model call, got %d", len(calls))
	}

	input := calls[0].Input
	if len(input) < 3 {
		t.Fatalf("expected system, user and injected context messages, got %#v", input)
	}
	if input[0].Role != schema.System || !strings.Contains(input[0].Content, "you are a tester") {
		t.Fatalf("expected formatted system prompt, got %#v", input[0])
	}
	if input[len(input)-2].Role != schema.User || input[len(input)-2].Content != "ping" {
		t.Fatalf("expected user message before dynamic context, got %#v", input)
	}
	assertInjectedContext(t, input[len(input)-1])
}

func drainAgent(t *testing.T, agent adk.Agent, messages ...adk.Message) []*adk.AgentEvent {
	t.Helper()

	iter := agent.Run(context.Background(), &adk.AgentInput{Messages: messages})
	var events []*adk.AgentEvent
	for {
		event, ok := iter.Next()
		if !ok {
			return events
		}
		if event.Err != nil {
			t.Fatalf("agent event error: %v", event.Err)
		}
		events = append(events, event)
	}
}

func assertInjectedContext(t *testing.T, msg *schema.Message) {
	t.Helper()

	if msg.Role != schema.User {
		t.Fatalf("expected injected context to be user message, got %s", msg.Role)
	}
	for _, want := range []string{"<system-reminder>", "当前时间", "工作目录"} {
		if !strings.Contains(msg.Content, want) {
			t.Fatalf("expected injected context to contain %q, got %q", want, msg.Content)
		}
	}
}
