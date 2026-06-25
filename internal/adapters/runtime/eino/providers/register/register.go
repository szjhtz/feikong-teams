package register

import (
	modelproviders "fkteams/internal/adapters/model/providers"
	"fkteams/internal/adapters/runtime/eino/providers/ark"
	"fkteams/internal/adapters/runtime/eino/providers/claude"
	"fkteams/internal/adapters/runtime/eino/providers/copilot"
	"fkteams/internal/adapters/runtime/eino/providers/deepseek"
	"fkteams/internal/adapters/runtime/eino/providers/gemini"
	"fkteams/internal/adapters/runtime/eino/providers/ollama"
	"fkteams/internal/adapters/runtime/eino/providers/openai"
	"fkteams/internal/adapters/runtime/eino/providers/openrouter"
	"fkteams/internal/adapters/runtime/eino/providers/qwen"
)

func init() {
	modelproviders.Register(modelproviders.OpenAI, openai.New)
	modelproviders.Register(modelproviders.DeepSeek, deepseek.New)
	modelproviders.Register(modelproviders.Claude, claude.New)
	modelproviders.Register(modelproviders.Ollama, ollama.New)
	modelproviders.Register(modelproviders.Ark, ark.New)
	modelproviders.Register(modelproviders.Gemini, gemini.New)
	modelproviders.Register(modelproviders.Qwen, qwen.New)
	modelproviders.Register(modelproviders.OpenRouter, openrouter.New)
	modelproviders.Register(modelproviders.Copilot, copilot.New)
}
