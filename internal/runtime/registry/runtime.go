package registry

import (
	runtimeport "fkteams/internal/ports/runtime"
	"fmt"
	"sort"
	"sync"
)

const DefaultRuntimeName = "eino"

var registry = struct {
	sync.RWMutex
	defaultName string
	engines     map[string]runtimeport.Engine
}{
	defaultName: DefaultRuntimeName,
	engines:     make(map[string]runtimeport.Engine),
}

func Engine() runtimeport.Engine {
	engine, err := EngineByName(DefaultName())
	if err != nil {
		panic(err)
	}
	return engine
}

func Register(name string, engine runtimeport.Engine) {
	if name == "" {
		panic("runtime name is empty")
	}
	if engine == nil {
		panic("runtime engine is nil")
	}
	registry.Lock()
	registry.engines[name] = engine
	registry.Unlock()
}

func Use(name string) error {
	registry.Lock()
	defer registry.Unlock()
	if _, ok := registry.engines[name]; !ok {
		return fmt.Errorf("runtime %s is not registered", name)
	}
	registry.defaultName = name
	return nil
}

func DefaultName() string {
	registry.RLock()
	defer registry.RUnlock()
	return registry.defaultName
}

func EngineByName(name string) (runtimeport.Engine, error) {
	registry.RLock()
	defer registry.RUnlock()
	engine, ok := registry.engines[name]
	if !ok {
		return nil, fmt.Errorf("runtime %s is not registered", name)
	}
	return engine, nil
}

func RegisteredNames() []string {
	registry.RLock()
	defer registry.RUnlock()
	names := make([]string, 0, len(registry.engines))
	for name := range registry.engines {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
