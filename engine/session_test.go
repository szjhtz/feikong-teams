package engine

import (
	"context"
	"fkteams/agentcore"
	"fkteams/events"
	"fkteams/tools/approval"
	"testing"
)

type historySinkStub struct {
	count int
}

func (h *historySinkStub) GetMessageCount() int {
	return h.count
}

func (h *historySinkStub) RecordUserMessage(agentcore.Message) {}

func (h *historySinkStub) SetSummary(string, int) {}

type runnerStub struct {
	input agentcore.TurnInput
}

func (r *runnerStub) Run(_ context.Context, input agentcore.TurnInput, _ agentcore.RunOptions) (*agentcore.RunResult, error) {
	r.input = input
	return &agentcore.RunResult{}, nil
}

func TestSessionBuilderConfiguresRunConfig(t *testing.T) {
	messages := []agentcore.Message{{Role: agentcore.RoleUser, Content: "hello"}}
	history := &historySinkStub{}
	approvalReg := approval.NewDefaultRegistry()
	eventHandler := func(events.Event) error { return nil }
	startHandler := func(context.Context) {}
	interruptHandler := FixedDecisionHandler(approval.Reject)
	finishHandler := func(context.Context, *agentcore.RunResult, error) {}

	session := NewSession(&runnerStub{}, "session-1").
		WithMessages(messages).
		OnEvent(eventHandler).
		WithHistory(history).
		OnStart(startHandler).
		OnInterrupt(interruptHandler).
		NonInteractive().
		WithContext(approval.RegistryContext(approvalReg)).
		OnFinish(finishHandler)

	if len(session.cfg.Input.Context) != 1 || session.cfg.Input.Context[0].Content != "hello" {
		t.Fatal("messages were not configured")
	}
	if session.cfg.EventCallback == nil {
		t.Fatal("event handler was not configured")
	}
	if session.cfg.Recorder != history {
		t.Fatal("history sink was not configured")
	}
	if session.cfg.OnStart == nil {
		t.Fatal("start handler was not configured")
	}
	if session.cfg.OnInterrupt == nil {
		t.Fatal("interrupt handler was not configured")
	}
	if !session.cfg.NonInteractive {
		t.Fatal("non-interactive flag was not configured")
	}
	if len(session.cfg.ContextHooks) != 1 {
		t.Fatal("context hook was not configured")
	}
	if session.cfg.OnFinish == nil {
		t.Fatal("finish handler was not configured")
	}
}

func TestSessionBuilderConfiguresTurnInput(t *testing.T) {
	input := TurnInput{
		Context: []agentcore.Message{{Role: agentcore.RoleSystem, Content: "context"}},
		Message: agentcore.Message{Role: agentcore.RoleUser, Content: "hello"},
	}

	session := NewSession(&runnerStub{}, "session-1").WithInput(input)

	if len(session.cfg.Input.Context) != 1 || session.cfg.Input.Context[0].Content != "context" {
		t.Fatal("input context was not configured")
	}
	if session.cfg.Input.Message.Content != "hello" {
		t.Fatalf("input message = %q, want hello", session.cfg.Input.Message.Content)
	}
}
