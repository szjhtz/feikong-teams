package mcp

import (
	"context"
	"strings"
	"testing"

	"fkteams/internal/app/config"
	runtimeport "fkteams/internal/ports/runtime"
	toolport "fkteams/internal/ports/tools"
)

type fakeTool struct {
	name string
}

func (f fakeTool) Info(context.Context) (*runtimeport.ToolInfo, error) {
	return &runtimeport.ToolInfo{Name: f.name}, nil
}

func (f fakeTool) Invoke(context.Context, runtimeport.ToolInvocation) (*runtimeport.ToolResult, error) {
	return &runtimeport.ToolResult{Content: f.name}, nil
}

func TestProviderUsesCacheAndClearsCache(t *testing.T) {
	provider := NewProvider()
	provider.cachedGroups = toolport.MCPToolGroups{
		"demo": {
			Name:  "demo",
			Desc:  "Demo tools",
			Tools: []runtimeport.Tool{fakeTool{name: "demo_tool"}},
		},
	}

	tools, err := provider.GetToolsByName(context.Background(), "demo")
	if err != nil {
		t.Fatalf("GetToolsByName returned error: %v", err)
	}
	if len(tools) != 1 {
		t.Fatalf("tool count = %d, want 1", len(tools))
	}
	info, err := tools[0].Info(context.Background())
	if err != nil {
		t.Fatalf("tool info: %v", err)
	}
	if info.Name != "demo_tool" {
		t.Fatalf("tool name = %q, want demo_tool", info.Name)
	}

	if _, err := provider.GetToolsByName(context.Background(), "missing"); err == nil || !strings.Contains(err.Error(), "MCP tool missing not found") {
		t.Fatalf("missing tool error = %v", err)
	}

	provider.ClearCache()
	if provider.cachedGroups != nil {
		t.Fatalf("cachedGroups = %#v, want nil", provider.cachedGroups)
	}
}

func TestProviderReturnsEmptyGroupsWithNoEnabledServers(t *testing.T) {
	t.Setenv("FEIKONG_APP_DIR", t.TempDir())
	if err := config.Save(&config.Config{
		Custom: config.Custom{
			MCPServers: []config.MCPServer{
				{ID: "disabled", Name: "Disabled", Enabled: false, Transport: "stdio"},
			},
		},
	}); err != nil {
		t.Fatalf("save config: %v", err)
	}

	groups, err := NewProvider().GetAllToolGroups(context.Background())
	if err != nil {
		t.Fatalf("GetAllToolGroups returned error: %v", err)
	}
	if len(groups) != 0 {
		t.Fatalf("groups = %#v, want empty", groups)
	}
}

func TestSetupMCPClientsRejectsUnsupportedTransport(t *testing.T) {
	t.Setenv("FEIKONG_APP_DIR", t.TempDir())
	if err := config.Save(&config.Config{
		Custom: config.Custom{
			MCPServers: []config.MCPServer{
				{ID: "bad", Name: "Bad", Enabled: true, Transport: "pipe"},
			},
		},
	}); err != nil {
		t.Fatalf("save config: %v", err)
	}

	clients, err := setupMCPClients(context.Background())
	if err == nil {
		t.Fatalf("setupMCPClients = %#v, want unsupported transport error", clients)
	}
	if !strings.Contains(err.Error(), "unsupported MCP transport type") {
		t.Fatalf("setupMCPClients error = %v", err)
	}
}
