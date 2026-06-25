package memory

import (
	"context"
	"fkteams/agentcore"
	eventlog "fkteams/internal/adapters/storage/file/history"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestManagerFlushExtractPersistsAndLoadsMarkdown(t *testing.T) {
	workspace := t.TempDir()
	llm := &fakeLLMClient{response: `[{"type":"preference","summary":"偏好中文","detail":"回复使用中文且简洁","tags":["中文","简洁"]}]`}
	manager := NewManager(workspace, llm, nil)

	manager.FlushExtract(context.Background(), longConversationMessages(), "session-1")

	if manager.Count() != 1 {
		t.Fatalf("count = %d, want 1", manager.Count())
	}
	if llm.calls != 1 {
		t.Fatalf("llm calls = %d, want 1", llm.calls)
	}
	if _, err := os.Stat(filepath.Join(workspace, "memory", "preference.md")); err != nil {
		t.Fatalf("preference.md should exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(workspace, "memory", "MEMORY.md")); err != nil {
		t.Fatalf("MEMORY.md should exist: %v", err)
	}

	reloaded := NewManager(workspace, nil, nil)
	entries := reloaded.List()
	if len(entries) != 1 || entries[0].Summary != "偏好中文" || entries[0].Type != Preference {
		t.Fatalf("reloaded entries = %#v", entries)
	}
}

func TestManagerFlushExtractSkipsShortContent(t *testing.T) {
	llm := &fakeLLMClient{response: `[]`}
	manager := NewManager(t.TempDir(), llm, nil)

	manager.FlushExtract(context.Background(), []Message{{Role: "user", Content: "短内容"}}, "session-1")

	if llm.calls != 0 {
		t.Fatalf("llm calls = %d, want 0", llm.calls)
	}
	if manager.Count() != 0 {
		t.Fatalf("count = %d, want 0", manager.Count())
	}
}

func TestManagerListDeleteAndClear(t *testing.T) {
	manager := NewManager(t.TempDir(), nil, nil)
	manager.entries = []MemoryEntry{
		{ID: "1", Type: Preference, Summary: "偏好中文", Detail: "中文回复", CreatedAt: time.Now()},
		{ID: "2", Type: Fact, Summary: "Go 工程师", Detail: "熟悉 Go", CreatedAt: time.Now()},
	}
	manager.rebuildIndex()

	list := manager.List()
	list[0].Summary = "外部修改"
	if manager.List()[0].Summary != "偏好中文" {
		t.Fatal("List should return a copy")
	}

	if deleted := manager.Delete(" 偏好中文 "); deleted != 1 {
		t.Fatalf("deleted = %d, want 1", deleted)
	}
	if manager.Count() != 1 {
		t.Fatalf("count after delete = %d, want 1", manager.Count())
	}
	if deleted := manager.Delete("missing"); deleted != 0 {
		t.Fatalf("deleted missing = %d, want 0", deleted)
	}

	manager.Clear()
	if manager.Count() != 0 {
		t.Fatalf("count after clear = %d, want 0", manager.Count())
	}
	if _, err := os.Stat(filepath.Join(manager.storeDir, "MEMORY.md")); !os.IsNotExist(err) {
		t.Fatalf("MEMORY.md should be removed after clear, err=%v", err)
	}
}

func TestManagerShouldExtract(t *testing.T) {
	manager := NewManager(t.TempDir(), nil, nil)
	longMessages := []Message{
		{Role: "user", Content: strings.Repeat("用户偏好", 80)},
		{Role: "assistant", Content: "收到"},
		{Role: "user", Content: strings.Repeat("继续补充", 80)},
	}

	if !manager.shouldExtract(longMessages, time.Time{}) {
		t.Fatal("expected long conversation with two user messages to extract")
	}
	if manager.shouldExtract(longMessages[:2], time.Time{}) {
		t.Fatal("expected single user message to skip extraction")
	}
	if manager.shouldExtract(longMessages, time.Now()) {
		t.Fatal("expected cooldown to skip extraction")
	}
	if manager.shouldExtract([]Message{{Role: "user", Content: "短"}, {Role: "user", Content: "短"}}, time.Time{}) {
		t.Fatal("expected short content to skip extraction")
	}
}

func TestManagerDuplicateDetection(t *testing.T) {
	manager := NewManager(t.TempDir(), nil, nil)
	manager.entries = []MemoryEntry{{
		Type:    Preference,
		Summary: "偏好简洁回复",
		Detail:  "少说废话",
		Tags:    []string{"风格", "简洁", "中文"},
	}}

	if action, _ := manager.checkDuplicate(MemoryEntry{Type: Preference, Summary: "偏好简洁回复"}); action != actionSkip {
		t.Fatalf("same summary action = %v, want skip", action)
	}
	if action, _ := manager.checkDuplicate(MemoryEntry{Type: Preference, Summary: "用户偏好简洁回复"}); action != actionUpdate {
		t.Fatalf("similar summary action = %v, want update", action)
	}
	if action, _ := manager.checkDuplicate(MemoryEntry{Type: Preference, Summary: "其他", Tags: []string{"风格", "简洁"}}); action != actionUpdate {
		t.Fatalf("overlap tags action = %v, want update", action)
	}
	if action, _ := manager.checkDuplicate(MemoryEntry{Type: Fact, Summary: "偏好简洁回复"}); action != actionAdd {
		t.Fatalf("different type action = %v, want add", action)
	}
}

func TestConvertRecorderMessages(t *testing.T) {
	recorder := eventlog.NewHistoryRecorder()
	recorder.RecordUserMessage(agentcore.Message{Role: agentcore.RoleUser, Content: "用户消息"})
	recorder.RecordEvent(eventlog.Event{Type: eventlog.EventMessageDelta, AgentName: "assistant", Content: "助手回复"})
	recorder.FinalizeCurrent()

	messages := ConvertRecorderMessages(recorder)
	if len(messages) != 2 {
		t.Fatalf("messages = %#v, want 2", messages)
	}
	if messages[0].Role != "user" || messages[0].Content != "用户消息" {
		t.Fatalf("user message = %#v", messages[0])
	}
	if messages[1].Role != "assistant" || messages[1].Content != "助手回复" {
		t.Fatalf("assistant message = %#v", messages[1])
	}
}
