package cli

// CommandInfo 命令信息。
type CommandInfo struct {
	Name string
	Desc string
}

// allCommands 是 TUI runtime 命令补全列表。
var allCommands = []CommandInfo{
	{"help", "帮助信息"},
	{"list_agents", "列出所有可用的智能体"},
	{"list_chat_history", "列出所有聊天历史会话"},
	{"load_chat_history", "选择并加载聊天历史会话"},
	{"save_chat_history", "保存聊天历史到当前会话文件"},
	{"clear_chat_history", "清空当前聊天历史"},
	{"switch_work_mode", "切换工作模式(团队/深度/讨论/自定义)"},
	{"save_chat_history_to_html", "导出聊天历史为 HTML 文件"},
	{"save_chat_history_to_markdown", "导出聊天历史为 Markdown 文件"},
	{"list_schedule", "列出所有定时任务"},
	{"cancel_schedule", "选择并取消定时任务"},
	{"delete_schedule", "选择并删除定时任务"},
	{"list_memory", "列出所有长期记忆条目"},
	{"delete_memory", "选择并删除记忆条目"},
	{"clear_memory", "清空所有长期记忆"},
	{"quit", "退出程序"},
}
