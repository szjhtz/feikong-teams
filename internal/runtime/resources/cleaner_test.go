package resources

import (
	"errors"
	"strings"
	"testing"
)

func TestCleanerExecutesLIFOAndClears(t *testing.T) {
	cleaner := NewCleaner()
	var order []string
	cleaner.Add(func() error {
		order = append(order, "first")
		return nil
	})
	cleaner.Add(func() error {
		order = append(order, "second")
		return nil
	})

	if err := cleaner.ExecuteAndClear(); err != nil {
		t.Fatalf("ExecuteAndClear: %v", err)
	}
	if len(order) != 2 || order[0] != "second" || order[1] != "first" {
		t.Fatalf("order = %#v, want LIFO", order)
	}

	if err := cleaner.ExecuteAndClear(); err != nil {
		t.Fatalf("second ExecuteAndClear: %v", err)
	}
	if len(order) != 2 {
		t.Fatalf("cleanups should be cleared, order=%#v", order)
	}
}

func TestCleanerReturnsFirstErrorAndRecoversPanic(t *testing.T) {
	cleaner := NewCleaner()
	cleaner.Add(func() error {
		return errors.New("first")
	})
	cleaner.Add(func() error {
		panic("boom")
	})

	err := cleaner.ExecuteAndClear()
	if err == nil || !strings.Contains(err.Error(), "panic during cleanup") {
		t.Fatalf("error = %v, want panic recovery as first LIFO error", err)
	}
}
