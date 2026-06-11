package tools

import (
	"fkteams/agentcore"
	"fkteams/common"
	"fkteams/tools/attachment"
	"fmt"
)

type BuiltinCapability struct {
	Name     string
	Provider func(cleaner *common.ResourceCleaner) ([]agentcore.Tool, error)
}

var builtinCapabilities = []BuiltinCapability{
	{
		Name: "session_attachment",
		Provider: func(*common.ResourceCleaner) ([]agentcore.Tool, error) {
			return attachment.GetTools()
		},
	},
}

func BuiltinCapabilityNames() []string {
	names := make([]string, 0, len(builtinCapabilities))
	for _, capability := range builtinCapabilities {
		names = append(names, capability.Name)
	}
	return names
}

func GetBuiltinCapabilityTools() ([]agentcore.Tool, error) {
	return GetBuiltinCapabilityToolsWithCleaner(nil)
}

func GetBuiltinCapabilityToolsWithCleaner(cleaner *common.ResourceCleaner) ([]agentcore.Tool, error) {
	var result []agentcore.Tool
	for _, capability := range builtinCapabilities {
		if capability.Provider == nil {
			return nil, fmt.Errorf("builtin capability %s provider is nil", capability.Name)
		}
		resolved, err := capability.Provider(cleaner)
		if err != nil {
			return nil, fmt.Errorf("init builtin capability %s: %w", capability.Name, err)
		}
		if err := MarkPolicyRequired(resolved); err != nil {
			return nil, fmt.Errorf("mark builtin capability %s policy: %w", capability.Name, err)
		}
		result = append(result, resolved...)
	}
	return result, nil
}
