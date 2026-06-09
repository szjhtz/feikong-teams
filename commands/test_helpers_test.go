package commands

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"fkteams/config"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = old
		_ = w.Close()
		_ = r.Close()
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read stdout pipe: %v", err)
	}
	return buf.String()
}

func useTempAppDir(t *testing.T) string {
	t.Helper()

	appDir := t.TempDir()
	t.Setenv("FEIKONG_APP_DIR", appDir)
	if err := os.MkdirAll(filepath.Join(appDir, "config"), 0755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	if err := config.Save(&config.Config{}); err != nil {
		t.Fatalf("save initial config: %v", err)
	}
	return appDir
}
