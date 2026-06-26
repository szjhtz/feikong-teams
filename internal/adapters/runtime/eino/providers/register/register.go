package register

import (
	"context"
	modelproviders "fkteams/internal/adapters/model/providers"
	"fkteams/internal/adapters/model/providers/providerkit"
	"fkteams/internal/adapters/runtime/eino/providers/ark"
	"fkteams/internal/adapters/runtime/eino/providers/claude"
	"fkteams/internal/adapters/runtime/eino/providers/copilot"
	"fkteams/internal/adapters/runtime/eino/providers/deepseek"
	"fkteams/internal/adapters/runtime/eino/providers/gemini"
	"fkteams/internal/adapters/runtime/eino/providers/ollama"
	"fkteams/internal/adapters/runtime/eino/providers/openai"
	"fkteams/internal/adapters/runtime/eino/providers/openrouter"
	"fkteams/internal/adapters/runtime/eino/providers/qwen"
	runtimeport "fkteams/internal/ports/runtime"
	modelregistry "fkteams/internal/runtime/model"
)

// RegisterDefaults 显式注册 Eino runtime 的内置模型提供者。
func RegisterDefaults(registry *modelregistry.Registry) {
	modelproviders.RegisterDefaultModelListers()

	modelproviders.Register(modelproviders.OpenAI, openai.New)
	modelproviders.Register(modelproviders.DeepSeek, deepseek.New)
	modelproviders.Register(modelproviders.Claude, claude.New)
	modelproviders.Register(modelproviders.Ollama, ollama.New)
	modelproviders.Register(modelproviders.Ark, ark.New)
	modelproviders.Register(modelproviders.Gemini, gemini.New)
	modelproviders.Register(modelproviders.Qwen, qwen.New)
	modelproviders.Register(modelproviders.OpenRouter, openrouter.New)
	modelproviders.Register(modelproviders.Copilot, copilot.New)

	registerRuntimeModel(registry, modelregistry.OpenAI, openai.New)
	registerRuntimeModel(registry, modelregistry.DeepSeek, deepseek.New)
	registerRuntimeModel(registry, modelregistry.Claude, claude.New)
	registerRuntimeModel(registry, modelregistry.Ollama, ollama.New)
	registerRuntimeModel(registry, modelregistry.Ark, ark.New)
	registerRuntimeModel(registry, modelregistry.Gemini, gemini.New)
	registerRuntimeModel(registry, modelregistry.Qwen, qwen.New)
	registerRuntimeModel(registry, modelregistry.OpenRouter, openrouter.New)
	registerRuntimeModel(registry, modelregistry.Copilot, copilot.New)
}

func registerRuntimeModel(registry *modelregistry.Registry, t modelregistry.Type, f func(context.Context, *providerkit.Config) (runtimeport.ChatModel, error)) {
	if registry == nil {
		return
	}
	registry.Register(t, func(ctx context.Context, cfg *modelregistry.Config) (runtimeport.ChatModel, error) {
		return f(ctx, &providerkit.Config{
			Provider:     string(cfg.Provider),
			APIKey:       cfg.APIKey,
			BaseURL:      cfg.BaseURL,
			Model:        cfg.Model,
			ExtraHeaders: cfg.ExtraHeaders,
		})
	})
}
