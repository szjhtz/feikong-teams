package turn

import (
	"context"
	domainevent "fkteams/internal/domain/event"
	"fkteams/internal/domain/message"
	runtimeport "fkteams/internal/ports/runtime"
	"fkteams/internal/runtime/events"
	"fkteams/internal/runtime/hooks"
	"testing"
)

type recordingRunner struct {
	input message.TurnInput
}

func (r *recordingRunner) Run(_ context.Context, input message.TurnInput, opts runtimeport.RunOptions) (*runtimeport.RunResult, error) {
	r.input = input
	event := domainevent.Event{
		Type:      domainevent.TypeAssistantText,
		Role:      message.RoleAssistant,
		DeltaKind: domainevent.DeltaOutput,
		Content:   "pong",
	}
	if opts.Sink != nil {
		if err := opts.Sink(event); err != nil {
			return nil, err
		}
	}
	return &runtimeport.RunResult{LastEvent: event}, nil
}

func TestExecutorRunsCoreRunner(t *testing.T) {
	ctx := context.Background()
	r := &recordingRunner{}

	var collectedEvents []events.Event
	_, err := NewExecutor().Run(ctx, Request{
		Runner:    r,
		SessionID: "test-session",
		Input:     TurnInput{Message: message.Message{Role: message.RoleUser, Content: "ping"}},
		EventSink: func(event events.Event) error {
			collectedEvents = append(collectedEvents, event)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("run session: %v", err)
	}

	messages := r.input.AllMessages()
	if len(messages) == 0 || messages[len(messages)-1].Content != "ping" {
		t.Fatalf("expected user input to reach runner, got %#v", messages)
	}

	found := false
	for _, event := range collectedEvents {
		if event.Content == "pong" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected pong event, got %#v", collectedEvents)
	}
}

func TestExecutorInvokesRunHooks(t *testing.T) {
	ctx := context.Background()
	r := &recordingRunner{}
	bus := hooks.NewBus()
	afterCalled := false

	bus.RegisterFunc("rewrite-input", []hooks.HookPoint{hooks.HookBeforeRun}, func(ctx hooks.Context, inv hooks.Invocation) (hooks.Result, error) {
		payload := inv.Payload.(hooks.BeforeRunPayload)
		payload.Input.Message.Content = "hooked"
		return hooks.Result{Payload: payload}, nil
	}, hooks.Options{})
	bus.RegisterFunc("after-run", []hooks.HookPoint{hooks.HookAfterRun}, func(ctx hooks.Context, inv hooks.Invocation) (hooks.Result, error) {
		payload := inv.Payload.(hooks.AfterRunPayload)
		if payload.Input.Message.Content != "hooked" {
			t.Fatalf("after input = %q, want hooked", payload.Input.Message.Content)
		}
		if payload.Result == nil {
			t.Fatal("after hook result is nil")
		}
		afterCalled = true
		return hooks.Result{}, nil
	}, hooks.Options{})

	_, err := NewExecutor().Run(ctx, Request{
		Runner:    r,
		SessionID: "test-session",
		HookBus:   bus,
		Input:     TurnInput{Message: message.Message{Role: message.RoleUser, Content: "ping"}},
	})
	if err != nil {
		t.Fatalf("run session: %v", err)
	}
	if r.input.Message.Content != "hooked" {
		t.Fatalf("runner input = %q, want hooked", r.input.Message.Content)
	}
	if !afterCalled {
		t.Fatal("after hook was not called")
	}
}

func TestExecutorHasNoImplicitGlobalHookBus(t *testing.T) {
	ctx := context.Background()
	r := &recordingRunner{}

	_, err := NewExecutor().Run(ctx, Request{
		Runner:    r,
		SessionID: "test-session",
		Input:     TurnInput{Message: message.Message{Role: message.RoleUser, Content: "ping"}},
	})
	if err != nil {
		t.Fatalf("run session: %v", err)
	}
	if r.input.Message.Content != "ping" {
		t.Fatalf("runner input = %q, want unchanged", r.input.Message.Content)
	}
}
