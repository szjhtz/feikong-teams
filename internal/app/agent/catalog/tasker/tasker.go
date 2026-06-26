package tasker

import (
	"context"
	"fkteams/internal/app/agent/catalog/common"
	"fkteams/internal/app/appdata"

	runtimeport "fkteams/internal/ports/runtime"
)

func NewAgent(ctx context.Context) (runtimeport.Agent, error) {
	workspaceDir := appdata.WorkspaceDir()

	return common.NewAgentBuilder("tasker", "后台任务执行器，独立完成定时任务中的检索、分析、命令和文件操作。").
		WithInstruction(taskerPrompt).
		WithTemplateVar("workspace_dir", workspaceDir).
		WithToolNames("command_reject", "search", "fetch", "file").
		WithSummary().
		Build(ctx)
}
