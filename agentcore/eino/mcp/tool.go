package mcp

import (
	"context"
	"fkteams/agentcore"

	einoMCP "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/mark3labs/mcp-go/client"
)

func GetTools(ctx context.Context, cli *client.Client) ([]agentcore.Tool, error) {
	tools, err := einoMCP.GetTools(ctx, &einoMCP.Config{Cli: cli})
	if err != nil {
		return nil, err
	}
	wrapped := make([]agentcore.Tool, 0, len(tools))
	for _, t := range tools {
		wrapped = append(wrapped, agentcore.WrapRuntimeTool(t))
	}
	return wrapped, nil
}
