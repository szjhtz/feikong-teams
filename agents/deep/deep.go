package deep

import (
	"context"
	"fkteams/agentcore"
	einoruntime "fkteams/agentcore/eino"
	"fkteams/agents/common"
	"fkteams/agents/middlewares/summary"
	rootcommon "fkteams/common"
	"fkteams/fkenv"
	"fkteams/tools"
	"fmt"
	"strconv"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/compose"
)

func NewAgent(ctx context.Context, subAgents []agentcore.Agent) (agentcore.Agent, error) {

	toolNames := []string{"file", "doc", "command", "search", "fetch"}
	var toolList []agentcore.Tool
	for _, toolName := range toolNames {
		baseTools, err := tools.GetToolsByName(toolName)
		if err != nil {
			return nil, fmt.Errorf("init tool %s: %w", toolName, err)
		}
		toolList = append(toolList, baseTools...)
	}
	runnerTools, err := einoruntime.AdaptToolsForRunner(ctx, toolList)
	if err != nil {
		return nil, fmt.Errorf("adapt tools: %w", err)
	}
	runnerSubAgents, err := einoruntime.AdaptAgentsForRunner(subAgents)
	if err != nil {
		return nil, fmt.Errorf("adapt sub agents: %w", err)
	}

	chatModel, err := common.NewChatModel()
	if err != nil {
		return nil, fmt.Errorf("create chat model: %w", err)
	}
	runnerModel, err := einoruntime.AdaptChatModelForRunner(chatModel)
	if err != nil {
		return nil, fmt.Errorf("adapt chat model: %w", err)
	}

	maxTokens := summary.DefaultMaxTokensBeforeSummary
	if v := fkenv.Get(fkenv.MaxTokensBeforeSummary); v != "" {
		if n, _ := strconv.Atoi(v); n > 0 {
			maxTokens = n
		}
	}
	summaryMiddleware, err := summary.New(ctx, &summary.Config{
		Model:                  chatModel,
		MaxTokensBeforeSummary: maxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("init summary middleware: %w", err)
	}
	runnerSummaryMiddleware, err := einoruntime.AdaptAgentMiddlewareForRunner(summaryMiddleware)
	if err != nil {
		return nil, fmt.Errorf("adapt summary middleware: %w", err)
	}

	agent, err := deep.New(ctx, &deep.Config{
		Name:             "deep_researcher",
		Description:      "深度研究智能体，负责深入分析问题并协调多个成员解决复杂任务。",
		ChatModel:        runnerModel,
		ModelRetryConfig: rootcommon.NewModelRetryConfig(),
		SubAgents:        runnerSubAgents,
		MaxIteration:     common.MaxIterations(),
		Handlers:         []adk.ChatModelAgentMiddleware{runnerSummaryMiddleware},
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: runnerTools,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return einoruntime.WrapNamedAgent("deep_researcher", "深度研究智能体，负责深入分析问题并协调多个成员解决复杂任务。", agent), nil
}
