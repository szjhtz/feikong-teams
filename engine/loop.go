package engine

import (
	"context"
	"fkteams/agentcore"
	"fkteams/fkevent"
)

// runLoop drives one engine-neutral runner invocation.
func (e *core) runLoop(ctx context.Context, input agentcore.TurnInput, handler InterruptHandler) (*agentcore.RunResult, error) {
	return e.runner.Run(ctx, input, agentcore.RunOptions{
		RunID:            e.checkpointID,
		CheckpointID:     e.checkpointID,
		Sink:             fkevent.Dispatch(ctx),
		InterruptHandler: agentcore.InterruptHandler(handler),
	})
}
