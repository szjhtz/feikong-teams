package coder

import (
	"context"
	"fkteams/agents/common"

	"fkteams/agentcore"
)

func NewAgent(ctx context.Context) (agentcore.Agent, error) {
	return common.NewAgentBuilder("coder", "软件工程师，负责代码实现、调试、重构和工程验证。").
		WithInstruction(coderPrompt).
		WithToolNames("file", "command").
		WithSummary().
		WithSkills().
		Build(ctx)
}
