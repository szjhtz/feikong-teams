package common

import (
	"fkteams/agentcore"
	"fkteams/internal/runtime/checkpoint"
)

// NewInMemoryStore 创建基于内存的 CheckPoint 存储
func NewInMemoryStore() agentcore.CheckPointStore {
	return checkpoint.NewMemoryStore()
}
