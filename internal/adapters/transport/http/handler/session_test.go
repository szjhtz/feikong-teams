package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"fkteams/internal/adapters/storage/file/history"
	"fkteams/internal/domain/message"
	runtimeevents "fkteams/internal/runtime/events"

	"github.com/gin-gonic/gin"
)

func TestGetSessionReturnsEmptyEventsWhenHistoryFileMissing(t *testing.T) {
	rt := newTestRuntime(t)
	gin.SetMode(gin.TestMode)

	sessionID := "empty-session"
	if err := eventlog.SaveMetadata(rt.sessionDirPath(sessionID), &eventlog.SessionMetadata{
		ID:           sessionID,
		Title:        "empty",
		Status:       "idle",
		CurrentAgent: "coder",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}); err != nil {
		t.Fatalf("save metadata: %v", err)
	}

	router := gin.New()
	router.GET("/sessions/:sessionID", rt.GetSessionHandler())

	req := httptest.NewRequest(http.MethodGet, "/sessions/"+sessionID, nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var got Response
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	data, ok := got.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected response data object, got %#v", got.Data)
	}
	if data["session_id"] != sessionID {
		t.Fatalf("unexpected session_id: %#v", data["session_id"])
	}
	if data["current_agent"] != "coder" {
		t.Fatalf("unexpected current_agent: %#v", data["current_agent"])
	}
	gotEvents, ok := data["events"].([]any)
	if !ok {
		t.Fatalf("expected events array, got %#v", data["events"])
	}
	if len(gotEvents) != 0 {
		t.Fatalf("expected empty events, got %#v", gotEvents)
	}
}

func TestGetSessionReturnsEventsInHistoryOrder(t *testing.T) {
	rt := newTestRuntime(t)
	gin.SetMode(gin.TestMode)

	sessionID := "ordered-session"
	sessionDir := rt.sessionDirPath(sessionID)
	if err := eventlog.SaveMetadata(sessionDir, &eventlog.SessionMetadata{
		ID:           sessionID,
		Title:        "ordered",
		Status:       "idle",
		CurrentAgent: "coordinator",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}); err != nil {
		t.Fatalf("save metadata: %v", err)
	}
	recorder := eventlog.NewHistoryRecorder()
	recorder.RecordEvent(runtimeevents.UserMessage("run-1", runtimeevents.TurnID("run-1", 1), "run-1:user", message.Message{Role: message.RoleUser, Content: "你好"}))
	recorder.RecordEvent(eventlog.Event{Type: eventlog.EventAssistantText, AgentName: "coordinator", Content: "你好！", Sequence: 42})
	recorder.RecordEvent(runtimeevents.UserMessage("run-2", runtimeevents.TurnID("run-2", 1), "run-2:user", message.Message{Role: message.RoleUser, Content: "你是谁"}))
	recorder.RecordEvent(eventlog.Event{Type: eventlog.EventAssistantText, AgentName: "coordinator", Content: "我是协调者", Sequence: 84})
	if err := recorder.SaveToFile(filepath.Join(sessionDir, eventlog.HistoryFileName)); err != nil {
		t.Fatalf("save history: %v", err)
	}

	router := gin.New()
	router.GET("/sessions/:sessionID", rt.GetSessionHandler())

	req := httptest.NewRequest(http.MethodGet, "/sessions/"+sessionID, nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	var got Response
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	data, ok := got.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected response data object, got %#v", got.Data)
	}
	gotEvents, ok := data["events"].([]any)
	if !ok {
		t.Fatalf("expected events array, got %#v", data["events"])
	}
	wantTypes := []string{
		string(runtimeevents.EventUserMessage),
		string(runtimeevents.EventAssistantText),
		string(runtimeevents.EventUserMessage),
		string(runtimeevents.EventAssistantText),
	}
	wantContents := []string{"你好", "你好！", "你是谁", "我是协调者"}
	if len(gotEvents) != len(wantTypes) {
		t.Fatalf("expected %d events, got %#v", len(wantTypes), gotEvents)
	}
	for i, raw := range gotEvents {
		event, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("event %d should be object, got %#v", i, raw)
		}
		if event["type"] != wantTypes[i] {
			t.Fatalf("event %d type = %#v, want %q", i, event["type"], wantTypes[i])
		}
		if event["content"] != wantContents[i] {
			t.Fatalf("event %d content = %#v, want %q", i, event["content"], wantContents[i])
		}
		if _, ok := event["sequence"]; !ok {
			t.Fatalf("event %d missing sequence: %#v", i, event)
		}
		if _, ok := event["stream_event_id"]; ok {
			t.Fatalf("event %d should not expose stream_event_id in history: %#v", i, event)
		}
	}
}

func TestHistoryLinesToChatEventsUsesEventSequenceAcrossMembers(t *testing.T) {
	rt := newTestRuntime(t)
	lines := []eventlog.HistoryLine{
		{
			MessageID:  "parent",
			EventIndex: 0,
			AgentName:  "coordinator",
			StartTime:  time.Now(),
			Event:      eventlog.MessageEvent{Type: eventlog.MsgTypeToolCall, Sequence: 10, ToolCall: &eventlog.ToolCallRecord{ID: "call-1", Ref: "tool_call:call-1", Name: "ask_fkagent_researcher"}},
		},
		{
			MessageID:  "parent",
			EventIndex: 1,
			AgentName:  "coordinator",
			StartTime:  time.Now(),
			Event:      eventlog.MessageEvent{Type: eventlog.MsgTypeText, Sequence: 30, Content: "最终回复"},
		},
		{
			MessageID:      "member",
			EventIndex:     0,
			AgentName:      "ask_fkagent_researcher",
			MemberCallID:   "call-1",
			MemberToolName: "ask_fkagent_researcher",
			MemberName:     "researcher",
			StartTime:      time.Now(),
			Event:          eventlog.MessageEvent{Type: eventlog.MsgTypeText, Sequence: 20, Content: "成员结果"},
		},
	}

	got := rt.historyLinesToChatEvents("session-1", lines)
	wantTypes := []string{
		string(runtimeevents.EventToolCallStarted),
		string(runtimeevents.EventAssistantText),
		string(runtimeevents.EventToolCallCompleted),
		string(runtimeevents.EventAssistantText),
	}
	wantContents := []string{"", "成员结果", "", "最终回复"}
	wantSequences := []int64{10, 20, 10, 30}
	if len(got) != len(wantTypes) {
		t.Fatalf("expected %d events, got %#v", len(wantTypes), got)
	}
	for i := range got {
		if fmt.Sprint(got[i]["type"]) != wantTypes[i] {
			t.Fatalf("event %d type = %#v, want %q", i, got[i]["type"], wantTypes[i])
		}
		content := ""
		if raw, ok := got[i]["content"]; ok {
			content = fmt.Sprint(raw)
		}
		if content != wantContents[i] {
			t.Fatalf("event %d content = %q, want %q", i, content, wantContents[i])
		}
		if got[i]["sequence"] != wantSequences[i] {
			t.Fatalf("event %d sequence = %#v, want %d", i, got[i]["sequence"], wantSequences[i])
		}
	}
	if got[1]["is_member_event"] != true {
		t.Fatalf("expected second event to be member event, got %#v", got[1])
	}
}

func TestHistoryLinesToChatEventsDoesNotMixSequencesAcrossTurns(t *testing.T) {
	rt := newTestRuntime(t)
	lines := []eventlog.HistoryLine{
		{
			MessageID: "user-1",
			AgentName: "user",
			StartTime: time.Now(),
			Event:     eventlog.MessageEvent{Type: eventlog.MsgTypeText, Content: "你好"},
		},
		{
			MessageID: "assistant-1",
			AgentName: "coordinator",
			StartTime: time.Now(),
			Event:     eventlog.MessageEvent{Type: eventlog.MsgTypeReasoning, Sequence: 5, Content: "第一轮思考"},
		},
		{
			MessageID: "assistant-1",
			AgentName: "coordinator",
			StartTime: time.Now(),
			Event:     eventlog.MessageEvent{Type: eventlog.MsgTypeText, Sequence: 25, Content: "第一轮回复"},
		},
		{
			MessageID: "user-2",
			AgentName: "user",
			StartTime: time.Now(),
			Event:     eventlog.MessageEvent{Type: eventlog.MsgTypeText, Content: "你好啊"},
		},
		{
			MessageID: "assistant-2",
			AgentName: "coordinator",
			StartTime: time.Now(),
			Event:     eventlog.MessageEvent{Type: eventlog.MsgTypeReasoning, Sequence: 5, Content: "第二轮思考"},
		},
		{
			MessageID: "assistant-2",
			AgentName: "coordinator",
			StartTime: time.Now(),
			Event:     eventlog.MessageEvent{Type: eventlog.MsgTypeText, Sequence: 14, Content: "第二轮回复"},
		},
	}

	got := rt.historyLinesToChatEvents("session-1", lines)
	wantContents := []string{"你好", "第一轮思考", "第一轮回复", "你好啊", "第二轮思考", "第二轮回复"}
	if len(got) != len(wantContents) {
		t.Fatalf("expected %d events, got %#v", len(wantContents), got)
	}
	for i, want := range wantContents {
		if got[i]["content"] != want {
			t.Fatalf("event %d content = %#v, want %q", i, got[i]["content"], want)
		}
	}
}

func TestGetSessionReturnsNotFoundWhenSessionMissing(t *testing.T) {
	rt := newTestRuntime(t)
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/sessions/:sessionID", rt.GetSessionHandler())

	req := httptest.NewRequest(http.MethodGet, "/sessions/missing-session", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", resp.Code, resp.Body.String())
	}
}

func newTestRuntime(t *testing.T) *Runtime {
	t.Helper()

	return NewRuntime(RuntimeOptions{HistoryDir: filepath.Join(t.TempDir(), "sessions")})
}
