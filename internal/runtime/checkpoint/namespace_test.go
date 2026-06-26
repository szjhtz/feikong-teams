package checkpoint

import (
	"context"
	"testing"
)

func TestNamespaceStoreSeparatesKeys(t *testing.T) {
	ctx := context.Background()
	inner := NewMemoryStore()
	first := NewNamespaceStore("first", inner)
	second := NewNamespaceStore("second", inner)

	if err := first.Set(ctx, "k", []byte("one")); err != nil {
		t.Fatalf("set first: %v", err)
	}
	if err := second.Set(ctx, "k", []byte("two")); err != nil {
		t.Fatalf("set second: %v", err)
	}

	got, ok, err := first.Get(ctx, "k")
	if err != nil {
		t.Fatalf("get first: %v", err)
	}
	if !ok || string(got) != "one" {
		t.Fatalf("first value = %q, %v; want one, true", got, ok)
	}

	got, ok, err = second.Get(ctx, "k")
	if err != nil {
		t.Fatalf("get second: %v", err)
	}
	if !ok || string(got) != "two" {
		t.Fatalf("second value = %q, %v; want two, true", got, ok)
	}
}
