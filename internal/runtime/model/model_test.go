package model

import (
	"context"
	runtimeport "fkteams/internal/ports/runtime"
	"strings"
	"testing"
)

type fakeChatModel struct {
	runtimeport.ChatModel
}

func TestNewChatModelUsesRegisteredFactoryAndDetectsProvider(t *testing.T) {
	registry := NewRegistry()
	registry.Register(DeepSeek, func(ctx context.Context, cfg *Config) (runtimeport.ChatModel, error) {
		if cfg.Model != "deepseek-chat" {
			t.Fatalf("model = %q, want deepseek-chat", cfg.Model)
		}
		return fakeChatModel{}, nil
	})
	got, err := registry.NewChatModel(context.Background(), &Config{Model: "deepseek-chat"})
	if err != nil {
		t.Fatalf("NewChatModel: %v", err)
	}
	if got == nil {
		t.Fatal("expected chat model")
	}
}

func TestNewChatModelReportsMissingFactory(t *testing.T) {
	registry := NewRegistry()
	_, err := registry.NewChatModel(context.Background(), &Config{Provider: Type("missing")})
	if err == nil || !strings.Contains(err.Error(), "未知的模型提供者") {
		t.Fatalf("error = %v, want missing provider error", err)
	}
}

func TestRegistryContext(t *testing.T) {
	registry := NewRegistry()
	ctx := WithRegistry(context.Background(), registry)
	got, err := RequireRegistry(ctx)
	if err != nil {
		t.Fatalf("RequireRegistry: %v", err)
	}
	if got != registry {
		t.Fatal("registry from context mismatch")
	}
}
