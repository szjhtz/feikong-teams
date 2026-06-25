package inputhistory

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestSaveAndLoadHistory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history", "input.txt")
	if err := Save(path, []string{"one", "two", "three"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(path, 2)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if want := []string{"two", "three"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Load = %#v, want %#v", got, want)
	}
}

func TestLoadMissingHistory(t *testing.T) {
	got, err := Load(filepath.Join(t.TempDir(), "missing.txt"), 10)
	if err != nil {
		t.Fatalf("Load missing: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("Load missing = %#v, want empty", got)
	}
}
