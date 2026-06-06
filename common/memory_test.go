package common

import (
	"context"
	"sync"
	"testing"
)

func TestInMemoryStoreCopiesValues(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	value := []byte("initial")
	if err := store.Set(ctx, "k", value); err != nil {
		t.Fatalf("set value: %v", err)
	}
	value[0] = 'x'

	got, ok, err := store.Get(ctx, "k")
	if err != nil {
		t.Fatalf("get value: %v", err)
	}
	if !ok {
		t.Fatal("expected stored value")
	}
	if string(got) != "initial" {
		t.Fatalf("store should copy input value, got %q", got)
	}

	got[0] = 'y'
	gotAgain, ok, err := store.Get(ctx, "k")
	if err != nil {
		t.Fatalf("get value again: %v", err)
	}
	if !ok {
		t.Fatal("expected stored value on second get")
	}
	if string(gotAgain) != "initial" {
		t.Fatalf("store should copy output value, got %q", gotAgain)
	}
}

func TestInMemoryStoreConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				if err := store.Set(ctx, "k", []byte("value")); err != nil {
					t.Errorf("set value: %v", err)
				}
				if _, _, err := store.Get(ctx, "k"); err != nil {
					t.Errorf("get value: %v", err)
				}
			}
		}()
	}
	wg.Wait()
}
