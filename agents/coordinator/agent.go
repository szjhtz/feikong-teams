package coordinator

import (
	"context"
	"fkteams/agentcore"
	"fkteams/agents/common"
)

func NewAgent(ctx context.Context, agentTools ...agentcore.Tool) (agentcore.Agent, error) {
	safeDir := common.WorkspaceDir()

	return common.NewAgentBuilder("coordinator", "核心工程智能体，直接完成常规工程任务，并按需指派专业成员。").
		WithInstruction(coordinatorPrompt).
		WithTemplateVar("workspace_dir", safeDir).
		WithToolNames("todo", "file", "scheduler", "ask").
		WithTools(agentTools...).
		WithSummary().
		WithSkills().
		Build(ctx)
}
