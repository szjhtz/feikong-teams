package agentcore

import "context"

type Agent interface {
	Name() string
	Description() string
	RuntimeAgent() any
}

type UnknownToolHandler func(ctx context.Context, name, arguments string) (string, error)

type RetryContext struct {
	Err error
}

type RetryDecision struct {
	Retry        bool
	RejectReason string
}

type ModelRetryConfig struct {
	MaxRetries  int
	ShouldRetry func(ctx context.Context, retryCtx *RetryContext) *RetryDecision
}

type ChatAgentConfig struct {
	Name               string
	Description        string
	Instruction        string
	Model              ChatModel
	Tools              []Tool
	ToolMiddlewares    []ToolMiddleware
	UnknownToolHandler UnknownToolHandler
	Middlewares        []AgentMiddleware
	ModelRetryConfig   *ModelRetryConfig
	MaxIterations      int
	EmitInternalEvents bool
}

type LoopAgentConfig struct {
	Name          string
	Description   string
	SubAgents     []Agent
	MaxIterations int
}

type runtimeAgent struct {
	name        string
	description string
	runtime     any
}

func WrapRuntimeAgent(name, description string, runtime any) Agent {
	return &runtimeAgent{name: name, description: description, runtime: runtime}
}

func (a *runtimeAgent) Name() string {
	if a == nil {
		return ""
	}
	return a.name
}

func (a *runtimeAgent) Description() string {
	if a == nil {
		return ""
	}
	return a.description
}

func (a *runtimeAgent) RuntimeAgent() any {
	if a == nil {
		return nil
	}
	return a.runtime
}
