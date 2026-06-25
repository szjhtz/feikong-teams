package checkpoint

import "context"

// Store 定义运行时 checkpoint 的最小持久化接口。
type Store interface {
	Set(ctx context.Context, key string, value []byte) error
	Get(ctx context.Context, key string) ([]byte, bool, error)
}
