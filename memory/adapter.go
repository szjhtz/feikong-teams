package memory

import (
	"context"
	"fkteams/agentcore"
	einoruntime "fkteams/agentcore/eino"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"fkteams/providers/copilot"
)

// einoLLMAdapter 将 Eino ChatModel 适配为 LLMClient 接口
type einoLLMAdapter struct {
	model model.BaseChatModel
}

// NewLLMClient 基于核心模型创建 LLMClient
func NewLLMClient(m agentcore.ChatModel) (LLMClient, error) {
	chatModel, err := einoruntime.AdaptChatModelForRunner(m)
	if err != nil {
		return nil, err
	}
	return &einoLLMAdapter{model: chatModel}, nil
}

func (a *einoLLMAdapter) Complete(ctx context.Context, prompt string) (string, error) {
	ctx = copilot.WithAgentInitiator(ctx)
	resp, err := a.model.Generate(ctx, []*schema.Message{
		schema.SystemMessage("You are a memory extraction assistant. Respond only in the requested format."),
		schema.UserMessage(prompt),
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}
