package eino

import (
	"context"
	"fkteams/agentcore"
	"fmt"

	"github.com/cloudwego/eino/adk"
)

func WrapAgent(agent adk.Agent) agentcore.Agent {
	if agent == nil {
		return nil
	}
	ctx := context.Background()
	return agentcore.WrapRuntimeAgent(agent.Name(ctx), agent.Description(ctx), agent)
}

func WrapNamedAgent(name, description string, agent adk.Agent) agentcore.Agent {
	return agentcore.WrapRuntimeAgent(name, description, agent)
}

func AdaptAgentForRunner(agent agentcore.Agent) (adk.Agent, error) {
	if agent == nil || agent.RuntimeAgent() == nil {
		return nil, fmt.Errorf("agent is nil")
	}
	runnerAgent, ok := agent.RuntimeAgent().(adk.Agent)
	if !ok {
		return nil, fmt.Errorf("unsupported runtime agent: %T", agent.RuntimeAgent())
	}
	return runnerAgent, nil
}

func AdaptAgentsForRunner(agents []agentcore.Agent) ([]adk.Agent, error) {
	result := make([]adk.Agent, 0, len(agents))
	for _, agent := range agents {
		runnerAgent, err := AdaptAgentForRunner(agent)
		if err != nil {
			return nil, err
		}
		result = append(result, runnerAgent)
	}
	return result, nil
}
