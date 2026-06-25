package patch

import (
	"context"
	einoruntime "fkteams/internal/adapters/runtime/eino"
	runtimeport "fkteams/internal/ports/runtime"

	"github.com/cloudwego/eino/adk/middlewares/patchtoolcalls"
)

func New(ctx context.Context) (runtimeport.AgentMiddleware, error) {
	chatModelAgentMiddleware, err := patchtoolcalls.New(ctx, nil)
	if err != nil {
		return nil, err
	}
	return einoruntime.WrapAgentMiddleware("patch_tool_calls", chatModelAgentMiddleware), nil
}
