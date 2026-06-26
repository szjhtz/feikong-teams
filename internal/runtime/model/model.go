// Package model 提供运行时无关的模型创建注册表。
package model

import (
	"context"
	runtimeport "fkteams/internal/ports/runtime"
	"fmt"
	"strings"
	"sync"
)

// Type 模型提供者类型。
type Type string

const (
	OpenAI     Type = "openai"
	DeepSeek   Type = "deepseek"
	Claude     Type = "claude"
	Ollama     Type = "ollama"
	Ark        Type = "ark"
	Gemini     Type = "gemini"
	Qwen       Type = "qwen"
	OpenRouter Type = "openrouter"
	Copilot    Type = "copilot"
)

// Config 是创建聊天模型所需的最小配置。
type Config struct {
	Provider     Type
	APIKey       string
	BaseURL      string
	Model        string
	ExtraHeaders map[string]string
}

// Factory 创建运行时聊天模型。
type Factory func(ctx context.Context, cfg *Config) (runtimeport.ChatModel, error)

var (
	mu        sync.RWMutex
	factories = map[Type]Factory{}
)

// Register 注册模型提供者工厂。
func Register(t Type, f Factory) {
	mu.Lock()
	defer mu.Unlock()
	factories[t] = f
}

// NewChatModel 根据配置创建聊天模型。
func NewChatModel(ctx context.Context, cfg *Config) (runtimeport.ChatModel, error) {
	if cfg == nil {
		return nil, fmt.Errorf("model config is nil")
	}
	t := cfg.Provider
	if t == "" {
		t = Detect(cfg.BaseURL, cfg.Model)
	}

	mu.RLock()
	f := factories[t]
	mu.RUnlock()
	if f == nil {
		return nil, fmt.Errorf("未知的模型提供者: %s", t)
	}
	return f(ctx, cfg)
}

// Detect 从 BaseURL 或模型名称自动检测提供者类型。
func Detect(baseURL, modelName string) Type {
	lower := strings.ToLower(baseURL + " " + modelName)
	switch {
	case strings.Contains(lower, "deepseek"):
		return DeepSeek
	case strings.Contains(lower, "anthropic") || strings.Contains(lower, "claude"):
		return Claude
	case strings.Contains(lower, "ollama") || strings.Contains(lower, "11434"):
		return Ollama
	case strings.Contains(lower, "volces.com") || strings.Contains(lower, "volcengine"):
		return Ark
	case strings.Contains(lower, "generativelanguage.googleapis.com") || strings.Contains(lower, "gemini"):
		return Gemini
	case strings.Contains(lower, "dashscope") || strings.Contains(lower, "qwen"):
		return Qwen
	case strings.Contains(lower, "openrouter"):
		return OpenRouter
	case strings.Contains(lower, "copilot") || strings.Contains(lower, "githubcopilot"):
		return Copilot
	default:
		return OpenAI
	}
}
