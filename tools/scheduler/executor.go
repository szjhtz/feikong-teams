package scheduler

import (
	"context"
	"fkteams/agentcore"
	"fkteams/engine"
	"fkteams/events/view"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RunnerCreator 创建任务执行用 Runner。
type RunnerCreator func(ctx context.Context) (agentcore.Runner, error)

// BackgroundExecutor 在后台执行任务。
type BackgroundExecutor struct {
	createRunner RunnerCreator
	resultsDir   string
}

// NewBackgroundExecutor 创建后台任务执行器。
func NewBackgroundExecutor(createRunner RunnerCreator, resultsDir string) *BackgroundExecutor {
	_ = os.MkdirAll(resultsDir, 0755)
	return &BackgroundExecutor{
		createRunner: createRunner,
		resultsDir:   resultsDir,
	}
}

// taskDir 返回任务结果目录。
func (e *BackgroundExecutor) taskDir(taskID string) string {
	return filepath.Join(e.resultsDir, taskID)
}

// taskResultPath 返回任务结果文件路径。
func (e *BackgroundExecutor) taskResultPath(taskID string) string {
	return filepath.Join(e.taskDir(taskID), "result.md")
}

// Execute 执行任务并写入结果目录。
func (e *BackgroundExecutor) Execute(ctx context.Context, taskID string, task string) (string, error) {
	if err := os.MkdirAll(e.taskDir(taskID), 0755); err != nil {
		return "", fmt.Errorf("create task dir: %w", err)
	}

	r, err := e.createRunner(ctx)
	if err != nil {
		return "", fmt.Errorf("create runner: %w", err)
	}

	callback, getResult := eventview.NewMarkdownCollector()

	input := engine.TurnInput{
		Message: agentcore.Message{Role: agentcore.RoleUser, Content: task},
	}

	_, err = engine.NewSession(r, "fkteams_scheduler").
		WithInput(input).
		OnEvent(callback).
		Run(ctx)
	if err != nil {
		errMsg := fmt.Sprintf("execution error: %v", err)
		e.writeResult(taskID, task, errMsg)
		return "", err
	}

	output := getResult()
	e.writeResult(taskID, task, output)
	return output, nil
}

// writeResult 写入最新结果和历史快照。
func (e *BackgroundExecutor) writeResult(taskID string, task string, result string) {
	now := time.Now()
	ts := now.Format("20060102_150405")

	content := fmt.Sprintf("# Task Result\n\n**Task ID**: %s\n\n**Time**: %s\n\n**Task**: %s\n\n## Result\n\n%s\n",
		taskID,
		now.Format("2006-01-02 15:04:05"),
		task,
		result,
	)

	_ = os.WriteFile(e.taskResultPath(taskID), []byte(content), 0644)

	historyDir := filepath.Join(e.taskDir(taskID), "history")
	_ = os.MkdirAll(historyDir, 0755)
	_ = os.WriteFile(filepath.Join(historyDir, ts+".md"), []byte(content), 0644)
}
