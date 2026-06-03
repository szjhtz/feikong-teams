package coordinator

import (
	"context"
	"fkteams/agents/common"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
)

func NewAgent(ctx context.Context, agentTools ...tool.BaseTool) (adk.Agent, error) {
	safeDir := common.WorkspaceDir()

	return common.NewAgentBuilder("coordinator", "核心工程智能体，直接完成常规工程任务，并按需指派专业成员。").
		WithTemplate(coordinatorPromptTemplate).
		WithTemplateVar("workspace_dir", safeDir).
		WithToolNames("todo", "file", "scheduler", "ask").
		WithTools(agentTools...).
		WithSummary().
		WithSkills().
		Build(ctx)
}
