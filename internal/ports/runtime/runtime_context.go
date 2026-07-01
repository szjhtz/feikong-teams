package runtime

import (
	"context"
	"fmt"
)

type agentRuntimeContextKey struct{}
type runnerRuntimeContextKey struct{}
type agentToolRuntimeContextKey struct{}
type pipelineRuntimeContextKey struct{}
type runtimeContextKey struct{}

// WithRuntime 将一个 runtime adapter 按其实现的能力拆分注入 context。
func WithRuntime(ctx context.Context, runtime any) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if runtime == nil {
		return ctx
	}
	if coreRuntime, ok := runtime.(Runtime); ok {
		ctx = context.WithValue(ctx, runtimeContextKey{}, coreRuntime)
	}
	if agentRuntime, ok := runtime.(AgentRuntime); ok {
		ctx = WithAgentRuntime(ctx, agentRuntime)
	}
	if runnerRuntime, ok := runtime.(RunnerRuntime); ok {
		ctx = WithRunnerRuntime(ctx, runnerRuntime)
	}
	if agentToolRuntime, ok := runtime.(AgentToolRuntime); ok {
		ctx = WithAgentToolRuntime(ctx, agentToolRuntime)
	}
	if pipelineRuntime, ok := runtime.(PipelineRuntime); ok {
		ctx = WithPipelineRuntime(ctx, pipelineRuntime)
	}
	return ctx
}

func RuntimeFromContext(ctx context.Context) (Runtime, bool) {
	if ctx == nil {
		return nil, false
	}
	runtime, ok := ctx.Value(runtimeContextKey{}).(Runtime)
	return runtime, ok && runtime != nil
}

func WithAgentRuntime(ctx context.Context, runtime AgentRuntime) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if runtime == nil {
		return ctx
	}
	return context.WithValue(ctx, agentRuntimeContextKey{}, runtime)
}

func AgentRuntimeFromContext(ctx context.Context) (AgentRuntime, bool) {
	if ctx == nil {
		return nil, false
	}
	runtime, ok := ctx.Value(agentRuntimeContextKey{}).(AgentRuntime)
	return runtime, ok && runtime != nil
}

func RequireAgentRuntime(ctx context.Context) (AgentRuntime, error) {
	if runtime, ok := AgentRuntimeFromContext(ctx); ok {
		return runtime, nil
	}
	return nil, fmt.Errorf("agent runtime is not configured")
}

func WithRunnerRuntime(ctx context.Context, runtime RunnerRuntime) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if runtime == nil {
		return ctx
	}
	return context.WithValue(ctx, runnerRuntimeContextKey{}, runtime)
}

func RunnerRuntimeFromContext(ctx context.Context) (RunnerRuntime, bool) {
	if ctx == nil {
		return nil, false
	}
	runtime, ok := ctx.Value(runnerRuntimeContextKey{}).(RunnerRuntime)
	return runtime, ok && runtime != nil
}

func RequireRunnerRuntime(ctx context.Context) (RunnerRuntime, error) {
	if runtime, ok := RunnerRuntimeFromContext(ctx); ok {
		return runtime, nil
	}
	return nil, fmt.Errorf("runner runtime is not configured")
}

func WithAgentToolRuntime(ctx context.Context, runtime AgentToolRuntime) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if runtime == nil {
		return ctx
	}
	return context.WithValue(ctx, agentToolRuntimeContextKey{}, runtime)
}

func AgentToolRuntimeFromContext(ctx context.Context) (AgentToolRuntime, bool) {
	if ctx == nil {
		return nil, false
	}
	runtime, ok := ctx.Value(agentToolRuntimeContextKey{}).(AgentToolRuntime)
	return runtime, ok && runtime != nil
}

func RequireAgentToolRuntime(ctx context.Context) (AgentToolRuntime, error) {
	if runtime, ok := AgentToolRuntimeFromContext(ctx); ok {
		return runtime, nil
	}
	return nil, fmt.Errorf("agent tool runtime is not configured")
}

func WithPipelineRuntime(ctx context.Context, runtime PipelineRuntime) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if runtime == nil {
		return ctx
	}
	return context.WithValue(ctx, pipelineRuntimeContextKey{}, runtime)
}

func PipelineRuntimeFromContext(ctx context.Context) (PipelineRuntime, bool) {
	if ctx == nil {
		return nil, false
	}
	runtime, ok := ctx.Value(pipelineRuntimeContextKey{}).(PipelineRuntime)
	return runtime, ok && runtime != nil
}
