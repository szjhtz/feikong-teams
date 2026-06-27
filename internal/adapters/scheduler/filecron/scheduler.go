package filecron

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	domainschedule "fkteams/internal/domain/schedule"
	schedulerport "fkteams/internal/ports/scheduler"
	"fkteams/internal/runtime/log"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	filePath    string
	resultsDir  string
	mu          sync.RWMutex
	stopCh      chan struct{}
	executor    schedulerport.TaskExecutor
	running     bool
	cronParser  cron.Parser
	semaphore   chan struct{}
	cancelFuncs map[string]context.CancelFunc
	cancelsMu   sync.Mutex
}

const (
	maxConcurrentTasks = 5
	taskResultTTL      = 7 * 24 * time.Hour
)

// NewScheduler 创建基于文件存储和 cron 计算的调度器。
func NewScheduler(baseDir string) (*Scheduler, error) {
	absPath, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create scheduler directory: %w", err)
	}

	resultsDir := filepath.Join(absPath, "tasks")
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create tasks directory: %w", err)
	}

	filePath := filepath.Join(absPath, "scheduled_tasks.json")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		emptyList := domainschedule.TaskList{Tasks: []domainschedule.Task{}}
		data, err := json.MarshalIndent(emptyList, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal task list: %w", err)
		}
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return nil, fmt.Errorf("failed to create task list file: %w", err)
		}
	}

	return &Scheduler{
		filePath:    filePath,
		resultsDir:  resultsDir,
		stopCh:      make(chan struct{}),
		cronParser:  cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
		semaphore:   make(chan struct{}, maxConcurrentTasks),
		cancelFuncs: make(map[string]context.CancelFunc),
	}, nil
}

// generateTaskID 生成基于 UUID v4 的任务 ID
func generateTaskID() string {
	return uuid.New().String()
}

// taskDir 返回任务的结果存储目录
func (s *Scheduler) taskDir(taskID string) string {
	return filepath.Join(s.resultsDir, taskID)
}

// taskResultPath 返回任务结果文件路径
func (s *Scheduler) taskResultPath(taskID string) string {
	return filepath.Join(s.taskDir(taskID), "result.md")
}

// ParseCronExpr 解析 cron 表达式并返回下次执行时间
func (s *Scheduler) ParseCronExpr(expr string) (time.Time, error) {
	sched, err := s.cronParser.Parse(expr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid cron expression: %w", err)
	}
	return sched.Next(time.Now()), nil
}

// ComputeNextRun 基于 cron 表达式计算指定时间之后的下次执行时间
func (s *Scheduler) ComputeNextRun(expr string, after time.Time) (time.Time, error) {
	sched, err := s.cronParser.Parse(expr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid cron expression: %w", err)
	}
	return sched.Next(after), nil
}

// AddTask 创建调度任务。
func (s *Scheduler) AddTask(ctx context.Context, req schedulerport.AddTaskRequest) (*domainschedule.Task, error) {
	task := domainschedule.Task{
		ID:        generateTaskID(),
		CreatedAt: time.Now(),
		Status:    domainschedule.StatusPending,
	}
	if err := s.applyTaskSchedule(&task, req); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, err := s.loadTasks()
	if err != nil {
		return nil, fmt.Errorf("load task list: %w", err)
	}
	tasks.Tasks = append(tasks.Tasks, task)
	if err := s.saveTasks(tasks); err != nil {
		return nil, fmt.Errorf("save task list: %w", err)
	}
	return &task, nil
}

// UpdateTask 更新非运行中的调度任务，并重新计算下次执行时间。
func (s *Scheduler) UpdateTask(ctx context.Context, taskID string, req schedulerport.AddTaskRequest) (*domainschedule.Task, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, err := s.loadTasks()
	if err != nil {
		return nil, fmt.Errorf("load task list: %w", err)
	}
	for i := range tasks.Tasks {
		if tasks.Tasks[i].ID != taskID {
			continue
		}
		if tasks.Tasks[i].Status == domainschedule.StatusRunning {
			return nil, fmt.Errorf("cannot update a running task")
		}
		next := tasks.Tasks[i]
		next.Status = domainschedule.StatusPending
		next.LastRunAt = nil
		if err := s.applyTaskSchedule(&next, req); err != nil {
			return nil, err
		}
		tasks.Tasks[i] = next
		if err := s.saveTasks(tasks); err != nil {
			return nil, fmt.Errorf("save task list: %w", err)
		}
		return &next, nil
	}
	return nil, fmt.Errorf("task not found")
}

func (s *Scheduler) applyTaskSchedule(task *domainschedule.Task, req schedulerport.AddTaskRequest) error {
	if strings.TrimSpace(req.Task) == "" {
		return fmt.Errorf("task description is required")
	}
	if req.CronExpr == "" && req.ExecuteAt == "" {
		return fmt.Errorf("must provide cron_expr (recurring) or execute_at (one-time)")
	}
	if req.CronExpr != "" && req.ExecuteAt != "" {
		return fmt.Errorf("cron_expr and execute_at are mutually exclusive")
	}

	task.Task = strings.TrimSpace(req.Task)
	task.CronExpr = ""
	task.OneTime = false
	if req.CronExpr != "" {
		expr := strings.TrimSpace(req.CronExpr)
		nextRun, err := s.ParseCronExpr(expr)
		if err != nil {
			return fmt.Errorf("invalid cron expression: %w", err)
		}
		task.CronExpr = expr
		task.NextRunAt = nextRun
		return nil
	}

	executeAt, err := time.Parse(time.RFC3339, req.ExecuteAt)
	if err != nil {
		return fmt.Errorf("invalid time format, use ISO 8601: %w", err)
	}
	if executeAt.Before(time.Now()) {
		return fmt.Errorf("execute_at must be in the future")
	}
	task.OneTime = true
	task.NextRunAt = executeAt
	return nil
}

// ListTasks 列出调度任务。
func (s *Scheduler) ListTasks(ctx context.Context, statusFilter domainschedule.Status) ([]domainschedule.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks, err := s.loadTasks()
	if err != nil {
		return nil, err
	}
	if statusFilter == "" {
		return tasks.Tasks, nil
	}

	var filtered []domainschedule.Task
	for _, task := range tasks.Tasks {
		if task.Status == statusFilter {
			filtered = append(filtered, task)
		}
	}
	return filtered, nil
}

// CancelTask 取消尚未执行的任务。
func (s *Scheduler) CancelTask(ctx context.Context, taskID string) error {
	if taskID == "" {
		return fmt.Errorf("task ID is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, err := s.loadTasks()
	if err != nil {
		return fmt.Errorf("load task list: %w", err)
	}

	for i := range tasks.Tasks {
		if tasks.Tasks[i].ID != taskID {
			continue
		}
		if tasks.Tasks[i].Status != domainschedule.StatusPending {
			return fmt.Errorf("task status is %s, only pending tasks can be cancelled", tasks.Tasks[i].Status)
		}
		tasks.Tasks[i].Status = domainschedule.StatusCancelled
		if err := s.saveTasks(tasks); err != nil {
			return fmt.Errorf("save task list: %w", err)
		}
		return nil
	}
	return fmt.Errorf("task not found")
}

// DeleteTask 删除非运行中的任务及其结果。
func (s *Scheduler) DeleteTask(ctx context.Context, taskID string) error {
	if taskID == "" {
		return fmt.Errorf("task ID is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, err := s.loadTasks()
	if err != nil {
		return fmt.Errorf("load task list: %w", err)
	}

	found := false
	remaining := make([]domainschedule.Task, 0, len(tasks.Tasks))
	for _, task := range tasks.Tasks {
		if task.ID != taskID {
			remaining = append(remaining, task)
			continue
		}
		if task.Status == domainschedule.StatusRunning {
			return fmt.Errorf("cannot delete a running task, cancel it first")
		}
		found = true
	}
	if !found {
		return fmt.Errorf("task not found")
	}

	if err := os.RemoveAll(s.taskDir(taskID)); err != nil {
		return fmt.Errorf("remove task dir: %w", err)
	}
	tasks.Tasks = remaining
	if err := s.saveTasks(tasks); err != nil {
		return fmt.Errorf("save task list: %w", err)
	}
	return nil
}

// SetExecutor 设置任务执行器
func (s *Scheduler) SetExecutor(executor schedulerport.TaskExecutor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.executor = executor
}

// Start 启动调度器后台轮询
func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	s.recoverStaleRunningTasks()
	go s.run()
}

// recoverStaleRunningTasks 将上次中断遗留的 running 状态任务恢复为 pending
func (s *Scheduler) recoverStaleRunningTasks() {
	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, err := s.loadTasks()
	if err != nil {
		log.Printf("[scheduler] recover stale tasks failed: %v", err)
		return
	}

	changed := false
	for i := range tasks.Tasks {
		if tasks.Tasks[i].Status == domainschedule.StatusRunning {
			tasks.Tasks[i].Status = domainschedule.StatusPending
			changed = true
			log.Printf("[scheduler] recover stale task: %s → pending", tasks.Tasks[i].ID)
		}
	}

	if changed {
		if err := s.saveTasks(tasks); err != nil {
			log.Printf("[scheduler] save recovered tasks failed: %v", err)
		}
	}
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		close(s.stopCh)
		s.running = false
	}

	// 取消所有正在执行的任务
	s.cancelsMu.Lock()
	for taskID, cancel := range s.cancelFuncs {
		log.Printf("[scheduler] cancelling running task: %s", taskID)
		cancel()
	}
	s.cancelsMu.Unlock()
}

func (s *Scheduler) run() {
	// 启动时立即检查一次
	s.checkAndExecute()

	for {
		// 计算下次检查间隔
		waitDuration := s.nextCheckDuration()

		timer := time.NewTimer(waitDuration)
		select {
		case <-s.stopCh:
			timer.Stop()
			return
		case <-timer.C:
			s.checkAndExecute()
		}
	}
}

// nextCheckDuration 计算距下次任务到期的等待时间
func (s *Scheduler) nextCheckDuration() time.Duration {
	tasks, err := s.loadTasks()
	if err != nil {
		return 30 * time.Second
	}

	now := time.Now()
	minWait := 30 * time.Second
	for _, t := range tasks.Tasks {
		if t.Status != domainschedule.StatusPending {
			continue
		}
		wait := t.NextRunAt.Sub(now)
		if wait <= 0 {
			return 0
		}
		if wait < minWait {
			minWait = wait + 500*time.Millisecond
		}
	}

	// 至少每 30 秒检查一次，用于 TTL 清理
	if minWait > 30*time.Second {
		minWait = 30 * time.Second
	}

	return minWait
}

func (s *Scheduler) checkAndExecute() {
	// TTL 清理（每轮检查时顺便执行）
	s.cleanupExpiredTasks()

	s.mu.RLock()
	executor := s.executor
	s.mu.RUnlock()

	if executor == nil {
		return
	}

	tasks, err := s.loadTasks()
	if err != nil {
		log.Printf("[scheduler] load tasks failed: %v", err)
		return
	}

	now := time.Now()
	for i := range tasks.Tasks {
		task := &tasks.Tasks[i]
		if task.Status != domainschedule.StatusPending {
			continue
		}
		if now.Before(task.NextRunAt) {
			continue
		}

		// 加写锁二次确认状态，防止重复执行
		s.mu.Lock()
		currentTasks, loadErr := s.loadTasks()
		if loadErr != nil {
			s.mu.Unlock()
			log.Printf("[scheduler] re-check load failed: %v", loadErr)
			continue
		}

		var currentTask *domainschedule.Task
		for j := range currentTasks.Tasks {
			if currentTasks.Tasks[j].ID == task.ID {
				currentTask = &currentTasks.Tasks[j]
				break
			}
		}

		if currentTask == nil || currentTask.Status != domainschedule.StatusPending {
			s.mu.Unlock()
			continue
		}

		currentTask.Status = domainschedule.StatusRunning
		currentTask.LastRunAt = &now
		if saveErr := s.saveTasks(currentTasks); saveErr != nil {
			s.mu.Unlock()
			log.Printf("[scheduler] save task status failed: %v", saveErr)
			continue
		}
		s.mu.Unlock()

		// 通过 semaphore 控制并发
		select {
		case s.semaphore <- struct{}{}:
			go func(tID, tContent, tCron string, tOneTime bool, tExec schedulerport.TaskExecutor) {
				defer func() { <-s.semaphore }()
				s.executeTask(tID, tContent, tCron, tOneTime, tExec)
			}(currentTask.ID, currentTask.Task, currentTask.CronExpr, currentTask.OneTime, executor)
		default:
			// 并发已满，回退状态到 pending
			log.Printf("[scheduler] max concurrent reached (%d), task %s deferred", maxConcurrentTasks, currentTask.ID)
			s.mu.Lock()
			fallbackTasks, _ := s.loadTasks()
			if fallbackTasks != nil {
				for j := range fallbackTasks.Tasks {
					if fallbackTasks.Tasks[j].ID == currentTask.ID {
						fallbackTasks.Tasks[j].Status = domainschedule.StatusPending
						_ = s.saveTasks(fallbackTasks)
						break
					}
				}
			}
			s.mu.Unlock()
		}
	}
}

func (s *Scheduler) executeTask(taskID string, taskContent string, cronExpr string, oneTime bool, executor schedulerport.TaskExecutor) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	s.cancelsMu.Lock()
	s.cancelFuncs[taskID] = cancel
	s.cancelsMu.Unlock()

	defer func() {
		s.cancelsMu.Lock()
		delete(s.cancelFuncs, taskID)
		s.cancelsMu.Unlock()
	}()

	log.Printf("[scheduler] task started: %s", taskID)
	_, err := executor.Execute(ctx, taskID, taskContent)

	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, loadErr := s.loadTasks()
	if loadErr != nil {
		log.Printf("[scheduler] load after execute failed (result lost): taskID=%s, err=%v", taskID, loadErr)
		return
	}

	for i := range tasks.Tasks {
		if tasks.Tasks[i].ID == taskID {
			now := time.Now()
			tasks.Tasks[i].LastRunAt = &now

			if err != nil {
				tasks.Tasks[i].Status = domainschedule.StatusFailed
				log.Printf("[scheduler] task failed: %s, err=%v", taskID, err)
			} else {
				if oneTime {
					tasks.Tasks[i].Status = domainschedule.StatusCompleted
				} else {
					nextRun, cronErr := s.ComputeNextRun(cronExpr, now)
					if cronErr != nil {
						tasks.Tasks[i].Status = domainschedule.StatusFailed
						log.Printf("[scheduler] cron parse failed: taskID=%s, err=%v", taskID, cronErr)
					} else {
						// 避免紧贴当前时间重复触发
						if nextRun.Sub(now) < 30*time.Second {
							nextRun, cronErr = s.ComputeNextRun(cronExpr, nextRun)
							if cronErr != nil {
								tasks.Tasks[i].Status = domainschedule.StatusFailed
								log.Printf("[scheduler] cron parse failed (skip): taskID=%s, err=%v", taskID, cronErr)
								break
							}
						}
						tasks.Tasks[i].Status = domainschedule.StatusPending
						tasks.Tasks[i].NextRunAt = nextRun
					}
				}
			}
			log.Printf("[scheduler] task done: %s, status=%s", taskID, tasks.Tasks[i].Status)
			break
		}
	}

	if saveErr := s.saveTasks(tasks); saveErr != nil {
		log.Printf("[scheduler] save result failed: taskID=%s, err=%v", taskID, saveErr)
	}
}

// cleanupExpiredTasks 清理超过 TTL 的已完成/失败/取消的一次性任务
func (s *Scheduler) cleanupExpiredTasks() {
	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, err := s.loadTasks()
	if err != nil {
		return
	}

	cutoff := time.Now().Add(-taskResultTTL)
	var remaining []domainschedule.Task
	removed := 0

	for _, t := range tasks.Tasks {
		if t.Status == domainschedule.StatusCompleted || t.Status == domainschedule.StatusFailed || t.Status == domainschedule.StatusCancelled {
			refTime := t.LastRunAt
			if refTime == nil {
				refTime = &t.CreatedAt
			}
			if refTime.Before(cutoff) {
				// 删除任务目录
				if err := os.RemoveAll(s.taskDir(t.ID)); err != nil {
					log.Printf("[scheduler] cleanup task dir failed: taskID=%s, err=%v", t.ID, err)
				}
				removed++
				continue
			}
		}
		remaining = append(remaining, t)
	}

	if removed > 0 {
		log.Printf("[scheduler] cleaned up %d expired tasks", removed)
		tasks.Tasks = remaining
		if err := s.saveTasks(tasks); err != nil {
			log.Printf("[scheduler] save after cleanup failed: %v", err)
		}
	}
}

// CancelExecution 取消正在执行的任务
func (s *Scheduler) CancelExecution(taskID string) {
	s.cancelsMu.Lock()
	if cancel, ok := s.cancelFuncs[taskID]; ok {
		cancel()
		delete(s.cancelFuncs, taskID)
		log.Printf("[scheduler] task execution cancelled: %s", taskID)
	}
	s.cancelsMu.Unlock()
}

// loadTaskByID 在持有锁的情况下根据 ID 查找任务
func (s *Scheduler) loadTaskByID(taskID string) (*domainschedule.Task, error) {
	tasks, err := s.loadTasks()
	if err != nil {
		return nil, err
	}
	for i := range tasks.Tasks {
		if tasks.Tasks[i].ID == taskID {
			return &tasks.Tasks[i], nil
		}
	}
	return nil, nil
}

func (s *Scheduler) loadTasks() (*domainschedule.TaskList, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read task list: %w", err)
	}

	var list domainschedule.TaskList
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("failed to parse task list: %w", err)
	}

	return &list, nil
}

func (s *Scheduler) saveTasks(list *domainschedule.TaskList) error {
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal task list: %w", err)
	}

	tmpPath := s.filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := os.Rename(tmpPath, s.filePath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// ReadTaskResult 读取任务执行结果。
func (s *Scheduler) ReadTaskResult(ctx context.Context, taskID string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, err := s.loadTaskByID(taskID)
	if err != nil {
		return "", fmt.Errorf("load task: %w", err)
	}
	if task == nil {
		return "", fmt.Errorf("task not found: %s", taskID)
	}
	resultPath := s.taskResultPath(taskID)
	if _, err := os.Stat(resultPath); os.IsNotExist(err) {
		return "", fmt.Errorf("task %s has no result yet", taskID)
	}

	data, err := os.ReadFile(resultPath)
	if err != nil {
		return "", fmt.Errorf("read result file: %w", err)
	}
	return string(data), nil
}

// ListHistoryEntries 列出任务的历史结果文件。
func (s *Scheduler) ListHistoryEntries(ctx context.Context, taskID string) ([]domainschedule.HistoryEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	historyDir := filepath.Join(s.taskDir(taskID), "history")
	entries, err := os.ReadDir(historyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read history dir: %w", err)
	}

	var result []domainschedule.HistoryEntry
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		name := entry.Name()
		timeStr := ""
		if len(name) >= 15 {
			timeStr = fmt.Sprintf("%s-%s-%s %s:%s:%s",
				name[0:4], name[4:6], name[6:8],
				name[9:11], name[11:13], name[13:15])
		}
		result = append(result, domainschedule.HistoryEntry{
			Filename: name,
			Time:     timeStr,
		})
	}
	return result, nil
}

// ReadHistoryFile 读取指定历史结果文件。
func (s *Scheduler) ReadHistoryFile(ctx context.Context, taskID string, filename string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filename = filepath.Base(filename)
	if filepath.Ext(filename) != ".md" {
		return "", fmt.Errorf("invalid file type")
	}

	filePath := filepath.Join(s.taskDir(taskID), "history", filename)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("read history file: %w", err)
	}
	return string(data), nil
}
