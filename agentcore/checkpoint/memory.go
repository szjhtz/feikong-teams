package checkpoint

import checkpointstore "fkteams/internal/runtime/checkpoint"

type MemoryStore = checkpointstore.MemoryStore

func NewMemoryStore() Store {
	return checkpointstore.NewMemoryStore()
}
