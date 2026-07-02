package eino

import (
	"context"
	"fkteams/internal/adapters/runtime/eino/middlewares/fkfs"
	"fkteams/internal/app/appdata"
	runtimeport "fkteams/internal/ports/runtime"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func NewDeepAgent(ctx context.Context, cfg *runtimeport.DeepAgentConfig) (runtimeport.Agent, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	chatModel, err := AdaptChatModelForRunner(cfg.Model)
	if err != nil {
		return nil, fmt.Errorf("adapt chat model: %w", err)
	}
	runnerTools, err := AdaptToolsForRunner(ctx, cfg.Tools)
	if err != nil {
		return nil, fmt.Errorf("adapt tools: %w", err)
	}
	runnerSubAgents, err := AdaptAgentsForRunner(cfg.SubAgents)
	if err != nil {
		return nil, fmt.Errorf("adapt sub agents: %w", err)
	}
	runnerHandlers, err := AdaptAgentMiddlewaresForRunner(cfg.Middlewares)
	if err != nil {
		return nil, fmt.Errorf("adapt middleware: %w", err)
	}

	deepCfg := &deep.Config{
		Name:                         cfg.Name,
		Description:                  cfg.Description,
		Instruction:                  cfg.Instruction,
		ChatModel:                    chatModel,
		ModelRetryConfig:             AdaptModelRetryConfigForRunner(cfg.ModelRetryConfig),
		SubAgents:                    runnerSubAgents,
		MaxIteration:                 cfg.MaxIterations,
		WithoutWriteTodos:            !cfg.Planning.Enabled,
		WithoutGeneralSubAgent:       !cfg.Delegation.GeneralAgent,
		TaskToolDescriptionGenerator: taskToolDescriptionGenerator(cfg.Delegation.TaskToolDescription),
		Middlewares:                  nil,
		Handlers:                     runnerHandlers,
		OutputKey:                    cfg.Output.Key,
		ToolsConfig: adk.ToolsConfig{
			EmitInternalEvents: cfg.EmitInternalEvents,
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: runnerTools,
			},
		},
	}
	if cfg.Workspace.Enabled {
		backend, err := fkfs.NewLocalBackend(appdata.WorkspaceDir())
		if err != nil {
			return nil, fmt.Errorf("init deep workspace backend: %w", err)
		}
		deepCfg.Backend = backend
	}
	if cfg.Shell.Enabled {
		deepCfg.Shell = fkfs.NewLocalShell(appdata.WorkspaceDir(), cfg.Shell.Timeout)
	}

	agent, err := deep.New(ctx, deepCfg)
	if err != nil {
		return nil, err
	}
	return WrapNamedAgent(cfg.Name, cfg.Description, agent), nil
}

func taskToolDescriptionGenerator(description string) func(context.Context, []adk.TypedAgent[*schema.Message]) (string, error) {
	if description == "" {
		return nil
	}
	return func(context.Context, []adk.TypedAgent[*schema.Message]) (string, error) {
		return description, nil
	}
}
