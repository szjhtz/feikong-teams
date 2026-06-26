package runtimes

import (
	modelproviders "fkteams/internal/adapters/model/providers"
	einoruntime "fkteams/internal/adapters/runtime/eino"
	einoengine "fkteams/internal/adapters/runtime/eino/engine"
	einoproviders "fkteams/internal/adapters/runtime/eino/providers/register"
	toolmcp "fkteams/internal/adapters/tools/mcp"
	runtimeport "fkteams/internal/ports/runtime"
	modelregistry "fkteams/internal/runtime/model"
	runtimeregistry "fkteams/internal/runtime/registry"
)

// Defaults 保存组合根创建的默认 runtime 依赖。
type Defaults struct {
	RuntimeRegistry       *runtimeregistry.Registry
	Engine                runtimeport.Engine
	Interrupt             runtimeport.InterruptRuntime
	ModelRegistry         *modelregistry.Registry
	ModelProviderRegistry *modelproviders.Registry
}

// NewDefaults 显式创建默认 runtime adapter 和关联桥接能力。
func NewDefaults() (*Defaults, error) {
	providerRegistry := modelproviders.NewRegistry()
	modelRegistry := modelregistry.NewRegistry()
	einoproviders.RegisterDefaults(providerRegistry, modelRegistry)

	engine := einoengine.NewEngine()
	runtimeRegistry := runtimeregistry.NewRegistry(runtimeregistry.DefaultRuntimeName)
	if err := runtimeRegistry.Register(runtimeregistry.DefaultRuntimeName, engine); err != nil {
		return nil, err
	}
	if err := runtimeRegistry.Use(runtimeregistry.DefaultRuntimeName); err != nil {
		return nil, err
	}

	// MCP tool provider 桥接仍由 MCP adapter 持有，组合根负责唯一装配点。
	toolmcp.RegisterToolProvider(engine.MCPTools)

	return &Defaults{
		RuntimeRegistry:       runtimeRegistry,
		Engine:                engine,
		Interrupt:             einoruntime.NewInterruptRuntime(),
		ModelRegistry:         modelRegistry,
		ModelProviderRegistry: providerRegistry,
	}, nil
}
