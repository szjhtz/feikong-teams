package qq

import (
	"container/list"
	"fmt"
	"testing"
	"time"
)

func TestMessageDeduplicationIsBounded(t *testing.T) {
	channel := newTestChannel()
	if channel.isDuplicate("message-1") {
		t.Fatal("first message must not be treated as duplicate")
	}
	if !channel.isDuplicate("message-1") {
		t.Fatal("second message must be treated as duplicate")
	}
	for i := 2; i <= maxSeenMessages+1; i++ {
		if channel.isDuplicate(fmt.Sprintf("message-%d", i)) {
			t.Fatalf("new message %d was treated as duplicate", i)
		}
	}
	channel.seenMu.Lock()
	count := len(channel.seen)
	_, oldestExists := channel.seen["message-1"]
	channel.seenMu.Unlock()
	if count != maxSeenMessages {
		t.Fatalf("seen message count = %d, want %d", count, maxSeenMessages)
	}
	if oldestExists {
		t.Fatal("oldest seen message should be evicted at capacity")
	}
}

func TestChatStateCacheIsBoundedAndExpiresIdleEntries(t *testing.T) {
	channel := newTestChannel()
	for i := 0; i <= maxChatStates; i++ {
		channel.getState(fmt.Sprintf("chat-%d", i))
	}
	channel.statesMu.Lock()
	count := len(channel.states)
	_, oldestExists := channel.states["chat-0"]
	newest := channel.states[fmt.Sprintf("chat-%d", maxChatStates)]
	channel.statesMu.Unlock()
	if count != maxChatStates {
		t.Fatalf("chat state count = %d, want %d", count, maxChatStates)
	}
	if oldestExists || newest == nil {
		t.Fatalf("unexpected state eviction: oldest=%v newest=%v", oldestExists, newest != nil)
	}

	newest.mu.Lock()
	newest.lastSeen = time.Now().Add(-chatStateTTL - time.Second)
	newest.mu.Unlock()
	channel.statesMu.Lock()
	channel.stateOrder.MoveToBack(newest.order)
	channel.pruneStatesLocked(time.Now())
	_, idleExists := channel.states[fmt.Sprintf("chat-%d", maxChatStates)]
	channel.statesMu.Unlock()
	if idleExists {
		t.Fatal("idle chat state should be removed")
	}
}

func newTestChannel() *Channel {
	return &Channel{
		seen:       make(map[string]time.Time),
		seenOrder:  list.New(),
		states:     make(map[string]*chatState),
		stateOrder: list.New(),
	}
}
