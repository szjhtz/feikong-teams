package tools

import (
	"context"
	"fmt"
	"sync"

	runtimeport "fkteams/internal/ports/runtime"
	toolport "fkteams/internal/ports/tools"
)

var (
	mcpProviderMu sync.RWMutex
	mcpProvider   toolport.MCPProvider
)

func RegisterMCPProvider(provider toolport.MCPProvider) {
	mcpProviderMu.Lock()
	defer mcpProviderMu.Unlock()
	mcpProvider = provider
}

func ClearMCPToolCache() {
	if provider := currentMCPProvider(); provider != nil {
		provider.ClearCache()
	}
}

func GetMCPToolsByName(name string) ([]runtimeport.Tool, error) {
	provider := currentMCPProvider()
	if provider == nil {
		return nil, fmt.Errorf("MCP provider is not registered")
	}
	return provider.GetToolsByName(context.Background(), name)
}

func GetAllMCPToolGroups() (toolport.MCPToolGroups, error) {
	provider := currentMCPProvider()
	if provider == nil {
		return nil, fmt.Errorf("MCP provider is not registered")
	}
	return provider.GetAllToolGroups(context.Background())
}

func currentMCPProvider() toolport.MCPProvider {
	mcpProviderMu.RLock()
	defer mcpProviderMu.RUnlock()
	return mcpProvider
}
