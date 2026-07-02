package runtime

import (
	"io"
	"os"
	"regexp"
	"strings"
)

// ReadPipeInput 检测 stdin 是否为管道并读取内容。
func ReadPipeInput() (content string, isPipe bool) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", false
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return "", false
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", true
	}
	return strings.TrimSpace(string(data)), true
}

// ExtractAgentMention 提取输入中的智能体 @ 提及。
func ExtractAgentMention(input string) (agentName string, query string) {
	input = strings.TrimSpace(input)
	re := regexp.MustCompile(`^@([\p{Han}\w]+)\s*(.*)$`)
	matches := re.FindStringSubmatch(input)
	if len(matches) == 3 {
		return matches[1], strings.TrimSpace(matches[2])
	}
	return "", input
}

// WorkMode 工作模式
type WorkMode string

const (
	ModeTeam  WorkMode = "team"
	ModeDeep  WorkMode = "deep"
	ModeGroup WorkMode = "group"
)

// String 返回模式字符串
func (m WorkMode) String() string {
	return string(m)
}

// GetPromptPrefix 获取提示符前缀
func (m WorkMode) GetPromptPrefix() string {
	switch m {
	case ModeTeam:
		return "团队模式> "
	case ModeDeep:
		return "深度模式> "
	case ModeGroup:
		return "多智能体讨论模式> "
	default:
		return "未知模式> "
	}
}

// ParseWorkMode 解析工作模式
func ParseWorkMode(mode string) WorkMode {
	switch mode {
	case "team":
		return ModeTeam
	case "deep":
		return ModeDeep
	case "group":
		return ModeGroup
	default:
		return ModeTeam
	}
}
