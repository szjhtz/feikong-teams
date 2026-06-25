package mcp

import (
	"context"
	"fmt"
	"sync"

	runtimeport "fkteams/internal/ports/runtime"
	toolport "fkteams/internal/ports/tools"
)

type ToolProvider func(context.Context, any) ([]runtimeport.Tool, error)

type Provider struct {
	mu           sync.RWMutex
	cachedGroups toolport.MCPToolGroups
	toolProvider ToolProvider
}

func NewProvider() *Provider {
	return &Provider{}
}

var defaultProvider = NewProvider()

func DefaultProvider() *Provider {
	return defaultProvider
}

func RegisterToolProvider(provider ToolProvider) {
	defaultProvider.RegisterToolProvider(provider)
}

func (p *Provider) RegisterToolProvider(provider ToolProvider) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.toolProvider = provider
	p.cachedGroups = nil
}

func (p *Provider) GetToolsByName(ctx context.Context, groupName string) ([]runtimeport.Tool, error) {
	groups, err := p.GetAllToolGroups(ctx)
	if err != nil {
		return nil, err
	}
	if group, exists := groups[groupName]; exists {
		return group.Tools, nil
	}
	return nil, fmt.Errorf("MCP tool %s not found", groupName)
}

func (p *Provider) GetAllToolGroups(ctx context.Context) (toolport.MCPToolGroups, error) {
	p.mu.RLock()
	cached := p.cachedGroups
	p.mu.RUnlock()
	if cached != nil {
		return cloneGroups(cached), nil
	}

	groups, err := p.loadToolGroups(ctx)
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	p.cachedGroups = cloneGroups(groups)
	p.mu.Unlock()
	return cloneGroups(groups), nil
}

func (p *Provider) ClearCache() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cachedGroups = nil
}

func (p *Provider) loadToolGroups(ctx context.Context) (toolport.MCPToolGroups, error) {
	clients, err := setupMCPClients(ctx)
	if err != nil {
		return nil, err
	}
	if len(clients) == 0 {
		return toolport.MCPToolGroups{}, nil
	}

	p.mu.RLock()
	toolProvider := p.toolProvider
	p.mu.RUnlock()
	if toolProvider == nil {
		return nil, fmt.Errorf("MCP tool provider is not registered")
	}

	groups := make(toolport.MCPToolGroups, len(clients))
	for _, mcpClient := range clients {
		tools, err := toolProvider(ctx, mcpClient.Client)
		if err != nil {
			return nil, fmt.Errorf("failed to get tools from MCP server %s: %v", mcpClient.Name, err)
		}
		groups[mcpClient.Name] = toolport.MCPToolGroup{
			Name:  mcpClient.Name,
			Desc:  mcpClient.Desc,
			Tools: tools,
		}
	}
	return groups, nil
}

func cloneGroups(groups toolport.MCPToolGroups) toolport.MCPToolGroups {
	if groups == nil {
		return nil
	}
	clone := make(toolport.MCPToolGroups, len(groups))
	for name, group := range groups {
		group.Tools = append([]runtimeport.Tool(nil), group.Tools...)
		clone[name] = group
	}
	return clone
}
