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
	Runtime               runtimeport.Runtime
	Interrupt             runtimeport.InterruptRuntime
	ModelRegistry         *modelregistry.Registry
	ModelProviderRegistry *modelproviders.Registry
}

// Options 描述 runtime 组合根的显式外部依赖。
type Options struct {
	MCPProvider *toolmcp.Provider
}

// NewDefaults 显式创建默认 runtime adapter 和关联桥接能力。
func NewDefaults(options ...Options) (*Defaults, error) {
	var opt Options
	if len(options) > 0 {
		opt = options[0]
	}

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

	if opt.MCPProvider != nil {
		opt.MCPProvider.RegisterToolProvider(engine.MCPTools)
	}

	return &Defaults{
		RuntimeRegistry:       runtimeRegistry,
		Runtime:               engine,
		Interrupt:             einoruntime.NewInterruptRuntime(),
		ModelRegistry:         modelRegistry,
		ModelProviderRegistry: providerRegistry,
	}, nil
}
