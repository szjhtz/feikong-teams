package turn

import (
	"context"
	"fkteams/internal/domain/message"
	"fkteams/internal/domain/session"
	runtimeport "fkteams/internal/ports/runtime"
	"fkteams/internal/runtime/events"
	"fkteams/internal/runtime/hooks"
)

// run 执行查询，处理事件和 HITL 中断。
// 根据 Request 自动装配 context（session ID、事件回调、摘要持久化、审批注册表等）。
func (e *core) run(ctx context.Context, cfg Request) (*runtimeport.RunResult, error) {
	ctx = cfg.prepareContext(ctx, e.checkpointID)

	input, err := cfg.invokeBeforeRun(ctx)
	if err != nil {
		return nil, err
	}
	ctx = cfg.prepareHistoryContext(ctx, input)

	if cfg.OnStart != nil {
		cfg.OnStart(ctx)
	}

	result, err := e.runLoop(ctx, input, cfg.RunID, cfg.interruptHandler())

	if hookErr := cfg.invokeAfterRun(ctx, input, result, err); hookErr != nil && err == nil {
		err = hookErr
	}

	if cfg.OnFinish != nil {
		cfg.OnFinish(ctx, result, err)
	}

	return result, err
}

func (cfg Request) prepareContext(ctx context.Context, checkpointID string) context.Context {
	ctx = session.WithID(ctx, checkpointID)
	ctx = hooks.WithBus(ctx, cfg.hookBus())

	if cfg.EventSink != nil {
		ctx = events.WithCallback(ctx, cfg.EventSink)
	}

	if cfg.NonInteractive {
		ctx = events.WithNonInteractive(ctx)
	}

	for _, hook := range cfg.ContextHooks {
		if hook != nil {
			ctx = hook(ctx)
		}
	}
	return ctx
}

func (cfg Request) prepareHistoryContext(ctx context.Context, input message.TurnInput) context.Context {
	if cfg.Summary != nil {
		countBefore := cfg.Summary.GetMessageCount()
		ctx = runtimeport.WithSummaryPersistCallback(ctx, func(s string) {
			cfg.Summary.SetSummary(s, countBefore)
		})
	}
	return ctx
}

func (cfg Request) invokeBeforeRun(ctx context.Context) (message.TurnInput, error) {
	return cfg.hookBus().InvokeBeforeRun(ctx, cfg.Input)
}

func (cfg Request) invokeAfterRun(ctx context.Context, input message.TurnInput, result *runtimeport.RunResult, runErr error) error {
	return cfg.hookBus().InvokeAfterRun(ctx, input, result, runErr)
}

func (cfg Request) hookBus() *hooks.Bus {
	if cfg.HookBus != nil {
		return cfg.HookBus
	}
	return nil
}

func (cfg Request) interruptHandler() InterruptHandler {
	if cfg.OnInterrupt != nil {
		return cfg.OnInterrupt
	}
	return FixedDecisionHandler(0)
}
