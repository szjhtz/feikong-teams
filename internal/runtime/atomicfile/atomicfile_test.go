package atomicfile

import (
	"errors"
	"io"
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

func TestWriteInRootPreservesTargetWhenWriterFails(t *testing.T) {
	directory := t.TempDir()
	target := filepath.Join(directory, "data.txt")
	if err := os.WriteFile(target, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	wantErr := errors.New("write failed")
	_, err = WriteInRoot(root, "data.txt", 0644, func(writer io.Writer) (int64, error) {
		written, writeErr := io.WriteString(writer, "partial")
		if writeErr != nil {
			return int64(written), writeErr
		}
		return int64(written), wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("WriteInRoot() error = %v, want writer error", err)
	}
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "old" {
		t.Fatalf("target content = %q, want old", content)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "data.txt" {
		t.Fatalf("temporary file was not cleaned up: %#v", entries)
	}
}

func TestWriteReaderInRootEnforcesLimit(t *testing.T) {
	directory := t.TempDir()
	root, err := os.OpenRoot(directory)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	_, err = WriteReaderInRoot(root, "data.txt", &repeatingReader{remaining: 5}, 4, 0644)
	if err == nil {
		t.Fatal("WriteReaderInRoot() should reject oversized content")
	}
	if _, statErr := os.Stat(filepath.Join(directory, "data.txt")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("target should not exist, stat error = %v", statErr)
	}
}

func TestWriteInRootRejectsEscape(t *testing.T) {
	directory := t.TempDir()
	root, err := os.OpenRoot(directory)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	err = WriteFileInRoot(root, filepath.Join("..", "outside.txt"), []byte("data"), 0644)
	if err == nil {
		t.Fatal("WriteFileInRoot() should reject paths outside root")
	}
}

type repeatingReader struct {
	remaining int
}

func (r *repeatingReader) Read(buffer []byte) (int, error) {
	if r.remaining == 0 {
		return 0, io.EOF
	}
	count := min(len(buffer), r.remaining)
	for i := range count {
		buffer[i] = 'x'
	}
	r.remaining -= count
	return count, nil
}
