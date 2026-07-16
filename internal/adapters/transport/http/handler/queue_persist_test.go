package handler

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"fkteams/internal/app/chat/taskstream"
	domainmessage "fkteams/internal/domain/message"
)

func TestPersistentQueueExternalizesAndRestoresBase64Images(t *testing.T) {
	rt := NewRuntime(RuntimeOptions{HistoryDir: t.TempDir()})
	sessionID := "session-1"
	imageData := []byte("fake image")
	encoded := base64.StdEncoding.EncodeToString(imageData)
	queue := []taskstream.QueuedMessage{{
		ID:          "queue-1",
		Kind:        taskstream.QueueFollowUp,
		Text:        "看图",
		DisplayText: "看图",
		Parts: []domainmessage.ContentPart{{
			Type:       domainmessage.ContentPartImageURL,
			Name:       "paste.png",
			Base64Data: encoded,
			MIMEType:   "image/png",
		}},
	}}

	if err := rt.savePersistentQueue(sessionID, queue); err != nil {
		t.Fatalf("save queue: %v", err)
	}

	sessionDir := rt.sessionDirPath(sessionID)
	if _, err := os.Stat(filepath.Join(sessionDir, queueSnapshotFileName)); err != nil {
		t.Fatalf("queue snapshot was not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(sessionDir, "attachments", "queue", "queue-1", "00-paste.png")); err != nil {
		t.Fatalf("queue attachment was not written: %v", err)
	}

	restored, err := rt.loadPersistentQueue(sessionID)
	if err != nil {
		t.Fatalf("load queue: %v", err)
	}
	if len(restored) != 1 {
		t.Fatalf("restored queue length = %d, want 1", len(restored))
	}
	if got := restored[0].Parts[0].Base64Data; got != encoded {
		t.Fatalf("restored base64 = %q, want %q", got, encoded)
	}
}

func TestPersistentQueueCleanupRemovesSnapshotAndAttachments(t *testing.T) {
	rt := NewRuntime(RuntimeOptions{HistoryDir: t.TempDir()})
	sessionID := "session-1"
	queue := []taskstream.QueuedMessage{{
		ID:          "queue-1",
		Kind:        taskstream.QueueFollowUp,
		Text:        "看图",
		DisplayText: "看图",
		Parts: []domainmessage.ContentPart{{
			Type:       domainmessage.ContentPartImageURL,
			Name:       "paste.png",
			Base64Data: base64.StdEncoding.EncodeToString([]byte("fake image")),
			MIMEType:   "image/png",
		}},
	}}
	if err := rt.savePersistentQueue(sessionID, queue); err != nil {
		t.Fatalf("save queue: %v", err)
	}
	if err := rt.savePersistentQueue(sessionID, nil); err != nil {
		t.Fatalf("clear queue: %v", err)
	}

	sessionDir := rt.sessionDirPath(sessionID)
	if _, err := os.Stat(filepath.Join(sessionDir, queueSnapshotFileName)); !os.IsNotExist(err) {
		t.Fatalf("queue snapshot should be removed, err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(sessionDir, "attachments", "queue")); !os.IsNotExist(err) {
		t.Fatalf("queue attachments should be removed, err=%v", err)
	}
}

func TestPersistentQueueRejectsOversizedState(t *testing.T) {
	rt := NewRuntime(RuntimeOptions{HistoryDir: t.TempDir()})
	sessionID := "session-1"
	sessionDir := rt.sessionDirPath(sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(persistentQueuePath(sessionDir), make([]byte, maxPersistentQueueSnapshotBytes+1), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := rt.loadPersistentQueue(sessionID); err == nil {
		t.Fatal("oversized queue snapshot was accepted")
	}

	items := make([]taskstream.QueuedMessage, maxPersistentQueueItems+1)
	data, err := json.Marshal(queueSnapshotFile{Version: 1, SessionID: sessionID, Items: items})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(persistentQueuePath(sessionDir), data, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := rt.loadPersistentQueue(sessionID); err == nil {
		t.Fatal("queue snapshot with too many items was accepted")
	}

	attachmentPath := filepath.Join(sessionDir, "attachments", "queue", "queue-1", "00-image.png")
	if err := os.MkdirAll(filepath.Dir(attachmentPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(attachmentPath, make([]byte, maxPersistentQueueAttachmentBytes+1), 0644); err != nil {
		t.Fatal(err)
	}
	data, err = json.Marshal(queueSnapshotFile{
		Version:   1,
		SessionID: sessionID,
		Items: []taskstream.QueuedMessage{{
			ID: "queue-1",
			Parts: []domainmessage.ContentPart{{
				Type: domainmessage.ContentPartImageURL,
				URL:  "attachments/queue/queue-1/00-image.png",
			}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(persistentQueuePath(sessionDir), data, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := rt.loadPersistentQueue(sessionID); err == nil {
		t.Fatal("oversized queue attachment was accepted")
	}
}
