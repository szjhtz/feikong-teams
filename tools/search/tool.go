package search

import (
	"context"
	"fkteams/agentcore"
)

func GetTools() (tools []agentcore.Tool, err error) {
	duckduckgoTool, err := NewDuckDuckGoTool(context.Background())
	if err != nil {
		return nil, err
	}
	tools = append(tools, duckduckgoTool)
	return tools, nil
}
