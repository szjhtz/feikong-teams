package eino

import (
	"context"
	"testing"

	"fkteams/agentcore"
	"fkteams/internal/testmodel"

	"github.com/cloudwego/eino/schema"
)

func TestAdaptNativeChatModelForRunner(t *testing.T) {
	ctx := context.Background()
	cm := testmodel.New(testmodel.AssistantMessage("ok"))

	runnerModel, err := AdaptChatModelForRunner(cm)
	if err != nil {
		t.Fatalf("adapt model: %v", err)
	}
	bound, err := runnerModel.WithTools([]*schema.ToolInfo{{Name: "test_tool", Desc: "test tool"}})
	if err != nil {
		t.Fatalf("bind tools: %v", err)
	}

	resp, err := bound.Generate(ctx, []*schema.Message{schema.UserMessage("hello")})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if resp.Content != "ok" {
		t.Fatalf("unexpected response: %q", resp.Content)
	}

	calls := cm.GenerateCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 generate call, got %d", len(calls))
	}
	if calls[0].Input[0].Role != agentcore.RoleUser || calls[0].Input[0].Content != "hello" {
		t.Fatalf("unexpected core input: %#v", calls[0].Input)
	}
	if len(calls[0].Tools) != 1 || calls[0].Tools[0].Name != "test_tool" {
		t.Fatalf("expected core tool binding, got %#v", calls[0].Tools)
	}
}
