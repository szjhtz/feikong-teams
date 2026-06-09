package engine

import (
	"context"
	"fkteams/agentcore"
	"fkteams/events"
)

// runLoop 装配引擎级选项后执行一次 Runner 调用。
func (e *core) runLoop(ctx context.Context, input agentcore.TurnInput, handler InterruptHandler) (*agentcore.RunResult, error) {
	return e.runner.Run(ctx, input, agentcore.RunOptions{
		RunID:            e.checkpointID,
		CheckpointID:     e.checkpointID,
		Sink:             events.Dispatch(ctx),
		InterruptHandler: agentcore.InterruptHandler(handler),
	})
}
