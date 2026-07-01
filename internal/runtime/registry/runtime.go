package registry

import (
	"fmt"
	"sort"
	"sync"

	runtimeport "fkteams/internal/ports/runtime"
)

const DefaultRuntimeName = "eino"

// Registry 保存一组可用 runtime adapter。
type Registry struct {
	mu          sync.RWMutex
	defaultName string
	runtimes    map[string]runtimeport.Runtime
}

// NewRegistry 创建 runtime engine 注册表。
func NewRegistry(defaultName string) *Registry {
	if defaultName == "" {
		defaultName = DefaultRuntimeName
	}
	return &Registry{
		defaultName: defaultName,
		runtimes:    make(map[string]runtimeport.Runtime),
	}
}

// Runtime 返回当前默认 runtime adapter。
func (r *Registry) Runtime() (runtimeport.Runtime, error) {
	if r == nil {
		return nil, fmt.Errorf("runtime registry is nil")
	}
	return r.RuntimeByName(r.DefaultName())
}

// Register 注册 runtime adapter。
func (r *Registry) Register(name string, runtime runtimeport.Runtime) error {
	if r == nil {
		return fmt.Errorf("runtime registry is nil")
	}
	if name == "" {
		return fmt.Errorf("runtime name is empty")
	}
	if runtime == nil {
		return fmt.Errorf("runtime adapter is nil")
	}
	r.mu.Lock()
	r.runtimes[name] = runtime
	r.mu.Unlock()
	return nil
}

// Use 设置默认 runtime engine。
func (r *Registry) Use(name string) error {
	if r == nil {
		return fmt.Errorf("runtime registry is nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.runtimes[name]; !ok {
		return fmt.Errorf("runtime %s is not registered", name)
	}
	r.defaultName = name
	return nil
}

// DefaultName 返回当前默认 runtime 名称。
func (r *Registry) DefaultName() string {
	if r == nil {
		return ""
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.defaultName
}

// RuntimeByName 返回指定 runtime adapter。
func (r *Registry) RuntimeByName(name string) (runtimeport.Runtime, error) {
	if r == nil {
		return nil, fmt.Errorf("runtime registry is nil")
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	runtime, ok := r.runtimes[name]
	if !ok {
		return nil, fmt.Errorf("runtime %s is not registered", name)
	}
	return runtime, nil
}

// RegisteredNames 返回已注册 runtime 名称。
func (r *Registry) RegisteredNames() []string {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.runtimes))
	for name := range r.runtimes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
