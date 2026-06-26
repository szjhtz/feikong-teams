package commands

import (
	"context"
	"regexp"
	"strings"
	"testing"
)

func TestGenerateAPIKeyCommand(t *testing.T) {
	output := captureStdout(t, func() {
		if err := generateCommand().Run(context.Background(), []string{"fkteams", "apikey"}); err != nil {
			t.Fatalf("generate apikey returned error: %v", err)
		}
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 3 {
		t.Fatalf("output lines = %#v, want key and config hint", lines)
	}
	key := lines[0]
	if !regexp.MustCompile(`^sk-fkteams-[0-9a-f]{48}$`).MatchString(key) {
		t.Fatalf("generated key = %q, want sk-fkteams prefix with 48 hex chars", key)
	}
	if !strings.Contains(output, `api_keys = ["`+key+`"]`) {
		t.Fatalf("output missing config hint for generated key: %q", output)
	}
}

func TestAllProviderNames(t *testing.T) {
	got := AllProviderNames()
	want := []string{"copilot", "openai", "deepseek", "claude", "gemini", "qwen", "ollama", "ark", "openrouter", "custom"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("AllProviderNames = %#v, want %#v", got, want)
	}
}

func TestDefaultOrNone(t *testing.T) {
	if got := defaultOrNone(""); got != "无" {
		t.Fatalf("defaultOrNone empty = %q, want 无", got)
	}
	if got := defaultOrNone("https://example.com"); got != "https://example.com" {
		t.Fatalf("defaultOrNone value = %q", got)
	}
}
