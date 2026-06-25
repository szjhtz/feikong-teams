package checkpoint

import (
	"context"
	"strings"
)

// NewNamespaceStore 为底层 Store 增加 key 命名空间，避免不同 Runner 共享存储时互相覆盖。
func NewNamespaceStore(namespace string, inner Store) Store {
	if inner == nil {
		inner = NewMemoryStore()
	}
	namespace = strings.TrimSpace(namespace)
	if namespace == "" {
		return inner
	}
	return &NamespaceStore{
		namespace: namespace,
		inner:     inner,
	}
}

// NamespaceStore 为共享 Store 提供固定前缀隔离。
type NamespaceStore struct {
	namespace string
	inner     Store
}

func (s *NamespaceStore) Set(ctx context.Context, key string, value []byte) error {
	return s.inner.Set(ctx, s.key(key), value)
}

func (s *NamespaceStore) Get(ctx context.Context, key string) ([]byte, bool, error) {
	return s.inner.Get(ctx, s.key(key))
}

func (s *NamespaceStore) key(key string) string {
	return s.namespace + ":" + key
}
