package memory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSaveAndLoadAllMarkdownRoundTrip(t *testing.T) {
	dir := t.TempDir()
	lastHit := time.Date(2026, 1, 2, 3, 4, 5, 0, time.Local)
	created := time.Date(2026, 1, 1, 2, 3, 4, 0, time.Local)
	entries := []MemoryEntry{
		{Type: Preference, Summary: "偏好中文", Detail: "回答使用中文", Tags: []string{"中文", "风格"}, CreatedAt: created, HitCount: 3, LastHitAt: &lastHit},
		{Type: Lesson, Summary: "先跑测试", Detail: "提交前完整验证", Tags: []string{"测试"}, CreatedAt: created, HitCount: 1},
	}

	if err := saveAllMarkdown(dir, entries); err != nil {
		t.Fatalf("saveAllMarkdown: %v", err)
	}

	index, err := os.ReadFile(filepath.Join(dir, "MEMORY.md"))
	if err != nil {
		t.Fatalf("read MEMORY.md: %v", err)
	}
	if !strings.Contains(string(index), "用户偏好") || !strings.Contains(string(index), "避坑记录") {
		t.Fatalf("index content = %s", string(index))
	}

	loaded := loadAllMarkdown(dir)
	if len(loaded) != 2 {
		t.Fatalf("loaded = %#v, want 2 entries", loaded)
	}
	if loaded[0].Summary != "偏好中文" || loaded[0].HitCount != 3 || loaded[0].LastHitAt == nil {
		t.Fatalf("loaded preference = %#v", loaded[0])
	}
	if loaded[1].Summary != "先跑测试" || loaded[1].Type != Lesson {
		t.Fatalf("loaded lesson = %#v", loaded[1])
	}
}

func TestSaveAllMarkdownRemovesEmptyTypeFilesAndIndex(t *testing.T) {
	dir := t.TempDir()
	if err := saveAllMarkdown(dir, []MemoryEntry{{
		Type:      Preference,
		Summary:   "偏好中文",
		Detail:    "回答使用中文",
		CreatedAt: time.Now(),
	}}); err != nil {
		t.Fatalf("save non-empty: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "preference.md")); err != nil {
		t.Fatalf("preference.md should exist: %v", err)
	}

	if err := saveAllMarkdown(dir, nil); err != nil {
		t.Fatalf("save empty: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "preference.md")); !os.IsNotExist(err) {
		t.Fatalf("preference.md should be removed, err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "MEMORY.md")); !os.IsNotExist(err) {
		t.Fatalf("MEMORY.md should be removed, err=%v", err)
	}
}

func TestBuildMemoryContextGroupsKnownTypes(t *testing.T) {
	context := BuildMemoryContext([]MemoryEntry{
		{Type: Preference, Summary: "偏好中文", Detail: "回答使用中文"},
		{Type: MemoryType("unknown"), Summary: "UNKNOWN_SENTINEL", Detail: "未知类型"},
	})

	if !strings.Contains(context, "## 长期记忆") || !strings.Contains(context, "### 用户偏好") {
		t.Fatalf("context = %q", context)
	}
	if !strings.Contains(context, "偏好中文") || strings.Contains(context, "UNKNOWN_SENTINEL") {
		t.Fatalf("context should include known entries only: %q", context)
	}
	if BuildMemoryContext(nil) != "" {
		t.Fatal("empty entries should produce empty context")
	}
	if BuildMemoryContext([]MemoryEntry{{Type: MemoryType("unknown"), Summary: "忽略"}}) != "" {
		t.Fatal("unknown-only entries should produce empty context")
	}
}
