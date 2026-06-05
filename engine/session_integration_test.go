package engine

import (
	"context"
	"fkteams/common"
	"fkteams/fkevent"
	"fkteams/internal/testmodel"
	"testing"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

func TestSessionRunsChatModelAgentWithTestModel(t *testing.T) {
	ctx := context.Background()
	cm := testmodel.New(schema.AssistantMessage("pong", nil))

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "test_agent",
		Description: "test agent",
		Model:       cm,
	})
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}

	r := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: false,
		CheckPointStore: common.NewInMemoryStore(),
	})

	var events []fkevent.Event
	_, err = NewSession(r, "test-session").
		WithMessages([]adk.Message{schema.UserMessage("ping")}).
		OnEvent(func(event fkevent.Event) error {
			events = append(events, event)
			return nil
		}).
		Run(ctx)
	if err != nil {
		t.Fatalf("run session: %v", err)
	}

	calls := cm.GenerateCalls()
	if len(calls) != 1 {
		t.Fatalf("expected one model call, got %d", len(calls))
	}
	if len(calls[0].Input) == 0 || calls[0].Input[len(calls[0].Input)-1].Content != "ping" {
		t.Fatalf("expected user input to reach model, got %#v", calls[0].Input)
	}

	found := false
	for _, event := range events {
		if event.Content == "pong" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected pong event, got %#v", events)
	}
}
