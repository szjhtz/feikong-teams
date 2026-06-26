package scheduler

import (
	"context"
	"time"

	domainschedule "fkteams/internal/domain/schedule"
)

// AddTaskRequest 描述创建调度任务所需的最小输入。
type AddTaskRequest struct {
	Task      string
	CronExpr  string
	ExecuteAt string
}

// TaskExecutor 执行已经到期的调度任务。
type TaskExecutor interface {
	Execute(ctx context.Context, taskID string, task string) (string, error)
}

// Scheduler 管理任务计划、状态流转和结果存档。
type Scheduler interface {
	SetExecutor(executor TaskExecutor)
	Start()
	Stop()
	AddTask(ctx context.Context, req AddTaskRequest) (*domainschedule.Task, error)
	ListTasks(ctx context.Context, statusFilter domainschedule.Status) ([]domainschedule.Task, error)
	CancelTask(ctx context.Context, taskID string) error
	DeleteTask(ctx context.Context, taskID string) error
	ReadTaskResult(ctx context.Context, taskID string) (string, error)
	ListHistoryEntries(ctx context.Context, taskID string) ([]domainschedule.HistoryEntry, error)
	ReadHistoryFile(ctx context.Context, taskID string, filename string) (string, error)
	ComputeNextRun(expr string, after time.Time) (time.Time, error)
}
