package memory

import (
	"context"
	"fkteams/agentcore"
	"fmt"

	"fkteams/providers/copilot"
)

type chatModelLLMAdapter struct {
	model agentcore.NativeChatModel
}

// NewLLMClient 基于核心模型创建 LLMClient
func NewLLMClient(m agentcore.ChatModel) (LLMClient, error) {
	chatModel, ok := m.(agentcore.NativeChatModel)
	if !ok {
		return nil, fmt.Errorf("model does not support native generation: %T", m)
	}
	return &chatModelLLMAdapter{model: chatModel}, nil
}

func (a *chatModelLLMAdapter) Complete(ctx context.Context, prompt string) (string, error) {
	ctx = copilot.WithAgentInitiator(ctx)
	resp, err := a.model.Generate(ctx, []agentcore.Message{
		{Role: agentcore.RoleSystem, Content: "You are a memory extraction assistant. Respond only in the requested format."},
		{Role: agentcore.RoleUser, Content: prompt},
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}
