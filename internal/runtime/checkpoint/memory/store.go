package memory

import checkpointstore "fkteams/internal/runtime/checkpoint"

// NewStore 创建内存 checkpoint 存储。
func NewStore() checkpointstore.Store {
	return checkpointstore.NewMemoryStore()
}
