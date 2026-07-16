package handler

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPersistedShareStoreRejectsOversizedAndMalformedFiles(t *testing.T) {
	oversizedPath := filepath.Join(t.TempDir(), "oversized.json")
	if err := os.WriteFile(oversizedPath, bytes.Repeat([]byte{'x'}, maxPersistedShareStoreBytes+1), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := readPersistedShareStore(oversizedPath); err == nil {
		t.Fatal("oversized share store was accepted")
	}

	previewPath := filepath.Join(t.TempDir(), "preview.json")
	if err := os.WriteFile(previewPath, []byte(`{"broken":null}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := NewPreviewLinkStore(previewPath).LoadError(); err == nil {
		t.Fatal("malformed preview link entry was accepted")
	}

	sessionPath := filepath.Join(t.TempDir(), "session.json")
	if err := os.WriteFile(sessionPath, []byte(`{"broken":null}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := NewSessionShareStore(sessionPath).LoadError(); err == nil {
		t.Fatal("malformed session share entry was accepted")
	}
}

func TestShareStoresEnforceEntryLimit(t *testing.T) {
	previewStore := NewPreviewLinkStore(filepath.Join(t.TempDir(), "preview.json"))
	previewStore.Lock()
	for index := range maxPersistedShareEntries {
		previewStore.m[fmt.Sprintf("link-%d", index)] = newPreviewLinkEntry([]string{"file.txt"}, nil, "", time.Time{}, time.Now())
	}
	previewStore.Unlock()
	if err := previewStore.Put("overflow", newPreviewLinkEntry([]string{"file.txt"}, nil, "", time.Time{}, time.Now())); !errors.Is(err, errShareStoreFull) {
		t.Fatalf("preview overflow error = %v", err)
	}

	sessionStore := NewSessionShareStore(filepath.Join(t.TempDir(), "session.json"))
	sessionStore.Lock()
	for index := range maxPersistedShareEntries {
		sessionStore.m[fmt.Sprintf("share-%d", index)] = &sessionShareEntry{SessionID: "session-1"}
	}
	sessionStore.Unlock()
	if err := sessionStore.Put("overflow", &sessionShareEntry{SessionID: "session-1"}); !errors.Is(err, errShareStoreFull) {
		t.Fatalf("session overflow error = %v", err)
	}
}

func TestSessionShareTouchThrottlesPersistence(t *testing.T) {
	store := NewSessionShareStore(filepath.Join(t.TempDir(), "session.json"))
	store.Lock()
	store.m["share-1"] = &sessionShareEntry{SessionID: "session-1", CreatedAt: time.Now()}
	store.Unlock()

	first := time.Unix(1000, 0)
	if err := store.Touch("share-1", first); err != nil {
		t.Fatal(err)
	}
	if err := store.Touch("share-1", first.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	store.RLock()
	lastAccessedAt := store.m["share-1"].LastAccessedAt
	store.RUnlock()
	if !lastAccessedAt.Equal(first) {
		t.Fatalf("last access = %v, want throttled value %v", lastAccessedAt, first)
	}

	second := first.Add(6 * time.Minute)
	if err := store.Touch("share-1", second); err != nil {
		t.Fatal(err)
	}
	store.RLock()
	lastAccessedAt = store.m["share-1"].LastAccessedAt
	store.RUnlock()
	if !lastAccessedAt.Equal(second) {
		t.Fatalf("last access = %v, want %v", lastAccessedAt, second)
	}
}
