package memorymodel

import (
	"context"

	domainmessage "fkteams/internal/domain/message"
	memoryport "fkteams/internal/ports/memory"
	runtimeport "fkteams/internal/ports/runtime"

	"fkteams/internal/adapters/model/providers/copilot"
)

type chatModelLLMAdapter struct {
	model runtimeport.ChatModel
}

// NewLLMClient 基于运行时模型创建长期记忆提取客户端。
func NewLLMClient(m runtimeport.ChatModel) (memoryport.LLMClient, error) {
	return &chatModelLLMAdapter{model: m}, nil
}

func (a *chatModelLLMAdapter) Complete(ctx context.Context, prompt string) (string, error) {
	ctx = copilot.WithAgentInitiator(ctx)
	resp, err := a.model.Generate(ctx, []domainmessage.Message{
		{Role: domainmessage.RoleSystem, Content: "You are a memory extraction assistant. Respond only in the requested format."},
		{Role: domainmessage.RoleUser, Content: prompt},
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}
