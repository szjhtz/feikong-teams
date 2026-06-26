package deepseek

import (
	"context"

	deepseekModel "github.com/cloudwego/eino-ext/components/model/deepseek"

	"fkteams/internal/adapters/model/providers/providerkit"
	einoruntime "fkteams/internal/adapters/runtime/eino"
	runtimeport "fkteams/internal/ports/runtime"
)

// New 创建 DeepSeek 原生 API 的聊天模型
func New(ctx context.Context, cfg *providerkit.Config) (runtimeport.ChatModel, error) {
	chatModel, err := deepseekModel.NewChatModel(ctx, &deepseekModel.ChatModelConfig{
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
