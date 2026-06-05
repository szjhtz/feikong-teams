package fkevent

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
		firstDone <- DispatchEvent(slowCtx, Event{Type: EventMessage, Content: "slow"})
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("slow callback did not start")
	}

	go func() {
		fastDone <- DispatchEvent(fastCtx, Event{Type: EventMessage, Content: "fast"})
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
