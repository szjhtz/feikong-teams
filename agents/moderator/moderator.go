package moderator

import (
	"context"
	"fkteams/agentcore"
	"fkteams/agents/common"
)

func NewAgent(ctx context.Context, agentTools ...agentcore.Tool) (agentcore.Agent, error) {
	return common.NewAgentBuilder("moderator", "会议主持人，负责引导讨论、指定发言成员并形成结论。").
		WithInstruction(moderatorPrompt).
		WithTools(agentTools...).
		WithSummary().
		Build(ctx)
}
