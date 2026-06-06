// Package engine 提供统一的执行引擎，封装 Runner 事件循环和 HITL 中断处理。
package engine

import (
	"fkteams/agentcore"
)

type core struct {
	runner       agentcore.Runner
	checkpointID string
}

func newEngine(runner agentcore.Runner, checkpointID string) *core {
	return &core{runner: runner, checkpointID: checkpointID}
}
