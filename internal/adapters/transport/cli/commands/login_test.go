package commands

import (
	"strings"
	"testing"

	"fkteams/internal/app/config"
)

func TestAPIKeyFlags(t *testing.T) {
	withKey := apiKeyFlags("https://example.com", true)
	if len(withKey) != 4 || withKey[0].Names()[0] != "api-key" {
		t.Fatalf("apiKeyFlags with key = %#v", withKey)
	}

	withoutKey := apiKeyFlags("", false)
	if len(withoutKey) != 3 {
		t.Fatalf("apiKeyFlags without key count = %d, want 3", len(withoutKey))
	}
	if got := withoutKey[0].Names()[0]; got != "base-url" {
		t.Fatalf("first flag without key = %q, want base-url", got)
	}
}

func TestSaveProviderConfigCreatesDefault(t *testing.T) {
	useTempAppDir(t)

	output := captureStdout(t, func() {
		if err := saveProviderConfig("deepseek", "deepseek", "sk-test", "https://api.deepseek.com", "deepseek-chat"); err != nil {
			t.Fatalf("saveProviderConfig returned error: %v", err)
		}
	})
	if !strings.Contains(output, "已新增供应商配置") || !strings.Contains(output, "已自动设为默认对话模型") {
		t.Fatalf("saveProviderConfig output = %q", output)
	}

	cfg := config.Get()
	if len(cfg.Models) != 1 {
		t.Fatalf("models count = %d, want one provider model: %#v", len(cfg.Models), cfg.Models)
	}
	model := cfg.ResolveModel("deepseek")
	if model == nil || model.Provider != "deepseek" || model.APIKey != "sk-test" || model.Model != "deepseek-chat" {
		t.Fatalf("deepseek model = %#v", model)
	}
	defaultModel := cfg.ResolveDefaultModel(config.ModelUseChat)
	if defaultModel == nil || defaultModel.ID != "deepseek" || defaultModel.Provider != "deepseek" || defaultModel.Model != "deepseek-chat" {
		t.Fatalf("default chat model = %#v", defaultModel)
	}
}

func TestSaveProviderConfigUpdatesExisting(t *testing.T) {
	useTempAppDir(t)
	if err := config.Save(&config.Config{Models: []config.ModelConfig{
		{ID: "openai", Name: "OpenAI", Provider: "openai", APIKey: "old", BaseURL: "https://old.example.com", Model: "old-model"},
		{ID: "deepseek", Name: "DeepSeek", UseFor: []string{config.ModelUseChat}, Provider: "deepseek", APIKey: "keep", Model: "deepseek-chat"},
	}}); err != nil {
		t.Fatalf("save initial config: %v", err)
	}

	output := captureStdout(t, func() {
		if err := saveProviderConfig("openai", "openai", "new-key", "", "gpt-5"); err != nil {
			t.Fatalf("saveProviderConfig returned error: %v", err)
		}
	})
	if !strings.Contains(output, "已更新供应商配置") {
		t.Fatalf("saveProviderConfig output = %q", output)
	}

	cfg := config.Get()
	model := cfg.ResolveModel("openai")
	if model == nil || model.APIKey != "new-key" || model.BaseURL != "https://old.example.com" || model.Model != "gpt-5" {
		t.Fatalf("updated openai model = %#v", model)
	}
	defaultModel := cfg.ResolveDefaultModel(config.ModelUseChat)
	if defaultModel == nil || defaultModel.Provider != "deepseek" {
		t.Fatalf("default chat model should be preserved, got %#v", defaultModel)
	}
}

func TestProviderEntriesContainExpectedDefaults(t *testing.T) {
	entries := make(map[string]providerEntry, len(knownProviders))
	for _, entry := range knownProviders {
		entries[entry.name] = entry
	}
	if entries["openai"].defaultURL != "https://api.openai.com/v1" || !entries["openai"].needKey {
		t.Fatalf("openai entry = %#v", entries["openai"])
	}
	if entries["ollama"].defaultURL != "http://localhost:11434/v1" || entries["ollama"].needKey {
		t.Fatalf("ollama entry = %#v", entries["ollama"])
	}
}
