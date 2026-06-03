package fkevent

import (
	"os"
	"strings"
	"sync"

	glamour "charm.land/glamour/v2"
	"charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"
	classiclipgloss "github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"golang.org/x/term"
)

var (
	mdRenderer     *glamour.TermRenderer
	mdRendererOnce sync.Once
)

func termWidth() int {
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	return 100
}

func sp(s string) *string { return &s }
func bp(b bool) *bool     { return &b }
func up(u uint) *uint     { return &u }

// customDarkStyle 基于 DarkStyleConfig 全面定制配色
func customDarkStyle() glamour.TermRendererOption {
	s := styles.DarkStyleConfig

	// ── 文档 ──
	s.Document.Color = sp("#e0e0e0")

	// ── 标题 ──
	s.Heading.Color = sp("#5c9cf5")
	s.Heading.Bold = bp(true)
	// H1: 去掉背景色块，统一蓝色粗体
	s.H1 = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{
		Prefix: "# ", Color: sp("#5c9cf5"), Bold: bp(true),
	}}
	s.H2.Color = sp("#5c9cf5")
	s.H3.Color = sp("#5c9cf5")
	s.H4.Color = sp("#5c9cf5")
	s.H5.Color = sp("#5c9cf5")
	s.H6 = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{
		Prefix: "###### ", Color: sp("#5c9cf5"), Bold: bp(false),
	}}

	// ── 行内样式 ──
	s.Strong = ansi.StylePrimitive{Bold: bp(true), Color: sp("#9d7cd8")}
	s.Emph = ansi.StylePrimitive{Italic: bp(true), Color: sp("#e5c07b")}
	s.Strikethrough = ansi.StylePrimitive{CrossedOut: bp(true), Color: sp("#6a6a6a")}

	// ── 行内代码 ──
	s.Code = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{
		Color: sp("#7fd88f"),
	}}

	// ── 引用块 ──
	s.BlockQuote = ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{Italic: bp(true), Color: sp("#e5c07b")},
		Indent:         up(1),
		IndentToken:    sp("┃ "),
	}

	// ── 列表 ──
	// Prefix（而非 BlockPrefix）才会应用 Item 自身的颜色
	s.Item = ansi.StylePrimitive{Prefix: "• ", Color: sp("#fab283")}
	s.Enumeration = ansi.StylePrimitive{BlockPrefix: ". ", Color: sp("#56b6c2")}
	s.Task = ansi.StyleTask{Ticked: "[✓] ", Unticked: "[ ] "}

	// ── 链接 ──
	s.Link = ansi.StylePrimitive{Color: sp("#fab283"), Underline: bp(true)}
	s.LinkText = ansi.StylePrimitive{Color: sp("#56b6c2"), Bold: bp(true)}

	// ── 图片 ──
	s.Image = ansi.StylePrimitive{Color: sp("#fab283"), Underline: bp(true)}
	s.ImageText = ansi.StylePrimitive{Color: sp("#56b6c2"), Format: "Image: {{.text}} →"}

	// ── 分割线 ──
	s.HorizontalRule = ansi.StylePrimitive{
		Color:  sp("#6a6a6a"),
		Format: "\n──────────────────────────────────────────\n",
	}

	// ── 表格：Unicode 分隔符 ──
	s.Table = ansi.StyleTable{
		CenterSeparator: sp("┼"),
		ColumnSeparator: sp("│"),
		RowSeparator:    sp("─"),
	}

	// ── 定义列表 ──
	s.DefinitionDescription = ansi.StylePrimitive{BlockPrefix: "\n❯ "}

	// ── 代码块 + 语法高亮 ──
	s.CodeBlock = ansi.StyleCodeBlock{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{Color: sp("#e0e0e0")},
			Margin:         up(1),
		},
		Chroma: &ansi.Chroma{
			Text:                ansi.StylePrimitive{Color: sp("#e0e0e0")},
			Error:               ansi.StylePrimitive{Color: sp("#e06c75"), BackgroundColor: sp("#3c2020")},
			Comment:             ansi.StylePrimitive{Color: sp("#6a6a6a")},
			CommentPreproc:      ansi.StylePrimitive{Color: sp("#fab283")},
			Keyword:             ansi.StylePrimitive{Color: sp("#5c9cf5")},
			KeywordReserved:     ansi.StylePrimitive{Color: sp("#9d7cd8")},
			KeywordNamespace:    ansi.StylePrimitive{Color: sp("#e06c75")},
			KeywordType:         ansi.StylePrimitive{Color: sp("#e5c07b")},
			Operator:            ansi.StylePrimitive{Color: sp("#56b6c2")},
			Punctuation:         ansi.StylePrimitive{Color: sp("#abb2bf")},
			Name:                ansi.StylePrimitive{Color: sp("#e0e0e0")},
			NameBuiltin:         ansi.StylePrimitive{Color: sp("#56b6c2")},
			NameTag:             ansi.StylePrimitive{Color: sp("#e06c75")},
			NameAttribute:       ansi.StylePrimitive{Color: sp("#e5c07b")},
			NameClass:           ansi.StylePrimitive{Color: sp("#e5c07b"), Bold: bp(true), Underline: bp(true)},
			NameConstant:        ansi.StylePrimitive{Color: sp("#fab283")},
			NameDecorator:       ansi.StylePrimitive{Color: sp("#e5c07b")},
			NameFunction:        ansi.StylePrimitive{Color: sp("#fab283")},
			LiteralNumber:       ansi.StylePrimitive{Color: sp("#9d7cd8")},
			LiteralString:       ansi.StylePrimitive{Color: sp("#7fd88f")},
			LiteralStringEscape: ansi.StylePrimitive{Color: sp("#56b6c2")},
			GenericDeleted:      ansi.StylePrimitive{Color: sp("#e06c75")},
			GenericEmph:         ansi.StylePrimitive{Italic: bp(true)},
			GenericInserted:     ansi.StylePrimitive{Color: sp("#7fd88f")},
			GenericStrong:       ansi.StylePrimitive{Bold: bp(true)},
			GenericSubheading:   ansi.StylePrimitive{Color: sp("#6a6a6a")},
			Background:          ansi.StylePrimitive{BackgroundColor: sp("#2d2d2d")},
		},
	}

	return glamour.WithStyles(s)
}

func initRenderer() *glamour.TermRenderer {
	mdRendererOnce.Do(func() {
		w := termWidth() - 4
		if w < 40 {
			w = 40
		}
		r, err := glamour.NewTermRenderer(
			customDarkStyle(),
			glamour.WithWordWrap(w),
			glamour.WithEmoji(),
			glamour.WithChromaFormatter("terminal16m"),
		)
		if err != nil {
			r, _ = glamour.NewTermRenderer(
				customDarkStyle(),
				glamour.WithWordWrap(w),
				glamour.WithChromaFormatter("terminal16m"),
			)
		}
		mdRenderer = r
	})
	return mdRenderer
}

// RenderMarkdown 渲染 Markdown 为 ANSI 输出，失败时返回原文
func RenderMarkdown(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}
	content = strings.ReplaceAll(content, "[^", `\[^`)
	if segments := splitMarkdownTables(content); hasMarkdownTableSegment(segments) {
		return renderMarkdownSegments(segments)
	}
	r := initRenderer()
	if r == nil {
		return content
	}
	out, err := r.Render(content)
	if err != nil {
		return content
	}
	return strings.Trim(out, "\n")
}

type markdownSegment struct {
	text  string
	table bool
}

func hasMarkdownTableSegment(segments []markdownSegment) bool {
	for _, seg := range segments {
		if seg.table {
			return true
		}
	}
	return false
}

func renderMarkdownSegments(segments []markdownSegment) string {
	var rendered []string
	for _, seg := range segments {
		text := strings.TrimSpace(seg.text)
		if text == "" {
			continue
		}
		if seg.table {
			rendered = append(rendered, renderMarkdownTable(text))
			continue
		}
		r := initRenderer()
		if r == nil {
			rendered = append(rendered, text)
			continue
		}
		out, err := r.Render(text)
		if err != nil {
			rendered = append(rendered, text)
			continue
		}
		rendered = append(rendered, strings.Trim(out, "\n"))
	}
	return strings.Join(rendered, "\n\n")
}

func splitMarkdownTables(content string) []markdownSegment {
	lines := strings.Split(content, "\n")
	var segments []markdownSegment
	var normal []string
	inFence := false

	flushNormal := func() {
		text := strings.TrimSpace(strings.Join(normal, "\n"))
		if text != "" {
			segments = append(segments, markdownSegment{text: text})
		}
		normal = nil
	}

	for i := 0; i < len(lines); {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			normal = append(normal, lines[i])
			i++
			continue
		}
		if !inFence && i+1 < len(lines) && isMarkdownTableRow(lines[i]) && isMarkdownTableSeparator(lines[i+1]) {
			flushNormal()
			start := i
			i += 2
			for i < len(lines) && isMarkdownTableRow(lines[i]) && strings.TrimSpace(lines[i]) != "" {
				i++
			}
			segments = append(segments, markdownSegment{
				text:  strings.Join(lines[start:i], "\n"),
				table: true,
			})
			continue
		}
		normal = append(normal, lines[i])
		i++
	}
	flushNormal()
	return segments
}

func isMarkdownTableRow(line string) bool {
	return strings.Contains(line, "|")
}

func isMarkdownTableSeparator(line string) bool {
	cells := splitMarkdownTableLine(line)
	if len(cells) < 2 {
		return false
	}
	for _, cell := range cells {
		cell = strings.TrimSpace(cell)
		if strings.Count(cell, "-") < 3 {
			return false
		}
		for _, r := range cell {
			if r != '-' && r != ':' {
				return false
			}
		}
	}
	return true
}

func splitMarkdownTableLine(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")

	var cells []string
	var b strings.Builder
	escaped := false
	for _, r := range line {
		if escaped {
			b.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if r == '|' {
			cells = append(cells, strings.TrimSpace(b.String()))
			b.Reset()
			continue
		}
		b.WriteRune(r)
	}
	cells = append(cells, strings.TrimSpace(b.String()))
	return cells
}

type tableAlign int

const (
	tableAlignLeft tableAlign = iota
	tableAlignCenter
	tableAlignRight
)

func parseTableAligns(separator []string, count int) []tableAlign {
	aligns := make([]tableAlign, count)
	for i := 0; i < count && i < len(separator); i++ {
		cell := strings.TrimSpace(separator[i])
		left := strings.HasPrefix(cell, ":")
		right := strings.HasSuffix(cell, ":")
		switch {
		case left && right:
			aligns[i] = tableAlignCenter
		case right:
			aligns[i] = tableAlignRight
		default:
			aligns[i] = tableAlignLeft
		}
	}
	return aligns
}

func renderMarkdownTable(tableMarkdown string) string {
	lines := strings.Split(strings.TrimSpace(tableMarkdown), "\n")
	if len(lines) < 2 {
		return tableMarkdown
	}

	rows := make([][]string, 0, len(lines)-1)
	header := splitMarkdownTableLine(lines[0])
	colCount := len(header)
	for _, line := range lines[2:] {
		cells := splitMarkdownTableLine(line)
		if len(cells) > colCount {
			colCount = len(cells)
		}
		rows = append(rows, cells)
	}
	normalizeTableRow(&header, colCount)
	for i := range rows {
		normalizeTableRow(&rows[i], colCount)
	}
	aligns := parseTableAligns(splitMarkdownTableLine(lines[1]), colCount)

	t := table.New().
		Border(classiclipgloss.NormalBorder()).
		BorderStyle(classiclipgloss.NewStyle().Foreground(classiclipgloss.Color("8"))).
		BorderTop(true).
		BorderBottom(true).
		BorderLeft(true).
		BorderRight(true).
		BorderHeader(true).
		BorderColumn(true).
		BorderRow(false).
		Wrap(true).
		Headers(header...).
		Rows(rows...).
		StyleFunc(func(row, col int) classiclipgloss.Style {
			style := classiclipgloss.NewStyle().Padding(0, 1)
			if row == table.HeaderRow {
				return style.Bold(true).Foreground(classiclipgloss.Color("12")).Align(classiclipgloss.Center)
			}
			if col < len(aligns) {
				return style.Align(tableAlignPosition(aligns[col]))
			}
			return style
		})

	rendered := t.String()
	maxWidth := termWidth() - 4
	if maxWidth < 40 {
		maxWidth = 40
	}
	if classiclipgloss.Width(rendered) > maxWidth {
		rendered = t.Width(maxWidth).String()
	}
	return strings.TrimRight(rendered, "\n")
}

func normalizeTableRow(row *[]string, count int) {
	for len(*row) < count {
		*row = append(*row, "")
	}
	if len(*row) > count {
		*row = (*row)[:count]
	}
}

func tableAlignPosition(align tableAlign) classiclipgloss.Position {
	switch align {
	case tableAlignRight:
		return classiclipgloss.Right
	case tableAlignCenter:
		return classiclipgloss.Center
	default:
		return classiclipgloss.Left
	}
}
