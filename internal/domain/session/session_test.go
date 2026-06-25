package session

import (
	"context"
	"strings"
	"testing"
)

func TestSessionIDContext(t *testing.T) {
	ctx := WithID(context.Background(), "session-1")
	got, ok := IDFromContext(ctx)
	if !ok || got != "session-1" {
		t.Fatalf("IDFromContext = %q/%v, want session-1/true", got, ok)
	}
	if _, ok := IDFromContext(context.Background()); ok {
		t.Fatal("empty context should not have session id")
	}
}

func TestNewIDLooksLikeUUID(t *testing.T) {
	id := NewID()
	if len(id) != 36 || strings.Count(id, "-") != 4 {
		t.Fatalf("session id = %q, want UUID-like string", id)
	}
}
