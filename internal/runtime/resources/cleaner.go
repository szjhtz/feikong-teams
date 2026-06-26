package resources

import (
	"fmt"
	"sync"
)

// CleanupFunc 定义清理函数类型。
type CleanupFunc func() error

// Cleaner 管理运行期资源清理函数。
type Cleaner struct {
	mu       sync.Mutex
	cleanups []CleanupFunc
}

// NewCleaner 创建新的资源清理器。
func NewCleaner() *Cleaner {
	return &Cleaner{}
}

// Add 添加一个清理函数。
func (c *Cleaner) Add(cleanup CleanupFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cleanups = append(c.cleanups, cleanup)
}

// ExecuteAndClear 执行所有清理函数（后进先出）并返回第一个错误。
func (c *Cleaner) ExecuteAndClear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var firstErr error
	for i := len(c.cleanups) - 1; i >= 0; i-- {
		if err := c.safeExecute(c.cleanups[i]); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	c.cleanups = nil
	return firstErr
}

func (c *Cleaner) safeExecute(cleanup CleanupFunc) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic during cleanup: %v", r)
		}
	}()
	return cleanup()
}
