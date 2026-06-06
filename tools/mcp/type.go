package mcp

import "fkteams/agentcore"

type ToolGroup struct {
	Name  string
	Desc  string
	Tools []agentcore.Tool
}

type DictToolGroup map[string]ToolGroup
