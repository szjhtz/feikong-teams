package engine

import (
	"context"
	"fkteams/fkevent"
	"fkteams/tools/approval"
	"testing"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

type historySinkStub struct {
	count int
}

func (h *historySinkStub) GetMessageCount() int {
	return h.count
}

func (h *historySinkStub) RecordUserInput(string) {}

func (h *historySinkStub) SetSummary(string, int) {}

func TestSessionBuilderConfiguresRunConfig(t *testing.T) {
	messages := []adk.Message{schema.UserMessage("hello")}
	history := &historySinkStub{}
	approvalReg := approval.NewDefaultRegistry()
	eventHandler := func(fkevent.Event) error { return nil }
	startHandler := func(context.Context) {}
	interruptHandler := AutoRejectHandler()
	finishHandler := func(context.Context, *adk.AgentEvent, error) {}

	session := NewSession(&adk.Runner{}, "session-1").
		WithMessages(messages).
		OnEvent(eventHandler).
		WithHistory(history).
		OnStart(startHandler).
		OnInterrupt(interruptHandler).
		NonInteractive().
		WithApproval(approvalReg).
		OnFinish(finishHandler)

	if len(session.cfg.Messages) != 1 || session.cfg.Messages[0].Content != "hello" {
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
	if session.cfg.ApprovalReg != approvalReg {
		t.Fatal("approval registry was not configured")
	}
	if session.cfg.OnFinish == nil {
		t.Fatal("finish handler was not configured")
	}
}
