package mcp

import (
	"context"
	"fmt"
	"log"
	"time"

	"fkteams/internal/app/config"

	"github.com/mark3labs/mcp-go/client"
	mcpsdk "github.com/mark3labs/mcp-go/mcp"
)

type MCPClient struct {
	Name   string
	Desc   string
	Client *client.Client
}

func setupMCPClients(ctx context.Context) ([]MCPClient, error) {
	cfg := config.Get()
	clients := make([]MCPClient, 0, len(cfg.Custom.MCPServers))

	for _, server := range cfg.Custom.MCPServers {
		if !server.Enabled {
			log.Printf("Skipping disabled MCP server: %s", server.Name)
			continue
		}

		log.Printf("Connecting to MCP server: %s", server.Name)
		mcpClient, err := newClient(server)
		if err != nil {
			return nil, err
		}
		if err := mcpClient.Start(ctx); err != nil {
			return nil, fmt.Errorf("failed to start MCP client for %s: %v", server.Name, err)
		}

		initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		initializeResult, err := mcpClient.Initialize(initCtx, initializeRequest())
		cancel()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize MCP client for %s: %v", server.Name, err)
		}

		log.Printf("Initialized MCP client for %s", server.Name)
		mcpClient.OnNotification(func(notification mcpsdk.JSONRPCNotification) {
			log.Printf("Received notification from MCP server %s: %+v", server.Name, notification)
		})

		desc := server.Desc
		if initializeResult.Instructions != "" {
			desc = initializeResult.Instructions
		}
		clients = append(clients, MCPClient{
			Name:   server.Name,
			Desc:   desc,
			Client: mcpClient,
		})
	}

	return clients, nil
}

func newClient(server config.MCPServer) (*client.Client, error) {
	switch server.TransportType {
	case "http":
		return client.NewStreamableHttpClient(server.URL)
	case "sse":
		return client.NewSSEMCPClient(server.URL)
	case "stdio":
		return client.NewStdioMCPClient(server.Command, server.EnvVars, server.Args...)
	default:
		return nil, fmt.Errorf("unsupported MCP transport type for %s: %s", server.Name, server.TransportType)
	}
}

func initializeRequest() mcpsdk.InitializeRequest {
	req := mcpsdk.InitializeRequest{}
	req.Params.ProtocolVersion = mcpsdk.LATEST_PROTOCOL_VERSION
	req.Params.ClientInfo = mcpsdk.Implementation{
		Name:    "fkteams",
		Version: "1.0.0",
	}
	req.Params.Capabilities = mcpsdk.ClientCapabilities{}
	return req
}
