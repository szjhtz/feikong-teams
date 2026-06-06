package engine

import (
	"context"
	"fkteams/agentcore"
	"fkteams/common"
	"fkteams/events"
)

// run 执行查询，处理事件和 HITL 中断。
// 根据 runConfig 自动装配 context（session ID、事件回调、摘要持久化、审批注册表等）。
func (e *core) run(ctx context.Context, cfg runConfig) (*agentcore.RunResult, error) {
	ctx = cfg.prepareContext(ctx, e.checkpointID)

	if cfg.OnStart != nil {
		cfg.OnStart(ctx)
	}

	result, err := e.runLoop(ctx, cfg.Input, cfg.interruptHandler())

	if cfg.OnFinish != nil {
		cfg.OnFinish(ctx, result, err)
	}

	return result, err
}

func (cfg runConfig) prepareContext(ctx context.Context, checkpointID string) context.Context {
	ctx = common.WithSessionID(ctx, checkpointID)

	if cfg.EventCallback != nil {
		ctx = events.WithCallback(ctx, cfg.EventCallback)
	}

	if cfg.Recorder != nil {
		countBefore := cfg.Recorder.GetMessageCount()
		if !cfg.Input.Message.IsEmpty() {
			cfg.Recorder.RecordUserMessage(cfg.Input.Message)
		}
		ctx = agentcore.WithSummaryPersistCallback(ctx, func(s string) {
			cfg.Recorder.SetSummary(s, countBefore)
		})
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

func (cfg runConfig) interruptHandler() InterruptHandler {
	if cfg.OnInterrupt != nil {
		return cfg.OnInterrupt
	}
	return FixedDecisionHandler(0)
}
