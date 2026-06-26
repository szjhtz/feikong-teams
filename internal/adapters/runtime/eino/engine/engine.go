package engine

import (
	"context"

	einoruntime "fkteams/internal/adapters/runtime/eino"
	"fkteams/internal/adapters/runtime/eino/middlewares/agentsmd"
	"fkteams/internal/adapters/runtime/eino/middlewares/autocontinue"
	"fkteams/internal/adapters/runtime/eino/middlewares/dispatch"
	"fkteams/internal/adapters/runtime/eino/middlewares/inject"
	"fkteams/internal/adapters/runtime/eino/middlewares/skills"
	"fkteams/internal/adapters/runtime/eino/middlewares/steering"
	"fkteams/internal/adapters/runtime/eino/middlewares/summary"
	"fkteams/internal/adapters/runtime/eino/middlewares/tools/destructiveguard"
	hooktools "fkteams/internal/adapters/runtime/eino/middlewares/tools/hooks"
	"fkteams/internal/adapters/runtime/eino/middlewares/tools/patch"
	"fkteams/internal/adapters/runtime/eino/middlewares/tools/trimresult"
	"fkteams/internal/adapters/runtime/eino/middlewares/tools/warperror"
	runtimeport "fkteams/internal/ports/runtime"

	einoMCP "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/mark3labs/mcp-go/client"
)

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) RuntimeInfo() runtimeport.RuntimeInfo {
	return runtimeport.RuntimeInfo{
		Name:        "eino",
		Description: "CloudWeGo Eino ADK runtime adapter",
		Capabilities: []string{
			"chat_agent",
			"loop_agent",
			"deep_agent",
			"agent_tools",
			"streaming_runner",
			"agent_middlewares",
			"tool_middlewares",
			"mcp_tools",
		},
	}
}

func (e *Engine) CheckHealth(ctx context.Context) runtimeport.RuntimeHealth {
	return runtimeport.RuntimeHealth{
		Name:  e.RuntimeInfo().Name,
		Ready: ctx.Err() == nil,
	}
}

func (e *Engine) NewChatModelAgent(ctx context.Context, cfg *runtimeport.ChatAgentConfig) (runtimeport.Agent, error) {
	return einoruntime.NewChatModelAgent(ctx, cfg)
}

func (e *Engine) NewLoopAgent(ctx context.Context, cfg *runtimeport.LoopAgentConfig) (runtimeport.Agent, error) {
	return einoruntime.NewLoopAgent(ctx, cfg)
}

func (e *Engine) NewDeepAgent(ctx context.Context, cfg *runtimeport.DeepAgentConfig) (runtimeport.Agent, error) {
	return einoruntime.NewDeepAgent(ctx, cfg)
}

func (e *Engine) NewRunner(ctx context.Context, cfg runtimeport.RunnerConfig) (runtimeport.Runner, error) {
	return einoruntime.NewRunnerFromConfig(ctx, cfg)
}

func (e *Engine) NewAgentTools(ctx context.Context, subAgents []runtimeport.Agent, cfg runtimeport.AgentToolConfig) ([]runtimeport.Tool, error) {
	return einoruntime.NewAgentTools(ctx, subAgents, cfg)
}

func (e *Engine) DecorateChatModel(ctx context.Context, chatModel runtimeport.ChatModel) (runtimeport.ChatModel, error) {
	return inject.NewForModel(chatModel)
}

func (e *Engine) DefaultAgentMiddlewares(ctx context.Context) ([]runtimeport.AgentMiddleware, error) {
	result := make([]runtimeport.AgentMiddleware, 0, 5)
	patchMiddleware, err := e.newPatchMiddleware(ctx)
	if err != nil {
		return nil, err
	}
	result = append(result, patchMiddleware)
	result = append(result, e.newToolErrorMiddleware())
	acMiddleware, err := e.newAutoContinueMiddleware()
	if err != nil {
		return nil, err
	}
	result = append(result, acMiddleware)
	result = append(result, e.newTrimResultMiddleware())
	result = append(result, e.NewSteeringMiddleware())
	return result, nil
}

func (e *Engine) DefaultToolMiddlewares() []runtimeport.ToolMiddleware {
	return []runtimeport.ToolMiddleware{
		e.newHookToolMiddleware(),
		e.newDestructiveGuardMiddleware(),
	}
}

func (e *Engine) newPatchMiddleware(ctx context.Context) (runtimeport.AgentMiddleware, error) {
	return patch.New(ctx)
}

func (e *Engine) newToolErrorMiddleware() runtimeport.AgentMiddleware {
	return warperror.NewHandler(nil)
}

func (e *Engine) newAutoContinueMiddleware() (runtimeport.AgentMiddleware, error) {
	return autocontinue.NewHandler()
}

func (e *Engine) newTrimResultMiddleware() runtimeport.AgentMiddleware {
	return trimresult.New(nil)
}

func (e *Engine) NewSteeringMiddleware() runtimeport.AgentMiddleware {
	return steering.New()
}

func (e *Engine) NewSummaryMiddleware(ctx context.Context, cfg *runtimeport.SummaryConfig) (runtimeport.AgentMiddleware, error) {
	if cfg == nil {
		return summary.New(ctx, nil)
	}
	return summary.New(ctx, &summary.Config{
		Model:                  cfg.Model,
		MaxTokensBeforeSummary: cfg.MaxTokensBeforeSummary,
	})
}

func (e *Engine) NewSkillsMiddleware(ctx context.Context) (runtimeport.AgentMiddleware, error) {
	return skills.New(ctx)
}

func (e *Engine) NewDispatchMiddleware(ctx context.Context, cfg *runtimeport.DispatchConfig) (runtimeport.AgentMiddleware, error) {
	if cfg == nil {
		return dispatch.New(ctx, &dispatch.Config{})
	}
	return dispatch.New(ctx, &dispatch.Config{
		Model:          cfg.Model,
		ToolNames:      cfg.ToolNames,
		Tools:          cfg.Tools,
		MaxConcurrency: cfg.MaxConcurrency,
		TaskTimeout:    cfg.TaskTimeout,
	})
}

func (e *Engine) NewAgentsMDMiddleware(ctx context.Context) (runtimeport.AgentMiddleware, error) {
	return agentsmd.New(ctx)
}

func (e *Engine) newDestructiveGuardMiddleware() runtimeport.ToolMiddleware {
	return destructiveguard.New()
}

func (e *Engine) newHookToolMiddleware() runtimeport.ToolMiddleware {
	return hooktools.New()
}

func (e *Engine) MCPTools(ctx context.Context, cli *client.Client) ([]runtimeport.Tool, error) {
	tools, err := einoMCP.GetTools(ctx, &einoMCP.Config{Cli: cli})
	if err != nil {
		return nil, err
	}
	result := make([]runtimeport.Tool, 0, len(tools))
	for _, t := range tools {
		result = append(result, einoruntime.WrapTool(t))
	}
	return result, nil
}
