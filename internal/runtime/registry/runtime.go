package registry

import (
	"fmt"
	"sort"
	"sync"

	runtimeport "fkteams/internal/ports/runtime"
)

const DefaultRuntimeName = "eino"

// Registry 保存一组可用 runtime engine。
type Registry struct {
	mu          sync.RWMutex
	defaultName string
	engines     map[string]runtimeport.Engine
}

// NewRegistry 创建 runtime engine 注册表。
func NewRegistry(defaultName string) *Registry {
	if defaultName == "" {
		defaultName = DefaultRuntimeName
	}
	return &Registry{
		defaultName: defaultName,
		engines:     make(map[string]runtimeport.Engine),
	}
}

// Engine 返回当前默认 runtime engine。
func (r *Registry) Engine() (runtimeport.Engine, error) {
	if r == nil {
		return nil, fmt.Errorf("runtime registry is nil")
	}
	return r.EngineByName(r.DefaultName())
}

// Register 注册 runtime engine。
func (r *Registry) Register(name string, engine runtimeport.Engine) error {
	if r == nil {
		return fmt.Errorf("runtime registry is nil")
	}
	if name == "" {
		return fmt.Errorf("runtime name is empty")
	}
	if engine == nil {
		return fmt.Errorf("runtime engine is nil")
	}
	r.mu.Lock()
	r.engines[name] = engine
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
	if _, ok := r.engines[name]; !ok {
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

// EngineByName 返回指定 runtime engine。
func (r *Registry) EngineByName(name string) (runtimeport.Engine, error) {
	if r == nil {
		return nil, fmt.Errorf("runtime registry is nil")
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	engine, ok := r.engines[name]
	if !ok {
		return nil, fmt.Errorf("runtime %s is not registered", name)
	}
	return engine, nil
}

// RegisteredNames 返回已注册 runtime 名称。
func (r *Registry) RegisteredNames() []string {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.engines))
	for name := range r.engines {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
