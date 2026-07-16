package chat

import (
	"context"
	"testing"
)

type metadataStoreSpy struct {
	update MetadataUpdate
}

func (s *metadataStoreSpy) UpdateMetadata(_ context.Context, update MetadataUpdate) error {
	s.update = update
	return nil
}

func TestMarkProcessingCreatesMissingSessionMetadata(t *testing.T) {
	store := &metadataStoreSpy{}
	lifecycle := NewSessionLifecycle(nil, store)

	if err := lifecycle.MarkProcessing(context.Background(), "session-1", "用户问题"); err != nil {
		t.Fatalf("mark processing: %v", err)
	}

	if !store.update.CreateIfMissing {
		t.Fatal("processing metadata update should create missing session metadata")
	}
	if store.update.SessionID != "session-1" || store.update.TitleSource != "用户问题" || store.update.Status != SessionStatusProcessing {
		t.Fatalf("metadata update = %#v", store.update)
	}
}

func TestMarkProcessingWithTargetPersistsExecutionSelection(t *testing.T) {
	store := &metadataStoreSpy{}
	lifecycle := NewSessionLifecycle(nil, store)

	if err := lifecycle.MarkProcessingWithTarget(context.Background(), "session-1", "用户问题", ExecutionTarget{
		Mode:         "deep",
		CurrentAgent: "coder",
	}); err != nil {
		t.Fatalf("mark processing with target: %v", err)
	}

	if store.update.Mode == nil || *store.update.Mode != "deep" {
		t.Fatalf("mode update = %#v", store.update.Mode)
	}
	if store.update.CurrentAgent == nil || *store.update.CurrentAgent != "coder" {
		t.Fatalf("agent update = %#v", store.update.CurrentAgent)
	}
}
