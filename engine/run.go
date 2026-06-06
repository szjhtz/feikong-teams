package engine

import (
	"context"
	"fkteams/agents/middlewares/summary"
	"fkteams/common"
	"fkteams/fkevent"
	"strings"

	"github.com/cloudwego/eino/adk"
)

// run 执行查询，处理事件和 HITL 中断。
// 根据 runConfig 自动装配 context（session ID、事件回调、摘要持久化、审批注册表等）。
func (e *core) run(ctx context.Context, cfg runConfig) (*adk.AgentEvent, error) {
	ctx = common.WithSessionID(ctx, e.checkpointID)

	if cfg.EventCallback != nil {
		ctx = fkevent.WithCallback(ctx, cfg.EventCallback)
	}

	if cfg.Recorder != nil {
		countBefore := cfg.Recorder.GetMessageCount()
		if userInput := strings.TrimSpace(cfg.UserInput); userInput != "" {
			cfg.Recorder.RecordUserInput(userInput)
		}
		ctx = summary.WithSummaryPersistCallback(ctx, func(s string) {
			cfg.Recorder.SetSummary(s, countBefore)
		})
	}

	if cfg.NonInteractive {
		ctx = fkevent.WithNonInteractive(ctx)
	}

	for _, hook := range cfg.ContextHooks {
		if hook != nil {
			ctx = hook(ctx)
		}
	}

	if cfg.OnStart != nil {
		cfg.OnStart(ctx)
	}

	handler := cfg.OnInterrupt
	if handler == nil {
		handler = FixedDecisionHandler(0)
	}

	lastEvent, err := e.runLoop(ctx, cfg.Messages, handler)

	if cfg.OnFinish != nil {
		cfg.OnFinish(ctx, lastEvent, err)
	}

	return lastEvent, err
}
