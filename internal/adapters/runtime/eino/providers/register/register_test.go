package register

import (
	"context"
	"strings"
	"testing"

	modelproviders "fkteams/internal/adapters/model/providers"
)

func TestInitRegistersAllProviderFactories(t *testing.T) {
	for _, provider := range []modelproviders.Type{
		modelproviders.OpenAI,
		modelproviders.DeepSeek,
		modelproviders.Claude,
		modelproviders.Ollama,
		modelproviders.Ark,
		modelproviders.Gemini,
		modelproviders.Qwen,
		modelproviders.OpenRouter,
		modelproviders.Copilot,
	} {
		t.Run(string(provider), func(t *testing.T) {
			_, err := modelproviders.NewChatModel(context.Background(), &modelproviders.Config{
				Provider: provider,
				BaseURL:  "http://127.0.0.1",
				Model:    "test-model",
			})
			if err != nil && strings.Contains(err.Error(), "未知的模型提供者") {
				t.Fatalf("%s factory was not registered: %v", provider, err)
			}
		})
	}
}
