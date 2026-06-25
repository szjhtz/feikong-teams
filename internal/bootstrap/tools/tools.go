package tools

import (
	"sync"

	eventlog "fkteams/internal/adapters/storage/file/history"
	schedulertool "fkteams/internal/adapters/tools/builtin/scheduler"
	"fkteams/internal/app/appdata"
	apptools "fkteams/internal/app/tools"
	"fkteams/internal/app/tools/attachment"
	runtimeport "fkteams/internal/ports/runtime"
	"fkteams/internal/runtime/resources"
)

var (
	registerOnce sync.Once
	registerErr  error
)

func init() {
	if err := RegisterDefaults(); err != nil {
		panic(err)
	}
}

// RegisterDefaults 将工具适配器连接到应用工具注册表。
func RegisterDefaults() error {
	registerOnce.Do(func() {
		attachment.SetSessionMessageReader(eventlog.NewSessionMessageReader(appdata.SessionsDir(), eventlog.GlobalSessionManager))
		registerErr = apptools.RegisterToolGroup(apptools.ToolGroupRegistration{
			Info: apptools.ToolGroupInfo{
				Name:          "scheduler",
				DisplayName:   "定时任务",
				Description:   "创建、查看和管理自然语言定时任务，适合提醒、周期执行和后台任务。",
				Category:      "自动化",
				Builtin:       true,
				IncludedTools: []string{"schedule_add", "schedule_list", "schedule_cancel", "schedule_delete"},
			},
			Factory: func(*resources.Cleaner) ([]runtimeport.Tool, error) {
				return schedulertool.NewTools(nil).GetTools()
			},
		})
	})
	return registerErr
}
