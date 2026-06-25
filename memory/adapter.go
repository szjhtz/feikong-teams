package memory

import (
	"context"
	domainmessage "fkteams/internal/domain/message"
	runtimeport "fkteams/internal/ports/runtime"

	"fkteams/providers/copilot"
)

type chatModelLLMAdapter struct {
	model runtimeport.ChatModel
}

// NewLLMClient 基于核心模型创建 LLMClient
func NewLLMClient(m runtimeport.ChatModel) (LLMClient, error) {
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
