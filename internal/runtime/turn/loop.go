package turn

import (
	"context"
	"fkteams/events"
	"fkteams/internal/domain/message"
	runtimeport "fkteams/internal/ports/runtime"
)

// runLoop 装配引擎级选项后执行一次 Runner 调用。
func (e *core) runLoop(ctx context.Context, input message.TurnInput, runID string, handler InterruptHandler) (*runtimeport.RunResult, error) {
	if runID == "" {
		runID = e.checkpointID
	}
	return e.runner.Run(ctx, input, runtimeport.RunOptions{
		RunID:            runID,
		CheckpointID:     e.checkpointID,
		Sink:             events.Dispatch(ctx),
		InterruptHandler: runtimeport.InterruptHandler(handler),
	})
}
