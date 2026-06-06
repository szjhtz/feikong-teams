package engine

import (
	"context"
	"fkteams/agentcore"
	"fkteams/events"
)

// runLoop delegates one runner invocation with engine-level options assembled.
// Runtime adapters own provider-specific interrupt and resume details.
func (e *core) runLoop(ctx context.Context, input agentcore.TurnInput, handler InterruptHandler) (*agentcore.RunResult, error) {
	return e.runner.Run(ctx, input, agentcore.RunOptions{
		RunID:            e.checkpointID,
		CheckpointID:     e.checkpointID,
		Sink:             events.Dispatch(ctx),
		InterruptHandler: agentcore.InterruptHandler(handler),
	})
}
