package tui

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"charm.land/lipgloss/v2"
	"github.com/mattn/go-isatty"
)

type MemberEvent struct {
	Key      string
	NewKey   string
	Name     string
	Type     string
	Content  string
	ToolKey  string
	ToolName string
	Append   bool
}

var (
	memberColorGreen  = lipgloss.Color("2")
	memberColorRed    = lipgloss.Color("1")
	memberColorYellow = lipgloss.Color("3")
	memberColorDim    = lipgloss.Color("8")
	memberDimStyle    = lipgloss.NewStyle().Foreground(memberColorDim)
)

type memberCard struct {
	key        string
	name       string
	status     string
	operations []string
	content    string
	tools      []memberToolFlow
	toolIndex  map[string]int
}

type memberToolFlow struct {
	key    string
	name   string
	status string
	args   string
	result string
}

type memberModel struct {
	members   []memberCard
	indexes   map[string]int
	width     int
	done      bool
	emptyText string
}

func newMemberModel(emptyText string) memberModel {
	return memberModel{
		indexes:   make(map[string]int),
		width:     80,
		emptyText: emptyText,
	}
}

func (m *memberModel) applyEvent(e MemberEvent) {
	if e.Key == "" {
		return
	}
	if e.Type == "rename" && e.NewKey != "" {
		if i, ok := m.indexes[e.Key]; ok {
			if existing, exists := m.indexes[e.NewKey]; exists && existing != i {
				dst := &m.members[existing]
				src := m.members[i]
				if e.Name != "" {
					dst.name = e.Name
				}
				if dst.status == "waiting" || src.status == "error" || (src.status == "done" && dst.status != "error") {
					dst.status = src.status
				}
				dst.operations = append(dst.operations, src.operations...)
				if dst.toolIndex == nil {
					dst.toolIndex = make(map[string]int)
				}
				for _, tool := range src.tools {
					dstTool := dst.ensureTool(tool.key, tool.name)
					dstTool.status = tool.status
					dstTool.args += tool.args
					dstTool.result += tool.result
				}
				if src.content != "" {
					dst.content += src.content
				}
				m.members = append(m.members[:i], m.members[i+1:]...)
				delete(m.indexes, e.Key)
				for key, idx := range m.indexes {
					if idx > i {
						m.indexes[key] = idx - 1
					}
				}
				return
			}
			delete(m.indexes, e.Key)
			m.indexes[e.NewKey] = i
			m.members[i].key = e.NewKey
			if e.Name != "" {
				m.members[i].name = e.Name
			}
		}
		return
	}
	i, ok := m.indexes[e.Key]
	if !ok {
		i = len(m.members)
		m.indexes[e.Key] = i
		name := e.Name
		if name == "" {
			name = e.Key
		}
		m.members = append(m.members, memberCard{key: e.Key, name: name, status: "waiting", toolIndex: make(map[string]int)})
	}

	card := &m.members[i]
	if card.toolIndex == nil {
		card.toolIndex = make(map[string]int)
	}
	if e.Name != "" {
		card.name = e.Name
	}
	switch e.Type {
	case "start":
		card.status = "running"
	case "op":
		if e.Content != "" {
			card.operations = append(card.operations, e.Content)
		}
	case "tool_prepare":
		tool := card.ensureTool(e.ToolKey, e.ToolName)
		tool.status = "参数准备中"
	case "tool_args":
		tool := card.ensureTool(e.ToolKey, e.ToolName)
		tool.status = "已调用"
		if e.Append {
			tool.args += e.Content
		} else {
			tool.args = e.Content
		}
	case "tool_result":
		tool := card.ensureTool(e.ToolKey, e.ToolName)
		tool.status = "已完成"
		if e.Append {
			tool.result += e.Content
		} else {
			tool.result = e.Content
		}
	case "content":
		card.content += e.Content
	case "done":
		card.status = "done"
		if e.Content != "" {
			card.content += e.Content
		}
	case "error":
		card.status = "error"
		if e.Content != "" {
			if card.content != "" {
				card.content += "\n"
			}
			card.content += "错误: " + e.Content
		}
	}
}

func (c *memberCard) ensureTool(key, name string) *memberToolFlow {
	if key == "" {
		key = name
	}
	if key == "" {
		key = fmt.Sprintf("tool:%d", len(c.tools)+1)
	}
	if name == "" {
		name = key
	}
	if c.toolIndex == nil {
		c.toolIndex = make(map[string]int)
	}
	if i, ok := c.toolIndex[key]; ok {
		if name != "" {
			c.tools[i].name = name
		}
		return &c.tools[i]
	}
	c.tools = append(c.tools, memberToolFlow{key: key, name: name, status: "参数准备中"})
	c.toolIndex[key] = len(c.tools) - 1
	return &c.tools[len(c.tools)-1]
}

func memberStatusIcon(status string) string {
	switch status {
	case "waiting":
		return "○"
	case "running":
		return "◐"
	case "done":
		return "✓"
	case "error":
		return "✗"
	default:
		return "?"
	}
}

func memberStatusColor(status string) lipgloss.Style {
	switch status {
	case "done":
		return lipgloss.NewStyle().Foreground(memberColorGreen)
	case "error":
		return lipgloss.NewStyle().Foreground(memberColorRed)
	case "running":
		return lipgloss.NewStyle().Foreground(memberColorYellow)
	default:
		return lipgloss.NewStyle().Foreground(memberColorDim)
	}
}

func (m memberModel) View() string {
	var b strings.Builder
	w := m.width
	if w < 40 {
		w = 80
	}

	if len(m.members) == 0 {
		emptyText := m.emptyText
		if emptyText == "" {
			emptyText = "等待事件..."
		}
		b.WriteString(memberDimStyle.Render("  " + emptyText))
		b.WriteString("\n")
		return b.String()
	}

	for i := range m.members {
		b.WriteString(m.renderCard(i, w))
		b.WriteString("\n")
	}

	return b.String()
}

func (m memberModel) renderCard(i, w int) string {
	card := m.members[i]
	lineWidth := max(20, w-4)

	icon := memberStatusIcon(card.status)
	var body strings.Builder
	body.WriteString(fmt.Sprintf("%s %s", memberStatusColor(card.status).Render(icon), truncateRunes(card.name, max(12, lineWidth-4))))

	if card.content != "" {
		preview := memberCompactLine(card.content)
		if preview != "" {
			body.WriteString("\n")
			body.WriteString(memberDimStyle.Render("输出: " + truncateRunes(preview, max(20, lineWidth-8))))
		}
	}

	if len(card.tools) > 0 || len(card.operations) > 0 {
		if chain := memberToolChain(card.tools, w); chain != "" {
			body.WriteString("\n")
			body.WriteString(memberDimStyle.Render("工具链: " + chain))
		}
		if current := currentMemberTool(card.tools); current != "" {
			body.WriteString("\n")
			body.WriteString(memberDimStyle.Render("正在: " + current))
		}
		if op := latestMemberOperation(card.operations); op != "" {
			body.WriteString("\n")
			body.WriteString(memberDimStyle.Render(truncateRunes(op, lineWidth)))
		}
		for _, line := range memberToolPreview(card.tools, w) {
			body.WriteString("\n")
			body.WriteString(line)
		}
	}

	return body.String()
}

func memberCompactLine(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	return strings.Join(strings.Fields(s), " ")
}

func latestMemberOperation(ops []string) string {
	for i := len(ops) - 1; i >= 0; i-- {
		if op := memberCompactLine(ops[i]); op != "" {
			return op
		}
	}
	return ""
}

func memberToolChain(tools []memberToolFlow, w int) string {
	if len(tools) == 0 {
		return ""
	}
	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		name := tool.name
		if name == "" {
			name = tool.key
		}
		if tool.status != "" {
			name = fmt.Sprintf("%s(%s)", name, tool.status)
		}
		names = append(names, name)
	}
	return truncateRunes(strings.Join(names, " → "), max(20, w-12))
}

func currentMemberTool(tools []memberToolFlow) string {
	for i := len(tools) - 1; i >= 0; i-- {
		tool := tools[i]
		if tool.status == "已完成" {
			continue
		}
		name := tool.name
		if name == "" {
			name = tool.key
		}
		if tool.status == "" {
			return name
		}
		return fmt.Sprintf("%s [%s]", name, tool.status)
	}
	return ""
}

func memberToolPreview(tools []memberToolFlow, w int) []string {
	if len(tools) == 0 {
		return nil
	}
	start := max(0, len(tools)-3)
	lines := make([]string, 0, (len(tools)-start)*3+1)
	if start > 0 {
		lines = append(lines, memberDimStyle.Render(fmt.Sprintf("... 已省略 %d 个较早工具", start)))
	}
	for _, tool := range tools[start:] {
		name := tool.name
		if name == "" {
			name = tool.key
		}
		status := tool.status
		if status == "" {
			status = "进行中"
		}
		lines = append(lines, memberDimStyle.Render(fmt.Sprintf("▸ %s [%s]", name, status)))
		if arg := memberCompactLine(tool.args); arg != "" && tool.result == "" {
			lines = append(lines, memberDimStyle.Render("  参数: "+truncateRunes(arg, max(20, w-12))))
		}
		if result := memberCompactLine(tool.result); result != "" {
			lines = append(lines, memberDimStyle.Render("  结果: "+truncateRunes(result, max(20, w-12))))
		}
	}
	return lines
}

type MemberPanel struct {
	mu        sync.Mutex
	model     memberModel
	active    bool
	enabled   bool
	emptyText string
	lastLines int
}

func NewMemberPanel() *MemberPanel {
	return &MemberPanel{enabled: isatty.IsTerminal(os.Stdout.Fd()), emptyText: "等待子智能体启动..."}
}

func (p *MemberPanel) Send(e MemberEvent) bool {
	if !p.enabled {
		return false
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.active {
		p.model = newMemberModel(p.emptyText)
		p.active = true
	}
	p.model.applyEvent(e)
	p.renderLocked()
	return true
}

func (p *MemberPanel) Finish() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.active {
		return
	}
	p.model.done = true
	p.renderLocked()
	p.active = false
	p.lastLines = 0
}

func (p *MemberPanel) renderLocked() {
	p.model.width = terminalWidth()
	if p.lastLines > 0 {
		fmt.Printf("\033[%dF\033[J", p.lastLines)
	}
	view := p.model.View()
	fmt.Print(view)
	p.lastLines = renderedLineCount(view, p.model.width)
}
