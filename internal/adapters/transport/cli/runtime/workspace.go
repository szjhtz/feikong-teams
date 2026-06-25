package runtime

import (
	"fkteams/internal/app/appdata"
)

// GetWorkspaceDir 获取工作目录路径
func GetWorkspaceDir() string {
	return appdata.WorkspaceDir()
}
