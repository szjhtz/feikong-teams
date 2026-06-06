package eino

import (
	"fkteams/agentcore"
	"fmt"

	"github.com/cloudwego/eino/components/model"
)

func AdaptChatModelForRunner(m agentcore.ChatModel) (model.ToolCallingChatModel, error) {
	if m == nil {
		return nil, fmt.Errorf("model is nil")
	}
	runtimeModel := m.RuntimeModel()
	chatModel, ok := runtimeModel.(model.ToolCallingChatModel)
	if !ok {
		return nil, fmt.Errorf("unsupported runtime model: %T", runtimeModel)
	}
	return chatModel, nil
}
