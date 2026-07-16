package handler

import (
	"testing"
	"time"
)

func TestAttemptLimiterWindowAndReset(t *testing.T) {
	limiter := newAttemptLimiter(2, time.Minute, 10)
	now := time.Unix(1_800_000_000, 0)

	if allowed, _ := limiter.Allow("client", now); !allowed {
		t.Fatal("first attempt should be allowed")
	}
	if allowed, _ := limiter.Allow("client", now.Add(time.Second)); !allowed {
		t.Fatal("second attempt should be allowed")
	}
	if allowed, retryAfter := limiter.Allow("client", now.Add(2*time.Second)); allowed || retryAfter != 58*time.Second {
		t.Fatalf("third attempt = (%v, %v), want (false, 58s)", allowed, retryAfter)
	}

	limiter.Reset("client")
	if allowed, _ := limiter.Allow("client", now.Add(3*time.Second)); !allowed {
		t.Fatal("reset should allow a new attempt")
	}
	if allowed, _ := limiter.Allow("client", now.Add(2*time.Minute)); !allowed {
		t.Fatal("expired window should allow a new attempt")
	}
}

func TestAttemptLimiterBoundsStoredClients(t *testing.T) {
	limiter := newAttemptLimiter(1, time.Minute, 2)
	now := time.Unix(1_800_000_000, 0)
	limiter.Allow("oldest", now)
	limiter.Allow("newer", now.Add(time.Second))
	limiter.Allow("newest", now.Add(2*time.Second))

	if len(limiter.entries) != 2 {
		t.Fatalf("stored entries = %d, want 2", len(limiter.entries))
	}
	if _, exists := limiter.entries["oldest"]; exists {
		t.Fatal("oldest client should be evicted")
	}
}
