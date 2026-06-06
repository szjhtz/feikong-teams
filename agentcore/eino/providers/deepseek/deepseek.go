package deepseek

import (
	"context"

	"fkteams/agentcore"
	einoruntime "fkteams/agentcore/eino"
	"fkteams/providers/providerkit"
)

// New 创建 DeepSeek 原生 API 的聊天模型
func New(ctx context.Context, cfg *providerkit.Config) (agentcore.ChatModel, error) {
	chatModel, err := NewChatModel(ctx, &ChatModelConfig{
		APIKey:     cfg.APIKey,
		BaseURL:    cfg.BaseURL,
		Model:      cfg.Model,
		HTTPClient: providerkit.HTTPClientWithHeaders(cfg.ExtraHeaders),
	})
	if err != nil {
		return nil, err
	}
	return einoruntime.WrapChatModel(chatModel), nil
}
