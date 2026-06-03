package fkevent

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
