package fkevent

import (
	"context"
	"sync"
	"sync/atomic"
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

func TestToolCallStateResolvesBeforeTTL(t *testing.T) {
	clearToolCallStateForTest()
	oldTTL := toolCallStateTTL
	toolCallStateTTL = time.Hour
	defer func() {
		toolCallStateTTL = oldTTL
		clearToolCallStateForTest()
	}()

	registerToolCallRef("call_1", "ref_1")

	event := NormalizeEvent(Event{Type: EventToolResult, ToolCallID: "call_1"})
	if event.ToolCallRef != "ref_1" {
		t.Fatalf("expected ref_1, got %q", event.ToolCallRef)
	}
}

func TestToolCallStateExpires(t *testing.T) {
	clearToolCallStateForTest()
	oldTTL := toolCallStateTTL
	toolCallStateTTL = time.Hour
	defer func() {
		toolCallStateTTL = oldTTL
		clearToolCallStateForTest()
	}()

	toolCallRefsByID.Store("call_2", toolCallStateEntry{
		Value:     "ref_2",
		CreatedAt: time.Now().Add(-2 * time.Hour),
	})

	event := NormalizeEvent(Event{Type: EventToolResult, ToolCallID: "call_2"})
	if event.ToolCallRef != "" {
		t.Fatalf("expected expired ref to be ignored, got %q", event.ToolCallRef)
	}
	if _, ok := toolCallRefsByID.Load("call_2"); ok {
		t.Fatal("expected expired ref to be deleted on read")
	}
}

func clearToolCallStateForTest() {
	for _, store := range []*sync.Map{
		&toolCallRefsByID,
		&toolCallOrdersByID,
		&toolCallSpansByID,
		&toolCallSpansByRef,
	} {
		store.Range(func(key, _ any) bool {
			store.Delete(key)
			return true
		})
	}
	atomic.StoreInt64(&toolCallStateStoreCount, 0)
}
