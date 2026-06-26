package copilot

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestTokenExpiration(t *testing.T) {
	if !(&Token{ExpiresAt: time.Now().Add(30 * time.Second).Unix()}).IsExpired() {
		t.Fatal("token expiring within 60 seconds should be expired")
	}
	if (&Token{ExpiresAt: time.Now().Add(2 * time.Minute).Unix()}).IsExpired() {
		t.Fatal("token expiring later should not be expired")
	}
}

func TestTokenManagerLoadsAndPersistsToken(t *testing.T) {
	t.Setenv("FEIKONG_APP_DIR", t.TempDir())
	token := &Token{
		GitHubToken:  "gh",
		CopilotToken: "copilot",
		ExpiresAt:    time.Now().Add(time.Hour).Unix(),
	}

	tm := NewTokenManager()
	if tm.HasToken() {
		t.Fatal("new token manager should not have a token before persistence")
	}
	if _, err := tm.GetToken(context.Background()); err == nil || !strings.Contains(err.Error(), "未登录 GitHub Copilot") {
		t.Fatalf("GetToken() error = %v, want login error", err)
	}
	if err := tm.SetToken(token); err != nil {
		t.Fatalf("SetToken() error = %v", err)
	}
	if !tm.HasToken() {
		t.Fatal("token manager should have token after SetToken")
	}
	got, err := tm.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken() error = %v", err)
	}
	if got != "copilot" {
		t.Fatalf("GetToken() = %q, want copilot", got)
	}

	loaded := NewTokenManager()
	if !loaded.HasToken() {
		t.Fatal("loaded token manager should have token")
	}
	got, err = loaded.GetToken(context.Background())
	if err != nil {
		t.Fatalf("loaded GetToken() error = %v", err)
	}
	if got != "copilot" {
		t.Fatalf("loaded token = %q, want copilot", got)
	}
}

func TestLoadTokenFromDiskRejectsInvalidContent(t *testing.T) {
	t.Setenv("FEIKONG_APP_DIR", t.TempDir())
	if err := os.MkdirAll(filepath.Dir(tokenFilePath()), 0700); err != nil {
		t.Fatalf("mkdir token dir: %v", err)
	}
	if err := os.WriteFile(tokenFilePath(), []byte(`{"copilot_token":"missing github token"}`), 0600); err != nil {
		t.Fatalf("write token file: %v", err)
	}

	if _, err := loadTokenFromDisk(); err == nil || !strings.Contains(err.Error(), "token 文件内容无效") {
		t.Fatalf("loadTokenFromDisk() error = %v, want invalid content", err)
	}
}

func TestGetOrCreateDeviceIDPersistsValue(t *testing.T) {
	t.Setenv("FEIKONG_APP_DIR", t.TempDir())

	first := getOrCreateDeviceID()
	if first == "" {
		t.Fatal("device id should not be empty")
	}
	second := getOrCreateDeviceID()
	if second != first {
		t.Fatalf("device id = %q, want persisted %q", second, first)
	}

	data, err := os.ReadFile(deviceIDFilePath())
	if err != nil {
		t.Fatalf("read device id file: %v", err)
	}
	if string(data) != first {
		t.Fatalf("device id file = %q, want %q", data, first)
	}
}

func TestImportFromVSCode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("LOCALAPPDATA", home)

	path := vsCodeTokenPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatalf("mkdir vscode token dir: %v", err)
	}
	content := `{"github.com:` + clientID + `":{"oauth_token":"gh-token"}}`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write vscode token: %v", err)
	}

	token, ok := ImportFromVSCode()
	if !ok || token != "gh-token" {
		t.Fatalf("ImportFromVSCode() = (%q, %v), want gh-token true", token, ok)
	}
}

func TestImportFromVSCodeHandlesMissingOrInvalidFiles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("LOCALAPPDATA", home)
	if runtime.GOOS != "windows" {
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	}

	if token, ok := ImportFromVSCode(); ok || token != "" {
		t.Fatalf("ImportFromVSCode() = (%q, %v), want empty false for missing file", token, ok)
	}
	path := vsCodeTokenPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatalf("mkdir vscode token dir: %v", err)
	}
	if err := os.WriteFile(path, []byte("{"), 0600); err != nil {
		t.Fatalf("write invalid vscode token: %v", err)
	}
	if token, ok := ImportFromVSCode(); ok || token != "" {
		t.Fatalf("ImportFromVSCode() = (%q, %v), want empty false for invalid file", token, ok)
	}
}

func TestCopilotHeadersAndBaseURL(t *testing.T) {
	if BaseURL() != copilotBaseURL {
		t.Fatalf("BaseURL() = %q, want %q", BaseURL(), copilotBaseURL)
	}
	headers := copilotHeaders()
	for _, key := range []string{"User-Agent", "Editor-Version", "Editor-Plugin-Version", "Copilot-Integration-Id", "X-Github-Api-Version"} {
		if headers[key] == "" {
			t.Fatalf("%s header is empty", key)
		}
	}
}
