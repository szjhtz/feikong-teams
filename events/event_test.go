package events

import (
	"context"
	"testing"
	"time"
)

func TestDispatchEventDoesNotSerializeCallbacksGlobally(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	firstDone := make(chan error, 1)

	slowCtx := WithCallback(context.Background(), func(Event) error {
		close(started)
		<-release
		return nil
	})
	fastDone := make(chan error, 1)
	fastCtx := WithCallback(context.Background(), func(Event) error {
		return nil
	})

	go func() {
		firstDone <- DispatchEvent(slowCtx, Event{Type: EventMessageDelta, Content: "slow"})
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("slow callback did not start")
	}

	go func() {
		fastDone <- DispatchEvent(fastCtx, Event{Type: EventMessageDelta, Content: "fast"})
	}()

	select {
	case err := <-fastDone:
		if err != nil {
			t.Fatalf("fast callback returned error: %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("fast callback was blocked by unrelated slow callback")
	}

	close(release)
	if err := <-firstDone; err != nil {
		t.Fatalf("slow callback returned error: %v", err)
	}
}

func TestNormalizeEventFillsCommonMetadata(t *testing.T) {
	event := NormalizeEvent(Event{Type: EventMessageDelta, RunID: "run_1", Content: "hello"})
	if event.EventID == "" {
		t.Fatal("event id was not set")
	}
	if event.Sequence == 0 {
		t.Fatal("sequence was not set")
	}
	if event.CreatedAt.IsZero() {
		t.Fatal("created_at was not set")
	}
	if event.RunID != "run_1" {
		t.Fatalf("run id = %q, want run_1", event.RunID)
	}
}

func TestIsMemberEventUsesMemberCallID(t *testing.T) {
	event := NormalizeEvent(Event{Type: EventMessageDelta, MemberCallID: "call_1"})
	if !IsMemberEvent(event) {
		t.Fatal("member event was not detected")
	}
}
