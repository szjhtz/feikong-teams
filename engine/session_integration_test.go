package engine

import (
	"context"
	"fkteams/agentcore"
	"fkteams/fkevent"
	"testing"
)

type recordingRunner struct {
	input agentcore.TurnInput
}

func (r *recordingRunner) Run(_ context.Context, input agentcore.TurnInput, opts agentcore.RunOptions) (*agentcore.RunResult, error) {
	r.input = input
	event := agentcore.Event{
		Type:      agentcore.EventMessageDelta,
		Role:      agentcore.RoleAssistant,
		DeltaKind: agentcore.DeltaOutput,
		Content:   "pong",
		Delta:     "pong",
	}
	if opts.Sink != nil {
		if err := opts.Sink(event); err != nil {
			return nil, err
		}
	}
	return &agentcore.RunResult{LastEvent: event}, nil
}

func TestSessionRunsCoreRunner(t *testing.T) {
	ctx := context.Background()
	r := &recordingRunner{}

	var events []fkevent.Event
	_, err := NewSession(r, "test-session").
		WithInput(TurnInput{Message: agentcore.Message{Role: agentcore.RoleUser, Content: "ping"}}).
		OnEvent(func(event fkevent.Event) error {
			events = append(events, event)
			return nil
		}).
		Run(ctx)
	if err != nil {
		t.Fatalf("run session: %v", err)
	}

	messages := r.input.AllMessages()
	if len(messages) == 0 || messages[len(messages)-1].Content != "ping" {
		t.Fatalf("expected user input to reach runner, got %#v", messages)
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
