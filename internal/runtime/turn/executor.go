package turn

import (
	"context"
	"fmt"

	runtimeport "fkteams/internal/ports/runtime"
)

// Executor 执行一次 turn 请求。
type Executor struct{}

func NewExecutor() *Executor {
	return &Executor{}
}

func (e *Executor) Run(ctx context.Context, req Request) (*runtimeport.RunResult, error) {
	if req.Runner == nil {
		return nil, fmt.Errorf("turn runner is required")
	}
	if req.SessionID == "" {
		return nil, fmt.Errorf("turn session ID is required")
	}
	return newEngine(req.Runner, req.SessionID).run(ctx, req)
}
