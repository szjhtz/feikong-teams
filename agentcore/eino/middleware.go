package eino

import (
	"fkteams/agentcore"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/compose"
)

func AdaptAgentMiddlewareForRunner(m agentcore.AgentMiddleware) (adk.ChatModelAgentMiddleware, error) {
	if m == nil {
		return nil, fmt.Errorf("middleware is nil")
	}
	runtimeMiddleware := m.RuntimeMiddleware()
	handler, ok := runtimeMiddleware.(adk.ChatModelAgentMiddleware)
	if !ok {
		return nil, fmt.Errorf("unsupported runtime agent middleware: %T", runtimeMiddleware)
	}
	return handler, nil
}

func AdaptAgentMiddlewaresForRunner(middlewares []agentcore.AgentMiddleware) ([]adk.ChatModelAgentMiddleware, error) {
	result := make([]adk.ChatModelAgentMiddleware, 0, len(middlewares))
	for _, m := range middlewares {
		if m == nil {
			continue
		}
		handler, err := AdaptAgentMiddlewareForRunner(m)
		if err != nil {
			return nil, err
		}
		result = append(result, handler)
	}
	return result, nil
}

func AdaptToolMiddlewareForRunner(m agentcore.ToolMiddleware) (compose.ToolMiddleware, error) {
	if m == nil {
		return compose.ToolMiddleware{}, fmt.Errorf("middleware is nil")
	}
	runtimeMiddleware := m.RuntimeMiddleware()
	handler, ok := runtimeMiddleware.(compose.ToolMiddleware)
	if !ok {
		return compose.ToolMiddleware{}, fmt.Errorf("unsupported runtime tool middleware: %T", runtimeMiddleware)
	}
	return handler, nil
}
