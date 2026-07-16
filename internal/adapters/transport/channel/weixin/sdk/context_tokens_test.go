package wechatbot

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestContextTokenCacheIsBoundedAndUsesLRU(t *testing.T) {
	cache := newContextTokenCache(3, time.Hour)
	cache.Set("user-1", "token-1")
	cache.Set("user-2", "token-2")
	cache.Set("user-3", "token-3")
	if _, ok := cache.Get("user-1"); !ok {
		t.Fatal("expected cached token")
	}
	cache.Set("user-4", "token-4")

	if _, ok := cache.Get("user-2"); ok {
		t.Fatal("least recently used token should be evicted")
	}
	for _, id := range []string{"user-1", "user-3", "user-4"} {
		if token, ok := cache.Get(id); !ok || token == "" {
			t.Fatalf("expected token for %s", id)
		}
	}
	cache.mu.Lock()
	count := len(cache.entries)
	cache.mu.Unlock()
	if count != 3 {
		t.Fatalf("context token count = %d, want 3", count)
	}
}

func TestContextTokenCacheResetAndConcurrentAccess(t *testing.T) {
	cache := newContextTokenCache(32, time.Hour)
	done := make(chan struct{})
	for worker := 0; worker < 8; worker++ {
		go func(worker int) {
			defer func() { done <- struct{}{} }()
			for i := 0; i < 100; i++ {
				userID := fmt.Sprintf("user-%d", (worker+i)%64)
				cache.Set(userID, "token")
				cache.Get(userID)
				if i%25 == 0 {
					cache.Reset()
				}
			}
		}(worker)
	}
	for worker := 0; worker < 8; worker++ {
		<-done
	}
	cache.mu.Lock()
	count := len(cache.entries)
	cache.mu.Unlock()
	if count > 32 {
		t.Fatalf("context token count = %d, want <= 32", count)
	}
}

func TestWaitForRetryCanBeCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	started := time.Now()
	if waitForRetry(ctx, time.Minute) {
		t.Fatal("cancelled retry wait should return false")
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("cancelled retry wait took %s", elapsed)
	}
}

func TestReplyRejectsNilMessage(t *testing.T) {
	bot := New()
	if err := bot.Reply(context.Background(), nil, "hello"); err == nil {
		t.Fatal("expected nil message to be rejected")
	}
	if err := bot.ReplyContent(context.Background(), nil, SendText("hello")); err == nil {
		t.Fatal("expected nil message content to be rejected")
	}
}
