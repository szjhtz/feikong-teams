package handler

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fkteams/internal/runtime/env"
)

func TestShareLinksFilePathUsesAppDir(t *testing.T) {
	appDir := t.TempDir()
	t.Setenv(env.AppDir, appDir)

	want := filepath.Join(appDir, "share", "share.json")
	if got := shareLinksFilePath(); got != want {
		t.Fatalf("unexpected share file path: got %q, want %q", got, want)
	}
}

func TestSaveShareLinksWritesToAppDir(t *testing.T) {
	appDir := t.TempDir()
	t.Setenv(env.AppDir, appDir)
	store := NewPreviewLinkStore("")
	store.Lock()
	store.m = map[string]*previewLinkEntry{
		"link-1": {
			FilePaths:     []string{"docs/report.pdf"},
			ResourcePaths: []string{"docs/report.pdf", "docs/cover.png"},
			CreatedAt:     time.Unix(100, 0),
		},
	}
	store.Unlock()

	if err := store.Save(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(appDir, "share", "share.json"))
	if err != nil {
		t.Fatalf("read share file: %v", err)
	}
	var entries map[string]*shareFileEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("unmarshal share file: %v", err)
	}
	if got := entries["link-1"].FilePaths[0]; got != "docs/report.pdf" {
		t.Fatalf("unexpected saved file path: %q", got)
	}
	if got := entries["link-1"].ResourcePaths; len(got) != 2 || got[1] != "docs/cover.png" {
		t.Fatalf("unexpected saved resource paths: %#v", got)
	}
}

func TestLoadShareLinksReadsFromAppDir(t *testing.T) {
	appDir := t.TempDir()
	t.Setenv(env.AppDir, appDir)

	shareDir := filepath.Join(appDir, "share")
	if err := os.MkdirAll(shareDir, 0755); err != nil {
		t.Fatal(err)
	}
	data := []byte(`{"link-2":{"file_paths":["docs/manual.md"],"created_at":200}}`)
	if err := os.WriteFile(filepath.Join(shareDir, "share.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	store := NewPreviewLinkStore("")

	store.RLock()
	entry := store.m["link-2"]
	store.RUnlock()
	if entry == nil {
		t.Fatal("expected share link to be loaded")
	}
	if got := entry.FilePaths[0]; got != "docs/manual.md" {
		t.Fatalf("unexpected loaded file path: %q", got)
	}
	if !entry.allowsResource("docs/manual.md") || entry.allowsResource("docs/private.txt") {
		t.Fatalf("legacy entry resource boundary is invalid: %#v", entry.ResourcePaths)
	}
}

func TestRuntimeStartRejectsCorruptedShareState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "share.json")
	if err := os.WriteFile(path, []byte(`{broken`), 0644); err != nil {
		t.Fatal(err)
	}
	store := NewPreviewLinkStore(path)
	if store.LoadError() == nil {
		t.Fatal("expected corrupted share state to fail loading")
	}
	runtime := NewRuntime(RuntimeOptions{PreviewLinks: store})
	if err := runtime.Start(context.Background()); err == nil {
		t.Fatal("expected runtime start to reject corrupted share state")
	}
}

func TestPreviewStoreSaveReportsFilesystemFailure(t *testing.T) {
	root := t.TempDir()
	blocker := filepath.Join(root, "blocker")
	if err := os.WriteFile(blocker, []byte("not a directory"), 0644); err != nil {
		t.Fatal(err)
	}
	store := NewPreviewLinkStore(filepath.Join(blocker, "share.json"))
	if err := store.Put("link-1", &previewLinkEntry{FilePaths: []string{"a.txt"}, CreatedAt: time.Now()}); err == nil {
		t.Fatal("expected save failure")
	}
	store.RLock()
	_, exists := store.m["link-1"]
	store.RUnlock()
	if exists {
		t.Fatal("failed transaction should roll back in-memory state")
	}
}

func newPreviewTestStore(t *testing.T, entries map[string]*previewLinkEntry) *PreviewLinkStore {
	t.Helper()
	store := NewPreviewLinkStore(filepath.Join(t.TempDir(), "share.json"))
	store.Lock()
	store.m = entries
	store.Unlock()
	return store
}

func newPreviewTestRuntime(t *testing.T, entries map[string]*previewLinkEntry) *Runtime {
	t.Helper()
	return NewRuntime(RuntimeOptions{PreviewLinks: newPreviewTestStore(t, entries)})
}
