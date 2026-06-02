package tui

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mattn/go-isatty"
)

type ToolEvent struct {
	Key     string
	Name    string
	Type    string
	Content string
	Append  bool
}

type toolDoneMsg struct{}
type toolAutoExitMsg struct{}

type toolItem struct {
	key    string
	name   string
	status string
	args   string
	result string
	error  string
}

type toolModel struct {
	items    []toolItem
	indexes  map[string]int
	cursor   int
	expanded int
	scrollY  int
	width    int
	done     bool
}

var (
	toolTitleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	toolDimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	toolDoneStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	toolErrorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	toolActiveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
)

func newToolModel() toolModel {
	return toolModel{
		indexes:  make(map[string]int),
		expanded: -1,
		width:    80,
	}
}

func (m toolModel) Init() tea.Cmd { return nil }

func (m toolModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case ToolEvent:
		m.applyEvent(msg)
		return m, nil
	case toolDoneMsg:
		m.done = true
		return m, tea.Tick(1500*time.Millisecond, func(time.Time) tea.Msg { return toolAutoExitMsg{} })
	case toolAutoExitMsg:
		return m, tea.Quit
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m toolModel) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		if m.done {
			return m, tea.Quit
		}
	case "enter":
		if m.expanded >= 0 {
			m.expanded = -1
			m.scrollY = 0
		} else if len(m.items) > 0 {
			m.expanded = m.cursor
			m.scrollY = 0
		}
	case "esc":
		if m.expanded >= 0 {
			m.expanded = -1
			m.scrollY = 0
		} else if m.done {
			return m, tea.Quit
		}
	case "up", "k":
		if m.expanded >= 0 {
			if m.scrollY > 0 {
				m.scrollY--
			}
		} else if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.expanded >= 0 {
			m.scrollY++
		} else if m.cursor < len(m.items)-1 {
			m.cursor++
		}
	}
	return m, nil
}

func (m *toolModel) applyEvent(e ToolEvent) {
	if e.Key == "" {
		return
	}
	i, ok := m.indexes[e.Key]
	if !ok {
		i = len(m.items)
		m.indexes[e.Key] = i
		name := e.Name
		if name == "" {
			name = e.Key
		}
		m.items = append(m.items, toolItem{key: e.Key, name: name, status: "准备参数"})
	}
	item := &m.items[i]
	if e.Name != "" {
		item.name = e.Name
	}
	switch e.Type {
	case "start":
		item.status = "准备参数"
	case "args":
		item.status = "已调用"
		if e.Append {
			item.args += e.Content
		} else {
			item.args = e.Content
		}
	case "result":
		item.status = "执行中"
		if e.Append {
			item.result += e.Content
		} else {
			item.result = e.Content
		}
	case "done":
		item.status = "已完成"
		if e.Content != "" {
			if e.Append {
				item.result += e.Content
			} else {
				item.result = e.Content
			}
		}
	case "error":
		item.status = "失败"
		item.error = e.Content
	}
}

func toolIcon(status string) string {
	switch status {
	case "已完成":
		return "✓"
	case "失败":
		return "✗"
	case "已调用", "执行中":
		return "◐"
	default:
		return "○"
	}
}

func toolStatusStyle(status string) lipgloss.Style {
	switch status {
	case "已完成":
		return toolDoneStyle
	case "失败":
		return toolErrorStyle
	case "已调用", "执行中":
		return toolActiveStyle
	default:
		return toolDimStyle
	}
}

func (m toolModel) View() tea.View {
	var b strings.Builder
	if len(m.items) == 0 {
		b.WriteString(toolDimStyle.Render("等待工具调用..."))
		b.WriteString("\n")
		return tea.NewView(b.String())
	}
	for i := range m.items {
		if i == m.expanded {
			b.WriteString(m.renderExpanded(i))
		} else {
			b.WriteString(m.renderCollapsed(i))
		}
		b.WriteString("\n")
	}
	if !m.done {
		if m.expanded >= 0 {
			b.WriteString(toolDimStyle.Render("↑↓ 滚动  Enter/Esc 收起"))
		} else {
			b.WriteString(toolDimStyle.Render("↑↓ 选择  Enter 展开"))
		}
		b.WriteString("\n")
	}
	return tea.NewView(b.String())
}

func (m toolModel) renderCollapsed(i int) string {
	item := m.items[i]
	prefix := "  "
	if !m.done && i == m.cursor {
		prefix = "▶ "
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s%s %s", prefix, toolStatusStyle(item.status).Render(toolIcon(item.status)), item.name)
	b.WriteString(toolDimStyle.Render("  " + item.status))
	if item.args != "" {
		b.WriteString("\n")
		b.WriteString(toolDimStyle.Render("  参数: " + truncateRunes(compactLine(item.args), 120)))
	}
	if item.status == "已完成" && item.result != "" {
		b.WriteString("\n")
		b.WriteString(toolDimStyle.Render("  结果: " + truncateRunes(compactLine(item.result), 180)))
	}
	if item.error != "" {
		b.WriteString("\n")
		b.WriteString(toolErrorStyle.Render("  " + truncateRunes(compactLine(item.error), 120)))
	}
	return b.String()
}

func (m toolModel) renderExpanded(i int) string {
	item := m.items[i]
	var b strings.Builder
	b.WriteString(fmt.Sprintf("▼ %s %s  %s\n", toolStatusStyle(item.status).Render(toolIcon(item.status)), item.name, toolDimStyle.Render(item.status)))
	if item.args != "" {
		b.WriteString("\n")
		b.WriteString(toolDimStyle.Render("参数:") + "\n")
		b.WriteString(wrapBlock(item.args, 8))
	}
	if item.result != "" {
		b.WriteString("\n")
		b.WriteString(toolDimStyle.Render("结果:") + "\n")
		lines := strings.Split(strings.TrimSpace(item.result), "\n")
		start := m.scrollY
		if start >= len(lines) {
			start = max(0, len(lines)-1)
		}
		end := min(start+12, len(lines))
		for _, line := range lines[start:end] {
			b.WriteString("  " + truncateRunes(line, 160) + "\n")
		}
		if end < len(lines) {
			b.WriteString(toolDimStyle.Render(fmt.Sprintf("  ... 还有 %d 行", len(lines)-end)) + "\n")
		}
	}
	if item.error != "" {
		b.WriteString("\n")
		b.WriteString(toolErrorStyle.Render("错误: "+item.error) + "\n")
	}
	return b.String()
}

func wrapBlock(text string, maxLines int) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	var b strings.Builder
	for i, line := range lines {
		if i >= maxLines {
			b.WriteString(toolDimStyle.Render(fmt.Sprintf("  ... 还有 %d 行", len(lines)-maxLines)) + "\n")
			break
		}
		b.WriteString("  " + truncateRunes(line, 160) + "\n")
	}
	return b.String()
}

func truncateRunes(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

func compactLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

type ToolPanel struct {
	mu      sync.Mutex
	program *tea.Program
	done    chan struct{}
	active  bool
	enabled bool
}

func NewToolPanel() *ToolPanel {
	return &ToolPanel{enabled: isatty.IsTerminal(os.Stdout.Fd())}
}

func (p *ToolPanel) Send(e ToolEvent) bool {
	if !p.enabled {
		return false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.active {
		p.program = tea.NewProgram(newToolModel())
		p.done = make(chan struct{})
		p.active = true
		go func(program *tea.Program, done chan struct{}) {
			_, _ = program.Run()
			close(done)
		}(p.program, p.done)
	}
	p.program.Send(e)
	return true
}

func (p *ToolPanel) Finish() {
	p.mu.Lock()
	if !p.active || p.program == nil {
		p.mu.Unlock()
		return
	}
	program := p.program
	done := p.done
	p.active = false
	p.program = nil
	p.done = nil
	p.mu.Unlock()
	program.Send(toolDoneMsg{})
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		program.Quit()
		<-done
	}
}
