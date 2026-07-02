package assistant

import (
	"context"
	"fkteams/internal/app/agent/catalog/common"

	runtimeport "fkteams/internal/ports/runtime"
)

func DefaultDefinition() common.Definition {
	return common.Definition{
		Name:           "generalist",
		Description:    "通用执行助手，负责综合命令、文件、搜索和文档工具完成开放任务。",
		Instruction:    assistantPrompt,
		Profile:        common.ProfileFull,
		ToolNames:      []string{"command", "file", "search", "fetch", "ask", "doc"},
		DispatchConfig: &runtimeport.DispatchConfig{},
	}
}

func NewAgent(ctx context.Context) (runtimeport.Agent, error) {
	return common.BuildAgent(ctx, DefaultDefinition())
}
