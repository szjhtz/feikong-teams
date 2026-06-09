package common

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadHistoryKeepsTail(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "history.txt")
	if err := SaveHistory(path, []string{"one", "two", "three"}); err != nil {
		t.Fatalf("SaveHistory: %v", err)
	}

	got, err := LoadHistory(path, 2)
	if err != nil {
		t.Fatalf("LoadHistory: %v", err)
	}
	if len(got) != 2 || got[0] != "two" || got[1] != "three" {
		t.Fatalf("history = %#v, want tail two entries", got)
	}
}

func TestLoadHistoryMissingFileReturnsEmpty(t *testing.T) {
	got, err := LoadHistory(filepath.Join(t.TempDir(), "missing.txt"), 10)
	if err != nil {
		t.Fatalf("LoadHistory missing: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("history = %#v, want empty", got)
	}
}

func TestLoadHistoryReturnsOpenError(t *testing.T) {
	dir := t.TempDir()
	if _, err := LoadHistory(dir, 10); err == nil {
		t.Fatal("expected error when loading directory as file")
	}
}

func TestSaveHistoryReturnsCreateError(t *testing.T) {
	dir := t.TempDir()
	blockingFile := filepath.Join(dir, "file")
	if err := os.WriteFile(blockingFile, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := SaveHistory(filepath.Join(blockingFile, "history.txt"), []string{"x"}); err == nil {
		t.Fatal("expected error when parent path is a file")
	}
}
