package tui

import (
	"fmt"
	"os"
	"strings"
	"sync"

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

type toolItem struct {
	key    string
	name   string
	status string
	args   string
	result string
	error  string
}

type toolModel struct {
	items   []toolItem
	indexes map[string]int
	width   int
	done    bool
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
		indexes: make(map[string]int),
		width:   80,
	}
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

func (m toolModel) View() string {
	var b strings.Builder
	if len(m.items) == 0 {
		b.WriteString(toolDimStyle.Render("等待工具调用..."))
		b.WriteString("\n")
		return b.String()
	}
	for i := range m.items {
		b.WriteString(m.renderItem(i))
		b.WriteString("\n")
	}
	return b.String()
}

func (m toolModel) renderItem(i int) string {
	item := m.items[i]
	w := max(40, m.width)
	lineWidth := max(20, w-4)
	var b strings.Builder
	name := truncateRunes(item.name, max(12, lineWidth-12))
	fmt.Fprintf(&b, "%s %s", toolStatusStyle(item.status).Render(toolIcon(item.status)), name)
	b.WriteString(toolDimStyle.Render("  " + item.status))
	if item.args != "" {
		b.WriteString("\n")
		b.WriteString(toolDimStyle.Render("  参数: " + truncateRunes(compactLine(item.args), max(20, lineWidth-8))))
	}
	if item.result != "" {
		b.WriteString("\n")
		b.WriteString(toolDimStyle.Render("  结果: " + truncateRunes(compactLine(item.result), max(20, lineWidth-8))))
	}
	if item.error != "" {
		b.WriteString("\n")
		b.WriteString(toolErrorStyle.Render("  " + truncateRunes(compactLine(item.error), lineWidth-2)))
	}
	return b.String()
}

func compactLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

type ToolPanel struct {
	mu        sync.Mutex
	model     toolModel
	active    bool
	enabled   bool
	lastLines int
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
		p.model = newToolModel()
		p.active = true
	}
	p.model.applyEvent(e)
	p.renderLocked()
	return true
}

func (p *ToolPanel) Finish() {
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

func (p *ToolPanel) renderLocked() {
	p.model.width = terminalWidth()
	if p.lastLines > 0 {
		fmt.Printf("\033[%dF\033[J", p.lastLines)
	}
	view := p.model.View()
	fmt.Print(view)
	p.lastLines = renderedLineCount(view, p.model.width)
}
