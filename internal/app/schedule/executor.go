package schedule

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	appchat "fkteams/internal/app/chat"
	"fkteams/internal/domain/message"
	domainsession "fkteams/internal/domain/session"
	runtimeport "fkteams/internal/ports/runtime"
	"fkteams/internal/runtime/atomicfile"
)

// RunnerCreator 为每次后台任务创建独立运行器。
type RunnerCreator func(ctx context.Context) (runtimeport.Runner, error)

// BackgroundExecutor 将调度任务转换为一次后台聊天运行。
type BackgroundExecutor struct {
	createRunner RunnerCreator
	resultsDir   string
	chat         *appchat.Service
	contextHook  func(context.Context) context.Context
}

// NewBackgroundExecutor 创建后台任务执行器。
func NewBackgroundExecutor(createRunner RunnerCreator, resultsDir string) *BackgroundExecutor {
	_ = os.MkdirAll(resultsDir, 0755)
	return &BackgroundExecutor{
		createRunner: createRunner,
		resultsDir:   resultsDir,
		chat:         appchat.NewService(),
	}
}

// WithContextHook 设置每次执行前的上下文装配逻辑。
func (e *BackgroundExecutor) WithContextHook(hook func(context.Context) context.Context) *BackgroundExecutor {
	e.contextHook = hook
	return e
}

func (e *BackgroundExecutor) taskDir(taskID string) string {
	return filepath.Join(e.resultsDir, taskID)
}

func (e *BackgroundExecutor) taskResultPath(taskID string) string {
	return filepath.Join(e.taskDir(taskID), "result.md")
}

// Execute 执行调度任务并写入当前结果和历史快照。
func (e *BackgroundExecutor) Execute(ctx context.Context, taskID string, task string) (string, error) {
	if !domainsession.ValidID(taskID) || len(taskID) > 160 {
		return "", fmt.Errorf("invalid task ID")
	}
	if e.contextHook != nil {
		ctx = e.contextHook(ctx)
	}
	if err := os.MkdirAll(e.taskDir(taskID), 0755); err != nil {
		return "", fmt.Errorf("create task dir: %w", err)
	}
	if e.createRunner == nil {
		return "", fmt.Errorf("create runner: runner creator is nil")
	}

	r, err := e.createRunner(ctx)
	if err != nil {
		return "", fmt.Errorf("create runner: %w", err)
	}

	callback, getResult := newMarkdownCollector()
	input := message.TurnInput{
		Message: message.Message{Role: message.RoleUser, Content: task},
	}

	_, err = e.chat.RunTurn(ctx, appchat.TurnRequest{
		SessionID: "fkteams_scheduler_" + taskID,
		Runner:    r,
		Input:     input,
		EventSink: callback,
	})
	if err != nil {
		errMsg := fmt.Sprintf("execution error: %v", err)
		if writeErr := e.writeResult(taskID, task, errMsg); writeErr != nil {
			return "", errors.Join(err, writeErr)
		}
		return "", err
	}

	output := getResult()
	if err := e.writeResult(taskID, task, output); err != nil {
		return "", err
	}
	return output, nil
}

func (e *BackgroundExecutor) writeResult(taskID string, task string, result string) error {
	now := time.Now()
	ts := now.Format("20060102_150405")

	content := fmt.Sprintf("# Task Result\n\n**Task ID**: %s\n\n**Time**: %s\n\n**Task**: %s\n\n## Result\n\n%s\n",
		taskID,
		now.Format("2006-01-02 15:04:05"),
		task,
		result,
	)

	if err := atomicfile.WriteFile(e.taskResultPath(taskID), []byte(content), 0644); err != nil {
		return fmt.Errorf("write task result: %w", err)
	}

	historyDir := filepath.Join(e.taskDir(taskID), "history")
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return fmt.Errorf("create task history dir: %w", err)
	}
	if err := atomicfile.WriteFile(filepath.Join(historyDir, ts+".md"), []byte(content), 0644); err != nil {
		return fmt.Errorf("write task history: %w", err)
	}
	return nil
}
