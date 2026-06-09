package scheduler

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newTestScheduler(t *testing.T) *Scheduler {
	t.Helper()
	s, err := newScheduler(t.TempDir())
	if err != nil {
		t.Fatalf("newScheduler failed: %v", err)
	}
	return s
}

func TestScheduleAddListCancelDelete(t *testing.T) {
	s := newTestScheduler(t)
	ctx := context.Background()

	addResp, err := s.ScheduleAdd(ctx, &ScheduleAddRequest{
		Task:      "生成日报",
		ExecuteAt: time.Now().Add(time.Hour).Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("ScheduleAdd returned error: %v", err)
	}
	if !addResp.Success || addResp.Task == nil {
		t.Fatalf("ScheduleAdd failed: %#v", addResp)
	}
	taskID := addResp.Task.ID

	listResp, err := s.ScheduleList(ctx, &ScheduleListRequest{StatusFilter: "pending"})
	if err != nil {
		t.Fatalf("ScheduleList returned error: %v", err)
	}
	if !listResp.Success || listResp.TotalCount != 1 || listResp.Tasks[0].ID != taskID {
		t.Fatalf("ScheduleList = %#v, want one pending task %s", listResp, taskID)
	}

	cancelResp, err := s.ScheduleCancel(ctx, &ScheduleCancelRequest{TaskID: taskID})
	if err != nil {
		t.Fatalf("ScheduleCancel returned error: %v", err)
	}
	if !cancelResp.Success {
		t.Fatalf("ScheduleCancel failed: %#v", cancelResp)
	}

	cancelled, err := s.GetTasks("cancelled")
	if err != nil {
		t.Fatalf("GetTasks returned error: %v", err)
	}
	if len(cancelled) != 1 || cancelled[0].ID != taskID {
		t.Fatalf("cancelled tasks = %#v, want %s", cancelled, taskID)
	}

	if err := os.MkdirAll(s.taskDir(taskID), 0755); err != nil {
		t.Fatalf("create task dir: %v", err)
	}
	deleteResp, err := s.ScheduleDelete(ctx, &ScheduleDeleteRequest{TaskID: taskID})
	if err != nil {
		t.Fatalf("ScheduleDelete returned error: %v", err)
	}
	if !deleteResp.Success {
		t.Fatalf("ScheduleDelete failed: %#v", deleteResp)
	}
	if _, err := os.Stat(s.taskDir(taskID)); !os.IsNotExist(err) {
		t.Fatalf("task dir still exists or unexpected stat error: %v", err)
	}
}

func TestScheduleAddValidation(t *testing.T) {
	s := newTestScheduler(t)
	ctx := context.Background()

	tests := []struct {
		name string
		req  ScheduleAddRequest
		want string
	}{
		{name: "missing task", req: ScheduleAddRequest{ExecuteAt: time.Now().Add(time.Hour).Format(time.RFC3339)}, want: "task description is required"},
		{name: "missing schedule", req: ScheduleAddRequest{Task: "do work"}, want: "must provide"},
		{name: "mutually exclusive", req: ScheduleAddRequest{Task: "do work", CronExpr: "* * * * *", ExecuteAt: time.Now().Add(time.Hour).Format(time.RFC3339)}, want: "mutually exclusive"},
		{name: "invalid cron", req: ScheduleAddRequest{Task: "do work", CronExpr: "bad cron"}, want: "invalid cron expression"},
		{name: "past time", req: ScheduleAddRequest{Task: "do work", ExecuteAt: time.Now().Add(-time.Hour).Format(time.RFC3339)}, want: "must be in the future"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.ScheduleAdd(ctx, &tt.req)
			if err != nil {
				t.Fatalf("ScheduleAdd returned error: %v", err)
			}
			if resp.Success {
				t.Fatalf("Success = true, want false")
			}
			if !strings.Contains(resp.ErrorMessage, tt.want) {
				t.Fatalf("ErrorMessage = %q, want containing %q", resp.ErrorMessage, tt.want)
			}
		})
	}
}

func TestHistoryAndResultReaders(t *testing.T) {
	s := newTestScheduler(t)
	taskID := "task-history"
	resultPath := s.taskResultPath(taskID)
	historyDir := filepath.Join(s.taskDir(taskID), "history")
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		t.Fatalf("create history dir: %v", err)
	}
	if err := os.WriteFile(resultPath, []byte("latest result"), 0644); err != nil {
		t.Fatalf("write result: %v", err)
	}
	if err := os.WriteFile(filepath.Join(historyDir, "20260430_150405.md"), []byte("history result"), 0644); err != nil {
		t.Fatalf("write history: %v", err)
	}
	if err := os.WriteFile(filepath.Join(historyDir, "ignore.txt"), []byte("ignored"), 0644); err != nil {
		t.Fatalf("write ignored file: %v", err)
	}

	if err := s.saveTasks(&ScheduledTaskList{Tasks: []ScheduledTask{{
		ID:         taskID,
		Task:       "history task",
		Status:     "completed",
		CreatedAt:  time.Now(),
		OneTime:    true,
		ResultPath: resultPath,
	}}}); err != nil {
		t.Fatalf("save tasks: %v", err)
	}

	result, err := s.ReadTaskResult(taskID)
	if err != nil {
		t.Fatalf("ReadTaskResult returned error: %v", err)
	}
	if result != "latest result" {
		t.Fatalf("result = %q, want latest result", result)
	}

	entries, err := s.ListHistoryEntries(taskID)
	if err != nil {
		t.Fatalf("ListHistoryEntries returned error: %v", err)
	}
	if len(entries) != 1 || entries[0].Filename != "20260430_150405.md" || entries[0].Time != "2026-04-30 15:04:05" {
		t.Fatalf("entries = %#v", entries)
	}

	content, err := s.ReadHistoryFile(taskID, "../20260430_150405.md")
	if err != nil {
		t.Fatalf("ReadHistoryFile returned error: %v", err)
	}
	if content != "history result" {
		t.Fatalf("history content = %q, want history result", content)
	}
}

func TestFormatTasksForDisplay(t *testing.T) {
	now := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	got := FormatTasksForDisplay([]ScheduledTask{{
		ID:        "task-1",
		Task:      "整理测试",
		Status:    "pending",
		CronExpr:  "0 9 * * *",
		NextRunAt: now,
	}})

	for _, want := range []string{"1 scheduled tasks", "整理测试", "task-1", "0 9 * * *"} {
		if !strings.Contains(got, want) {
			t.Fatalf("display = %q, want containing %q", got, want)
		}
	}
}
