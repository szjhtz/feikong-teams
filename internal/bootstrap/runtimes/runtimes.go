package runtimes

import (
	einoengine "fkteams/internal/adapters/runtime/eino/engine"
	toolmcp "fkteams/internal/adapters/tools/mcp"
	toolport "fkteams/internal/ports/tools"
	runtimeregistry "fkteams/internal/runtime/registry"

	_ "fkteams/internal/adapters/runtime/eino/providers/register"
)

func init() {
	engine := einoengine.NewEngine()
	runtimeregistry.Register(runtimeregistry.DefaultRuntimeName, engine)
	if provider, ok := any(engine).(toolport.MCPClientToolProvider); ok {
		toolmcp.RegisterToolProvider(provider.MCPTools)
	}
}
