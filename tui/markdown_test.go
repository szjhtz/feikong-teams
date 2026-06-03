package tui

import (
	"strings"
	"testing"
)

func TestRenderMarkdownTableUsesClosedBorder(t *testing.T) {
	out := RenderMarkdown("| 项目 | 状态 |\n| --- | --- |\n| 表格 | 完成 |")

	for _, token := range []string{"┌", "┐", "└", "┘", "│", "├", "┤"} {
		if !strings.Contains(out, token) {
			t.Fatalf("expected closed table border token %q in output:\n%s", token, out)
		}
	}
}

func TestRenderMarkdownMixedContentKeepsClosedTableBorder(t *testing.T) {
	out := RenderMarkdown("## 测试\n\n| 项目 | 状态 |\n| --- | --- |\n| 表格 | 完成 |")

	if !strings.Contains(out, "测试") {
		t.Fatalf("expected heading content in output:\n%s", out)
	}
	for _, token := range []string{"┌", "┐", "└", "┘"} {
		if !strings.Contains(out, token) {
			t.Fatalf("expected table border token %q in output:\n%s", token, out)
		}
	}
}

func TestRenderMarkdownCodeBlockUsesBackground(t *testing.T) {
	out := RenderMarkdown("```go\npackage main\n```")

	if !strings.Contains(out, "package") || !strings.Contains(out, "main") {
		t.Fatalf("expected code content in output:\n%s", out)
	}
	if !strings.Contains(out, "48;2;31;35;41") {
		t.Fatalf("expected code block background color in output:\n%q", out)
	}
	lines := strings.Split(out, "\n")
	if len(lines) < 3 {
		t.Fatalf("expected padded code block area, got:\n%q", out)
	}
	for _, line := range lines {
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, codeBlockBackgroundANSI) {
			t.Fatalf("expected every code block row to start with background color, got line:\n%q\nfull output:\n%q", line, out)
		}
	}
}

func TestRenderMarkdownCodeBlockExpandsTabsAndControlChars(t *testing.T) {
	out := RenderMarkdown("```go\nfunc main() {\n\tfmt.Println(\"hi\")\n\n}\x00\n```")

	if strings.Contains(out, "\t") || strings.Contains(out, "\x00") {
		t.Fatalf("expected tabs and control chars to be normalized:\n%q", out)
	}
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, codeBlockBackgroundANSI) {
			t.Fatalf("expected every code block row to keep background, got line:\n%q\nfull output:\n%q", line, out)
		}
	}
}

func TestNormalizeCodeBlockTextExpandsTabsByColumns(t *testing.T) {
	got := normalizeCodeBlockText("a\tb\n\tc")
	want := "a   b\n    c"

	if got != want {
		t.Fatalf("normalizeCodeBlockText() = %q, want %q", got, want)
	}
}
