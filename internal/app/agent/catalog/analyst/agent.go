package analyst

import (
	"context"
	"fkteams/internal/app/agent/catalog/common"

	runtimeport "fkteams/internal/ports/runtime"
)

func DefaultDefinition() common.Definition {
	return common.Definition{
		Name:        "analyst",
		Description: "数据分析师，负责使用表格、脚本和文档工具提取洞察。",
		Instruction: analystPrompt,
		Profile:     common.ProfileFull,
		ToolNames:   []string{"todo", "excel", "file", "uv", "doc"},
	}
}

func NewAgent(ctx context.Context) (runtimeport.Agent, error) {
	return common.BuildAgent(ctx, DefaultDefinition())
}
