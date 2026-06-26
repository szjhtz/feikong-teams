package tools

import (
	"fmt"
	"path/filepath"
	"sync"

	eventlog "fkteams/internal/adapters/storage/file/history"
	commandtool "fkteams/internal/adapters/tools/builtin/command"
	doctool "fkteams/internal/adapters/tools/builtin/doc"
	exceltool "fkteams/internal/adapters/tools/builtin/excel"
	fetchtool "fkteams/internal/adapters/tools/builtin/fetch"
	gittool "fkteams/internal/adapters/tools/builtin/git"
	schedulertool "fkteams/internal/adapters/tools/builtin/scheduler"
	buntool "fkteams/internal/adapters/tools/builtin/script/bun"
	uvtool "fkteams/internal/adapters/tools/builtin/script/uv"
	searchtool "fkteams/internal/adapters/tools/builtin/search"
	sshtool "fkteams/internal/adapters/tools/builtin/ssh"
	mcpadapter "fkteams/internal/adapters/tools/mcp"
	"fkteams/internal/app/appdata"
	"fkteams/internal/app/config"
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
	_ = RegisterDefaults()
}

func runtimeDir() string {
	return filepath.Join(appdata.Dir(), "runtime")
}

// RegisterDefaults 将工具适配器连接到应用工具注册表。
func RegisterDefaults() error {
	registerOnce.Do(func() {
		attachment.SetSessionMessageReader(eventlog.NewSessionMessageReader(appdata.SessionsDir(), eventlog.GlobalSessionManager))
		apptools.RegisterMCPProvider(mcpadapter.DefaultProvider())
		if err := apptools.RegisterToolGroup(apptools.ToolGroupRegistration{
			Info: apptools.ToolGroupInfo{
				Name:          "command",
				DisplayName:   "命令执行",
				Description:   "在工作区内执行 shell 命令，适合构建、测试、检查和自动化脚本。",
				Category:      "开发",
				Builtin:       true,
				IncludedTools: []string{"execute"},
			},
			Factory: commandToolGroup(commandtool.ApprovalModeHITL),
		}); err != nil {
			registerErr = err
			return
		}
		if err := apptools.RegisterToolGroup(apptools.ToolGroupRegistration{
			Info: apptools.ToolGroupInfo{
				Name:          "command_reject",
				DisplayName:   "命令执行（自动拒绝危险操作）",
				Description:   "后台任务内部使用的命令执行工具，遇到高风险命令时自动拒绝。",
				Category:      "内部",
				Builtin:       true,
				IncludedTools: []string{"execute"},
				Hidden:        true,
			},
			Factory: commandToolGroup(commandtool.ApprovalModeReject),
		}); err != nil {
			registerErr = err
			return
		}
		if err := apptools.RegisterToolGroup(apptools.ToolGroupRegistration{
			Info: apptools.ToolGroupInfo{
				Name:          "uv",
				DisplayName:   "Python",
				Description:   "使用 uv 创建隔离 Python 环境、安装依赖并执行 Python 代码。",
				Category:      "开发",
				Builtin:       true,
				IncludedTools: []string{"uv_python"},
			},
			Factory: func(*resources.Cleaner) ([]runtimeport.Tool, error) {
				uvTools, err := uvtool.NewUVTools(runtimeDir(), appdata.WorkspaceDir())
				if err != nil {
					return nil, fmt.Errorf("初始化 uv 工具失败: %w", err)
				}
				return uvTools.GetTools()
			},
		}); err != nil {
			registerErr = err
			return
		}
		if err := apptools.RegisterToolGroup(apptools.ToolGroupRegistration{
			Info: apptools.ToolGroupInfo{
				Name:          "bun",
				DisplayName:   "JavaScript",
				Description:   "使用 bun 创建隔离 JavaScript 环境、安装依赖并执行 JS/TS 代码。",
				Category:      "开发",
				Builtin:       true,
				IncludedTools: []string{"bun_javascript"},
			},
			Factory: func(*resources.Cleaner) ([]runtimeport.Tool, error) {
				bunTools, err := buntool.NewBunTools(runtimeDir(), appdata.WorkspaceDir())
				if err != nil {
					return nil, fmt.Errorf("初始化 bun 工具失败: %w", err)
				}
				return bunTools.GetTools()
			},
		}); err != nil {
			registerErr = err
			return
		}
		if err := apptools.RegisterToolGroup(apptools.ToolGroupRegistration{
			Info: apptools.ToolGroupInfo{
				Name:          "excel",
				DisplayName:   "Excel",
				Description:   "创建、读取和编辑 Excel 工作簿，处理表格数据、公式、样式和工作表。",
				Category:      "数据",
				Builtin:       true,
				IncludedTools: []string{"excel_create", "excel_read", "excel_write", "excel_style"},
			},
			Factory: func(*resources.Cleaner) ([]runtimeport.Tool, error) {
				excelTools, err := exceltool.NewExcelTools(appdata.WorkspaceDir())
				if err != nil {
					return nil, fmt.Errorf("初始化Excel工具失败: %w", err)
				}
				return excelTools.GetTools()
			},
		}); err != nil {
			registerErr = err
			return
		}
		if err := apptools.RegisterToolGroup(apptools.ToolGroupRegistration{
			Info: apptools.ToolGroupInfo{
				Name:          "search",
				DisplayName:   "网络搜索",
				Description:   "检索互联网信息，适合需要时效性、外部资料或交叉验证的问题。",
				Category:      "研究",
				Builtin:       true,
				IncludedTools: []string{"search"},
			},
			Factory: func(*resources.Cleaner) ([]runtimeport.Tool, error) {
				return searchtool.GetTools()
			},
		}); err != nil {
			registerErr = err
			return
		}
		if err := apptools.RegisterToolGroup(apptools.ToolGroupRegistration{
			Info: apptools.ToolGroupInfo{
				Name:          "fetch",
				DisplayName:   "网页抓取",
				Description:   "读取指定 URL 的网页内容，适合打开搜索结果、文档页面和公开资料。",
				Category:      "研究",
				Builtin:       true,
				IncludedTools: []string{"fetch"},
			},
			Factory: func(*resources.Cleaner) ([]runtimeport.Tool, error) {
				return fetchtool.GetTools()
			},
		}); err != nil {
			registerErr = err
			return
		}
		if err := apptools.RegisterToolGroup(apptools.ToolGroupRegistration{
			Info: apptools.ToolGroupInfo{
				Name:          "doc",
				DisplayName:   "文档",
				Description:   "读取和分析文档文件，支持文档信息、智能读取、按页和按行读取。",
				Category:      "文档",
				Builtin:       true,
				IncludedTools: []string{"doc_info", "doc_smart_read", "doc_read_pages", "doc_read_lines"},
			},
			Factory: func(*resources.Cleaner) ([]runtimeport.Tool, error) {
				return doctool.GetTools()
			},
		}); err != nil {
			registerErr = err
			return
		}
		if err := apptools.RegisterToolGroup(apptools.ToolGroupRegistration{
			Info: apptools.ToolGroupInfo{
				Name:          "git",
				DisplayName:   "Git",
				Description:   "查看状态、提交、分支、日志和差异，适合版本管理和代码变更检查。",
				Category:      "开发",
				Builtin:       true,
				IncludedTools: []string{"git_status", "git_diff", "git_log", "git_commit"},
			},
			Factory: func(*resources.Cleaner) ([]runtimeport.Tool, error) {
				gitTools, err := gittool.NewGitTools(appdata.WorkspaceDir())
				if err != nil {
					return nil, fmt.Errorf("初始化Git工具失败: %w", err)
				}
				return gitTools.GetTools()
			},
		}); err != nil {
			registerErr = err
			return
		}
		if err := apptools.RegisterToolGroup(apptools.ToolGroupRegistration{
			Info: apptools.ToolGroupInfo{
				Name:          "ssh",
				DisplayName:   "SSH",
				Description:   "连接远程服务器执行命令、上传下载文件和查看目录，需要先配置 SSH 连接信息。",
				Category:      "运维",
				Builtin:       true,
				IncludedTools: []string{"ssh_execute", "ssh_upload", "ssh_download", "ssh_list_dir"},
			},
			Factory: func(cleaner *resources.Cleaner) ([]runtimeport.Tool, error) {
				sshCfg := config.Get().Agents.SSHVisitor
				host := sshCfg.Host
				username := sshCfg.Username
				password := sshCfg.Password
				if host == "" || username == "" || password == "" {
					return nil, fmt.Errorf("SSH 连接信息未配置，请在配置文件 [agents.ssh_visitor] 中设置 host, username, password")
				}
				sshTools, err := sshtool.NewSSHTools(host, username, password)
				if err != nil {
					return nil, fmt.Errorf("初始化 SSH 工具失败: %w", err)
				}
				if cleaner != nil {
					cleaner.Add(func() error {
						sshTools.Close()
						return nil
					})
				}
				return sshTools.GetTools()
			},
		}); err != nil {
			registerErr = err
			return
		}
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

func commandToolGroup(mode commandtool.ApprovalMode) apptools.ToolGroupFactory {
	return func(cleaner *resources.Cleaner) ([]runtimeport.Tool, error) {
		if cleaner != nil {
			cleaner.Add(func() error {
				commandtool.TerminateAll()
				commandtool.CleanupTempFiles(appdata.WorkspaceDir())
				return nil
			})
		}
		return commandtool.NewCommandTools(appdata.WorkspaceDir(), commandtool.WithApprovalMode(mode)).GetTools()
	}
}
