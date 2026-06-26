// Package turn 提供统一的回合执行内核，封装 Runner 事件循环和 HITL 中断处理。
package turn

import (
	runtimeport "fkteams/internal/ports/runtime"
)

type core struct {
	runner       runtimeport.Runner
	checkpointID string
}

func newEngine(runner runtimeport.Runner, checkpointID string) *core {
	return &core{runner: runner, checkpointID: checkpointID}
}
