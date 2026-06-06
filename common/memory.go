package common

import (
	"context"
	"fkteams/agentcore"
	"sync"
)

// NewInMemoryStore 创建基于内存的 CheckPoint 存储
func NewInMemoryStore() agentcore.CheckPointStore {
	return &inMemoryStore{
		mem: map[string][]byte{},
	}
}

type inMemoryStore struct {
	mu  sync.RWMutex
	mem map[string][]byte
}

func (i *inMemoryStore) Set(ctx context.Context, key string, value []byte) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.mem[key] = append([]byte(nil), value...)
	return nil
}

func (i *inMemoryStore) Get(ctx context.Context, key string) ([]byte, bool, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	v, ok := i.mem[key]
	if !ok {
		return nil, false, nil
	}
	return append([]byte(nil), v...), true, nil
}
