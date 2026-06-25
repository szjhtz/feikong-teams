package atomicfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileCreatesParentAndWritesData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "data.txt")

	if err := WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("unexpected content: %q", got)
	}
}

func TestWriteFileReplacesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.txt")
	if err := os.WriteFile(path, []byte("old"), 0644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	if err := WriteFile(path, []byte("new"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(got) != "new" {
		t.Fatalf("unexpected content: %q", got)
	}
}
