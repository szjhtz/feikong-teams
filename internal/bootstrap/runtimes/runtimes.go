package runtimes

import (
	einoengine "fkteams/internal/adapters/runtime/eino/engine"
	runtimeport "fkteams/internal/ports/runtime"
	runtimeregistry "fkteams/internal/runtime/registry"
	toolmcp "fkteams/tools/mcp"

	_ "fkteams/internal/adapters/runtime/eino/providers/register"
)

func init() {
	engine := einoengine.NewEngine()
	runtimeregistry.Register(runtimeregistry.DefaultRuntimeName, engine)
	if provider, ok := any(engine).(runtimeport.MCPToolProvider); ok {
		toolmcp.RegisterToolProvider(provider.MCPTools)
	}
}
