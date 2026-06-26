package storage

import "context"

// CheckpointStore 定义运行时 checkpoint 的最小持久化能力。
type CheckpointStore interface {
	Set(ctx context.Context, key string, value []byte) error
	Get(ctx context.Context, key string) ([]byte, bool, error)
}
