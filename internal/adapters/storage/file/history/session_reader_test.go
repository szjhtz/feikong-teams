package eventlog

import (
	"context"
	"path/filepath"
	"testing"

	domainmessage "fkteams/internal/domain/message"
)

func TestSessionMessageReaderLoadsActiveRecorder(t *testing.T) {
	manager := NewSessionHistoryManager()
	recorder := manager.GetOrCreate("active-session", t.TempDir())
	recorder.RecordUserMessage(domainmessage.Message{Role: domainmessage.RoleUser, Content: "hello"})

	messages, err := NewSessionMessageReader(t.TempDir(), manager).LoadSessionMessages(context.Background(), "active-session")
	if err != nil {
		t.Fatalf("LoadSessionMessages returned error: %v", err)
	}
	if len(messages) != 1 || messages[0].GetTextContent() != "hello" {
		t.Fatalf("messages = %#v, want active recorder message", messages)
	}
}

func TestSessionMessageReaderLoadsPersistedHistory(t *testing.T) {
	dir := t.TempDir()
	sessionID := "persisted-session"
	recorder := NewHistoryRecorder()
	recorder.RecordUserMessage(domainmessage.Message{Role: domainmessage.RoleUser, Content: "from disk"})
	if err := recorder.SaveToFile(filepath.Join(dir, sessionID, HistoryFileName)); err != nil {
		t.Fatalf("SaveToFile returned error: %v", err)
	}

	messages, err := NewSessionMessageReader(dir, NewSessionHistoryManager()).LoadSessionMessages(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("LoadSessionMessages returned error: %v", err)
	}
	if len(messages) != 1 || messages[0].GetTextContent() != "from disk" {
		t.Fatalf("messages = %#v, want persisted message", messages)
	}
}
