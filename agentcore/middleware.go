package agentcore

type AgentMiddleware interface {
	RuntimeMiddleware() any
}

type ToolMiddleware interface {
	RuntimeMiddleware() any
}

type runtimeMiddleware struct {
	runtime any
}

func WrapRuntimeAgentMiddleware(runtime any) AgentMiddleware {
	return &runtimeMiddleware{runtime: runtime}
}

func WrapRuntimeToolMiddleware(runtime any) ToolMiddleware {
	return &runtimeMiddleware{runtime: runtime}
}

func (m *runtimeMiddleware) RuntimeMiddleware() any {
	if m == nil {
		return nil
	}
	return m.runtime
}
