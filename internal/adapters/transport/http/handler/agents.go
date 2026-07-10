package handler

import "github.com/gin-gonic/gin"

// AgentInfoResponse 智能体信息响应
type AgentInfoResponse struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name,omitempty"`
	Description string   `json:"description"`
	Aliases     []string `json:"aliases,omitempty"`
	Builtin     bool     `json:"builtin,omitempty"`
	TeamMember  bool     `json:"team_member,omitempty"`
	Prompt      string   `json:"prompt,omitempty"`
	ModelID     string   `json:"model_id,omitempty"`
	Tools       []string `json:"tools,omitempty"`
}

// GetAgentsHandler 获取所有可用智能体

func (rt *Runtime) GetAgentsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if rt.AgentRegistry == nil {
			Fail(c, 503, "agent registry is not configured")
			return
		}
		registry := rt.AgentRegistry.List()

		agentList := make([]AgentInfoResponse, 0, len(registry))
		for _, agent := range registry {
			agentList = append(agentList, AgentInfoResponse{
				Name:        agent.Name,
				DisplayName: agent.DisplayName,
				Description: agent.Description,
				Aliases:     agent.Aliases,
				Builtin:     agent.Builtin,
				TeamMember:  agent.TeamMember,
				Prompt:      agent.Prompt,
				ModelID:     agent.ModelID,
				Tools:       agent.ToolNames,
			})
		}

		OK(c, agentList)
	}
}
