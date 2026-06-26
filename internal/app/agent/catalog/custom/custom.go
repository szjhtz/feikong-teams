package custom

import (
	"context"
	"fkteams/internal/app/agent/catalog/common"
	runtimeport "fkteams/internal/ports/runtime"
	modelregistry "fkteams/internal/runtime/model"
	"fmt"
)

type Model struct {
	Provider string
	Name     string
	APIKey   string
	BaseURL  string
}

type Config struct {
	Name         string
	Description  string
	SystemPrompt string
	Model        Model
	ToolNames    []string
	Tools        []runtimeport.Tool
}

func NewAgent(ctx context.Context, cfg Config) (runtimeport.Agent, error) {
	builder := common.NewAgentBuilder(cfg.Name, cfg.Description).
		WithInstruction(cfg.SystemPrompt).
		WithTools(cfg.Tools...).
		WithToolNames(cfg.ToolNames...).
		WithSummary().
		WithSkills()

	if cfg.Model.Name != "" || cfg.Model.BaseURL != "" {
		chatModel, err := common.NewChatModelWithConfig(ctx, &modelregistry.Config{
			Provider: modelregistry.Type(cfg.Model.Provider),
			APIKey:   cfg.Model.APIKey,
			BaseURL:  cfg.Model.BaseURL,
			Model:    cfg.Model.Name,
		})
		if err != nil {
			return nil, fmt.Errorf("create chat model: %w", err)
		}
		builder = builder.WithModel(chatModel)
	}

	return builder.Build(ctx)
}
