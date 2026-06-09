package commands

import (
	"context"
	"strings"
	"testing"

	"fkteams/config"
)

func TestFindModelConfig(t *testing.T) {
	cfg := &config.Config{Models: []config.ModelConfig{
		{Name: "default", Provider: "openai", Model: "gpt-4o"},
		{Name: "work", Provider: "deepseek", Model: "deepseek-chat"},
	}}
	candidates := []config.ModelConfig{{Name: "candidate", Provider: "qwen", Model: "qwen-plus"}}

	if got := findModelConfig(cfg, candidates, "candidate"); got == nil || got.Provider != "qwen" {
		t.Fatalf("find candidate = %#v", got)
	}
	if got := findModelConfig(cfg, candidates, "deepseek"); got == nil || got.Name != "work" {
		t.Fatalf("find by provider = %#v", got)
	}
	if got := findModelConfig(cfg, candidates, "missing"); got != nil {
		t.Fatalf("find missing = %#v, want nil", got)
	}
}

func TestListModels(t *testing.T) {
	useTempAppDir(t)

	if err := listModels(); err != nil {
		t.Fatalf("listModels empty returned error: %v", err)
	}

	if err := config.Save(&config.Config{Models: []config.ModelConfig{
		{Name: "default", Provider: "openai", BaseURL: "https://api.openai.com/v1", Model: "gpt-4o"},
		{Name: "local", Provider: "ollama", Model: "llama3"},
	}}); err != nil {
		t.Fatalf("save config: %v", err)
	}
	if err := listModels(); err != nil {
		t.Fatalf("listModels returned error: %v", err)
	}
}

func TestListAvailableModelsValidationErrors(t *testing.T) {
	useTempAppDir(t)

	err := listAvailableModels(context.Background(), "", "missing")
	if err == nil || !strings.Contains(err.Error(), "未找到模型配置") {
		t.Fatalf("missing model config error = %v", err)
	}

	if err := config.Save(&config.Config{Models: []config.ModelConfig{{Name: "default"}}}); err != nil {
		t.Fatalf("save config: %v", err)
	}
	err = listAvailableModels(context.Background(), "", "default")
	if err == nil || !strings.Contains(err.Error(), "请指定服务商") {
		t.Fatalf("missing provider error = %v", err)
	}
}

func TestSwitchModelUpdatesDefaultModelName(t *testing.T) {
	useTempAppDir(t)
	if err := config.Save(&config.Config{Models: []config.ModelConfig{
		{Name: "default", Provider: "openai", APIKey: "sk-test", BaseURL: "https://api.openai.com/v1", Model: "gpt-4o"},
	}}); err != nil {
		t.Fatalf("save initial config: %v", err)
	}

	output := captureStdout(t, func() {
		if err := switchModel(context.Background(), "default", "gpt-5"); err != nil {
			t.Fatalf("switchModel returned error: %v", err)
		}
	})
	if !strings.Contains(output, "gpt-4o → gpt-5") {
		t.Fatalf("switchModel output = %q", output)
	}

	model := config.Get().ResolveModel("default")
	if model == nil || model.Model != "gpt-5" || model.Provider != "openai" {
		t.Fatalf("default model after switch = %#v", model)
	}
}

func TestSwitchModelCopiesNamedConfigToDefault(t *testing.T) {
	useTempAppDir(t)
	if err := config.Save(&config.Config{Models: []config.ModelConfig{
		{Name: "default", Provider: "openai", APIKey: "old-key", BaseURL: "https://api.openai.com/v1", Model: "gpt-4o"},
		{Name: "deepseek-work", Provider: "deepseek", APIKey: "new-key", BaseURL: "https://api.deepseek.com", Model: "deepseek-chat"},
	}}); err != nil {
		t.Fatalf("save initial config: %v", err)
	}

	output := captureStdout(t, func() {
		if err := switchModel(context.Background(), "deepseek-work", "deepseek-reasoner"); err != nil {
			t.Fatalf("switchModel returned error: %v", err)
		}
	})
	if !strings.Contains(output, "已切换默认模型") {
		t.Fatalf("switchModel output = %q", output)
	}

	model := config.Get().ResolveModel("default")
	if model == nil ||
		model.Provider != "deepseek" ||
		model.APIKey != "new-key" ||
		model.BaseURL != "https://api.deepseek.com" ||
		model.Model != "deepseek-reasoner" {
		t.Fatalf("default model after named switch = %#v", model)
	}
}

func TestSwitchModelErrorsWithoutDefault(t *testing.T) {
	useTempAppDir(t)

	err := switchModel(context.Background(), "default", "gpt-5")
	if err == nil || !strings.Contains(err.Error(), "尚未配置默认模型") {
		t.Fatalf("switchModel error = %v, want missing default", err)
	}
}

func TestRemoveModel(t *testing.T) {
	useTempAppDir(t)
	if err := config.Save(&config.Config{Models: []config.ModelConfig{
		{Name: "default", Provider: "openai"},
		{Name: "old", Provider: "deepseek"},
	}}); err != nil {
		t.Fatalf("save initial config: %v", err)
	}

	output := captureStdout(t, func() {
		if err := removeModel("old"); err != nil {
			t.Fatalf("removeModel returned error: %v", err)
		}
	})
	if !strings.Contains(output, "已移除模型配置") {
		t.Fatalf("removeModel output = %q", output)
	}
	if got := config.Get().ResolveModel("old"); got != nil {
		t.Fatalf("removed model still exists: %#v", got)
	}

	err := removeModel("missing")
	if err == nil || !strings.Contains(err.Error(), "未找到模型配置") {
		t.Fatalf("remove missing error = %v", err)
	}
}
