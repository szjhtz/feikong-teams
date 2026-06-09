package scheduler

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fkteams/agentcore"
)

func TestGlobalSchedulerAndTimingHelpers(t *testing.T) {
	oldGlobal := globalScheduler
	t.Cleanup(func() {
		globalScheduler = oldGlobal
	})
	globalScheduler = nil

	baseDir := t.TempDir()
	first, err := InitGlobal(baseDir)
	if err != nil {
		t.Fatalf("InitGlobal first: %v", err)
	}
	second, err := InitGlobal(filepath.Join(t.TempDir(), "other"))
	if err != nil {
		t.Fatalf("InitGlobal second: %v", err)
	}
	if first != second || Global() != first {
		t.Fatal("InitGlobal should be idempotent and set global scheduler")
	}

	next, err := first.ComputeNextRun("*/5 * * * *", time.Date(2026, 6, 9, 10, 1, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("ComputeNextRun: %v", err)
	}
	if next.Minute()%5 != 0 || !next.After(time.Date(2026, 6, 9, 10, 1, 0, 0, time.UTC)) {
		t.Fatalf("unexpected next run: %s", next)
	}
	if _, err := first.ComputeNextRun("bad cron", time.Now()); err == nil {
		t.Fatal("expected invalid cron error")
	}

	if err := first.saveTasks(&ScheduledTaskList{Tasks: []ScheduledTask{{
		ID:        "due",
		Task:      "due task",
		Status:    "pending",
		NextRunAt: time.Now().Add(-time.Second),
		CreatedAt: time.Now(),
	}}}); err != nil {
		t.Fatalf("save due task: %v", err)
	}
	if wait := first.nextCheckDuration(); wait != 0 {
		t.Fatalf("nextCheckDuration for due task = %s, want 0", wait)
	}

	if err := first.saveTasks(&ScheduledTaskList{Tasks: []ScheduledTask{{
		ID:        "running",
		Task:      "running task",
		Status:    "running",
		NextRunAt: time.Now(),
		CreatedAt: time.Now(),
	}}}); err != nil {
		t.Fatalf("save running task: %v", err)
	}
	first.recoverStaleRunningTasks()
	tasks, err := first.GetTasks("pending")
	if err != nil {
		t.Fatalf("GetTasks pending: %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != "running" {
		t.Fatalf("recovered tasks = %#v", tasks)
	}
}

func TestSchedulerExecuteTaskAndCleanup(t *testing.T) {
	s := newTestScheduler(t)
	now := time.Now()
	old := now.Add(-taskResultTTL - time.Hour)
	tasks := &ScheduledTaskList{Tasks: []ScheduledTask{
		{ID: "ok", Task: "ok task", Status: "running", OneTime: true, CreatedAt: now},
		{ID: "fail", Task: "fail task", Status: "running", OneTime: true, CreatedAt: now},
		{ID: "old", Task: "old task", Status: "completed", OneTime: true, CreatedAt: old, LastRunAt: &old},
	}}
	if err := s.saveTasks(tasks); err != nil {
		t.Fatalf("save tasks: %v", err)
	}
	if err := os.MkdirAll(s.taskDir("old"), 0755); err != nil {
		t.Fatalf("mkdir old task dir: %v", err)
	}

	s.executeTask("ok", "ok task", "", true, fakeTaskExecutor{})
	s.executeTask("fail", "fail task", "", true, fakeTaskExecutor{err: errors.New("boom")})

	updated, err := s.GetTasks("")
	if err != nil {
		t.Fatalf("GetTasks: %v", err)
	}
	statusByID := map[string]string{}
	for _, task := range updated {
		statusByID[task.ID] = task.Status
		if task.ID == "ok" && task.ResultPath == "" {
			t.Fatalf("ok task result path was not set: %#v", task)
		}
	}
	if statusByID["ok"] != "completed" || statusByID["fail"] != "failed" {
		t.Fatalf("statuses = %#v", statusByID)
	}

	s.cleanupExpiredTasks()
	if _, err := os.Stat(s.taskDir("old")); !os.IsNotExist(err) {
		t.Fatalf("old task dir should be cleaned, stat err=%v", err)
	}
	remaining, err := s.GetTasks("")
	if err != nil {
		t.Fatalf("GetTasks after cleanup: %v", err)
	}
	for _, task := range remaining {
		if task.ID == "old" {
			t.Fatal("old task should be removed after cleanup")
		}
	}
}

func TestBackgroundExecutorExecuteWritesResult(t *testing.T) {
	resultsDir := t.TempDir()
	executor := NewBackgroundExecutor(func(context.Context) (agentcore.Runner, error) {
		return fakeRunner{content: "执行完成"}, nil
	}, resultsDir)

	output, err := executor.Execute(context.Background(), "task-1", "生成报告")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(output, "执行完成") {
		t.Fatalf("output = %q", output)
	}
	resultPath := executor.taskResultPath("task-1")
	content, err := os.ReadFile(resultPath)
	if err != nil {
		t.Fatalf("read result file: %v", err)
	}
	for _, want := range []string{"# Task Result", "task-1", "生成报告", "执行完成"} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("result file missing %q: %s", want, string(content))
		}
	}
	entries, err := os.ReadDir(filepath.Join(executor.taskDir("task-1"), "history"))
	if err != nil {
		t.Fatalf("read history dir: %v", err)
	}
	if len(entries) != 1 || filepath.Ext(entries[0].Name()) != ".md" {
		t.Fatalf("history entries = %#v", entries)
	}
}

func TestBackgroundExecutorExecuteErrorPaths(t *testing.T) {
	createErr := errors.New("create failed")
	executor := NewBackgroundExecutor(func(context.Context) (agentcore.Runner, error) {
		return nil, createErr
	}, t.TempDir())
	if _, err := executor.Execute(context.Background(), "task-create", "task"); !errors.Is(err, createErr) {
		t.Fatalf("create runner error = %v", err)
	}

	runErr := errors.New("run failed")
	executor = NewBackgroundExecutor(func(context.Context) (agentcore.Runner, error) {
		return fakeRunner{err: runErr}, nil
	}, t.TempDir())
	if _, err := executor.Execute(context.Background(), "task-run", "task"); !errors.Is(err, runErr) {
		t.Fatalf("run error = %v", err)
	}
	content, err := os.ReadFile(executor.taskResultPath("task-run"))
	if err != nil {
		t.Fatalf("read error result: %v", err)
	}
	if !strings.Contains(string(content), "execution error") {
		t.Fatalf("error result content = %s", string(content))
	}
}

func TestSchedulerFormattingAndTools(t *testing.T) {
	s := newTestScheduler(t)
	task := ScheduledTask{
		ID:        "task-json",
		Task:      "整理测试",
		Status:    "completed",
		CreatedAt: time.Now(),
	}
	detail := FormatTaskDetailJSON(task)
	if !strings.Contains(detail, `"id": "task-json"`) || !strings.Contains(detail, `"task": "整理测试"`) {
		t.Fatalf("detail json = %s", detail)
	}
	if got := FormatTasksForDisplay(nil); got != "no scheduled tasks" {
		t.Fatalf("empty display = %q", got)
	}

	tools, err := s.GetTools()
	if err != nil {
		t.Fatalf("GetTools: %v", err)
	}
	if len(tools) != 4 {
		t.Fatalf("tool count = %d, want 4", len(tools))
	}
	info, err := tools[0].Info(context.Background())
	if err != nil {
		t.Fatalf("tool info: %v", err)
	}
	if info.Name != "schedule_add" {
		t.Fatalf("first tool = %s", info.Name)
	}
	if _, err := (*Scheduler)(nil).GetTools(); err == nil {
		t.Fatal("nil scheduler GetTools should fail")
	}
}

type fakeTaskExecutor struct {
	err error
}

func (e fakeTaskExecutor) Execute(context.Context, string, string) (string, error) {
	return "ok", e.err
}

type fakeRunner struct {
	content string
	err     error
}

func (r fakeRunner) Run(_ context.Context, input agentcore.TurnInput, opts agentcore.RunOptions) (*agentcore.RunResult, error) {
	if opts.Sink != nil {
		_ = opts.Sink(agentcore.Event{
			Type:      agentcore.EventMessageDelta,
			AgentName: "tasker",
			Role:      agentcore.RoleAssistant,
			DeltaKind: agentcore.DeltaOutput,
			Content:   r.content,
		})
	}
	if r.err != nil {
		return nil, r.err
	}
	return &agentcore.RunResult{}, nil
}
