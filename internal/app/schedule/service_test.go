package schedule

import (
	"context"
	"errors"
	"testing"
	"time"

	domainschedule "fkteams/internal/domain/schedule"
	schedulerport "fkteams/internal/ports/scheduler"
)

func TestServiceDelegatesToSchedulerPort(t *testing.T) {
	fake := &fakeScheduler{}
	service := NewService(fake)

	task, err := service.AddTask(context.Background(), schedulerport.AddTaskRequest{Task: "生成日报"})
	if err != nil {
		t.Fatalf("AddTask: %v", err)
	}
	if task.ID != "task-1" || fake.addReq.Task != "生成日报" {
		t.Fatalf("task = %#v, addReq = %#v", task, fake.addReq)
	}

	updated, err := service.UpdateTask(context.Background(), "task-1", schedulerport.AddTaskRequest{Task: "更新日报"})
	if err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	if updated.Task != "更新日报" || fake.updateID != "task-1" {
		t.Fatalf("updated = %#v, updateID = %s", updated, fake.updateID)
	}

	tasks, err := service.ListTasks(context.Background(), domainschedule.StatusPending)
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(tasks) != 1 || fake.listStatus != domainschedule.StatusPending {
		t.Fatalf("tasks = %#v, status = %s", tasks, fake.listStatus)
	}

	if err := service.CancelTask(context.Background(), "task-1"); err != nil {
		t.Fatalf("CancelTask: %v", err)
	}
	if fake.cancelID != "task-1" {
		t.Fatalf("cancelID = %s", fake.cancelID)
	}
}

func TestServiceRequiresScheduler(t *testing.T) {
	if _, err := (*Service)(nil).ListTasks(context.Background(), ""); err == nil {
		t.Fatal("expected nil service error")
	}
	if err := NewService(nil).CancelTask(context.Background(), "task-1"); err == nil {
		t.Fatal("expected nil scheduler error")
	}
	if err := NewService(&fakeScheduler{}).CancelTask(context.Background(), ""); err == nil {
		t.Fatal("expected task ID error")
	}
}

type fakeScheduler struct {
	addReq     schedulerport.AddTaskRequest
	updateReq  schedulerport.AddTaskRequest
	updateID   string
	listStatus domainschedule.Status
	cancelID   string
}

func (s *fakeScheduler) SetExecutor(schedulerport.TaskExecutor) {}
func (s *fakeScheduler) Start()                                 {}
func (s *fakeScheduler) Stop()                                  {}

func (s *fakeScheduler) AddTask(ctx context.Context, req schedulerport.AddTaskRequest) (*domainschedule.Task, error) {
	s.addReq = req
	return &domainschedule.Task{ID: "task-1", Task: req.Task, Status: domainschedule.StatusPending}, nil
}

func (s *fakeScheduler) UpdateTask(ctx context.Context, taskID string, req schedulerport.AddTaskRequest) (*domainschedule.Task, error) {
	s.updateID = taskID
	s.updateReq = req
	return &domainschedule.Task{ID: taskID, Task: req.Task, Status: domainschedule.StatusPending}, nil
}

func (s *fakeScheduler) ListTasks(ctx context.Context, statusFilter domainschedule.Status) ([]domainschedule.Task, error) {
	s.listStatus = statusFilter
	return []domainschedule.Task{{ID: "task-1", Task: "生成日报", Status: domainschedule.StatusPending}}, nil
}

func (s *fakeScheduler) CancelTask(ctx context.Context, taskID string) error {
	s.cancelID = taskID
	return nil
}

func (s *fakeScheduler) DeleteTask(ctx context.Context, taskID string) error {
	return nil
}

func (s *fakeScheduler) ReadTaskResult(ctx context.Context, taskID string) (string, error) {
	return "", errors.New("not used")
}

func (s *fakeScheduler) ListHistoryEntries(ctx context.Context, taskID string) ([]domainschedule.HistoryEntry, error) {
	return nil, errors.New("not used")
}

func (s *fakeScheduler) ReadHistoryFile(ctx context.Context, taskID string, filename string) (string, error) {
	return "", errors.New("not used")
}

func (s *fakeScheduler) ComputeNextRun(expr string, after time.Time) (time.Time, error) {
	return time.Time{}, errors.New("not used")
}
