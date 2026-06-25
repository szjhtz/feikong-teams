package tools

import (
	"context"

	runtimeport "fkteams/internal/ports/runtime"
)

type MCPToolGroup struct {
	Name  string
	Desc  string
	Tools []runtimeport.Tool
}

type MCPToolGroups map[string]MCPToolGroup

type MCPProvider interface {
	GetToolsByName(ctx context.Context, groupName string) ([]runtimeport.Tool, error)
	GetAllToolGroups(ctx context.Context) (MCPToolGroups, error)
	ClearCache()
}
