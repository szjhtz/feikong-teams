package researcher

import (
	"context"
	"fkteams/agents/common"

	"fkteams/agentcore"
)

func NewAgent(ctx context.Context) (agentcore.Agent, error) {
	return common.NewAgentBuilder("researcher", "网络研究员，负责检索、抓取、交叉验证和整理时效信息。").
		WithInstruction(researcherPrompt).
		WithToolNames("search", "fetch").
		WithSummary().
		Build(ctx)
}
