package taskstream

import (
	"context"
	"testing"
	"time"
)

func TestCleanupLifecycleCanBeReplacedAndStopped(t *testing.T) {
	manager := NewManager()
	manager.StartCleanup(context.Background(), time.Hour)
	manager.cleanupMu.Lock()
	first := manager.cleanupCancel
	manager.cleanupMu.Unlock()
	if first == nil {
		t.Fatal("cleanup was not started")
	}

	manager.StartCleanup(context.Background(), time.Hour)
	manager.cleanupMu.Lock()
	second := manager.cleanupCancel
	manager.cleanupMu.Unlock()
	if second == nil {
		t.Fatal("replacement cleanup was not started")
	}

	manager.StopCleanup()
	manager.cleanupMu.Lock()
	defer manager.cleanupMu.Unlock()
	if manager.cleanupCancel != nil {
		t.Fatal("cleanup cancel function was not cleared")
	}
}
