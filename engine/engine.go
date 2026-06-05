// Package engine 提供统一的执行引擎，封装 Runner 事件循环和 HITL 中断处理。
package engine

import (
	"github.com/cloudwego/eino/adk"
)

type core struct {
	runner       *adk.Runner
	checkpointID string
}

func newEngine(runner *adk.Runner, checkpointID string) *core {
	return &core{runner: runner, checkpointID: checkpointID}
}
