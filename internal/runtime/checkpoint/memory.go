package checkpoint

import (
	"context"
	"sync"
)

// NewMemoryStore 创建进程内 checkpoint 存储。
func NewMemoryStore() Store {
	return &MemoryStore{
		mem: map[string][]byte{},
	}
}

// MemoryStore 是线程安全的进程内 checkpoint 存储。
type MemoryStore struct {
	mu  sync.RWMutex
	mem map[string][]byte
}

func (s *MemoryStore) Set(ctx context.Context, key string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mem[key] = append([]byte(nil), value...)
	return nil
}

func (s *MemoryStore) Get(ctx context.Context, key string) ([]byte, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.mem[key]
	if !ok {
		return nil, false, nil
	}
	return append([]byte(nil), v...), true, nil
}
