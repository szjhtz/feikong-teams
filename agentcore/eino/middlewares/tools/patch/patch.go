package patch

import (
	"context"
	"fkteams/agentcore"

	"github.com/cloudwego/eino/adk/middlewares/patchtoolcalls"
)

func New(ctx context.Context) (agentcore.AgentMiddleware, error) {
	chatModelAgentMiddleware, err := patchtoolcalls.New(ctx, nil)
	if err != nil {
		return nil, err
	}
	return agentcore.WrapRuntimeAgentMiddleware(chatModelAgentMiddleware), nil
}
