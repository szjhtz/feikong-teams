package agents

import (
	"context"
	"fkteams/internal/app/agent/catalog/analyst"
	assistantagent "fkteams/internal/app/agent/catalog/assistant"
	"fkteams/internal/app/agent/catalog/coder"
	"fkteams/internal/app/agent/catalog/custom"
	"fkteams/internal/app/agent/catalog/researcher"
	"fkteams/internal/app/agent/catalog/visitor"
	"fkteams/internal/app/config"
	runtimeport "fkteams/internal/ports/runtime"
	"log"
	"sync"
)

// AgentInfo 智能体信息
type AgentInfo struct {
	Name        string
	Description string
	Aliases     []string
	Creator     func(ctx context.Context) (runtimeport.Agent, error)
}

var (
	// Registry 智能体注册表
	Registry     []AgentInfo
	registryOnce sync.Once
	registryMu   sync.RWMutex
)

// initRegistry 初始化注册表（延迟加载）
func initRegistry() {
	registryOnce.Do(func() {
		buildRegistry()
	})
}

// buildRegistry 构建智能体注册表
func buildRegistry() {
	cfg := config.Get()

	type agentCreator struct {
		name        string
		description string
		aliases     []string
		creator     func(ctx context.Context) (runtimeport.Agent, error)
	}

	// 基础智能体（始终可用）
	creators := []agentCreator{
		{name: "coder", description: "软件工程师，负责代码实现、调试、重构和工程验证。", aliases: []string{"小码"}, creator: coder.NewAgent},
	}

	// 可选智能体（根据配置文件启用）
	if cfg.Agents.Researcher {
		creators = append(creators, agentCreator{name: "researcher", description: "网络研究员，负责检索、抓取、交叉验证和整理时效信息。", aliases: []string{"小搜"}, creator: researcher.NewAgent})
	}

	if cfg.Agents.Analyst {
		creators = append(creators, agentCreator{name: "analyst", description: "数据分析师，负责使用表格、脚本和文档工具提取洞察。", aliases: []string{"小析"}, creator: analyst.NewAgent})
	}

	if cfg.Agents.SSHVisitor.Enabled {
		creators = append(creators, agentCreator{name: "remote", description: "远程运维专家，负责通过 SSH 管理服务器、执行命令和传输文件。", aliases: []string{"小访", "visitor"}, creator: visitor.NewAgent})
	}

	if cfg.Agents.Assistant {
		creators = append(creators, agentCreator{name: "generalist", description: "通用执行助手，负责综合命令、文件、搜索和文档工具完成开放任务。", aliases: []string{"小助", "assistant"}, creator: assistantagent.NewAgent})
	}

	// 动态构建注册表
	Registry = make([]AgentInfo, 0, len(creators))
	for _, c := range creators {
		creator := c.creator
		Registry = append(Registry, AgentInfo{
			Name:        c.name,
			Description: c.description,
			Aliases:     c.aliases,
			Creator:     creator,
		})
	}

	// 加载配置文件中的自定义智能体
	loadCustomAgents()
}

// loadCustomAgents 从配置文件加载自定义智能体并添加到注册表
func loadCustomAgents() {
	cfg := config.Get()

	if len(cfg.Custom.Agents) == 0 {
		return
	}

	existingNames := make(map[string]bool, len(Registry))
	for _, info := range Registry {
		existingNames[info.Name] = true
	}

	for _, agentCfg := range cfg.Custom.Agents {
		if agentCfg.Name == "" {
			continue
		}

		if existingNames[agentCfg.Name] {
			log.Printf("[agent] 自定义智能体 %q 与已有智能体名称重复，不建议使用相同名称", agentCfg.Name)
		}

		agentCfg := agentCfg
		Registry = append(Registry, AgentInfo{
			Name:        agentCfg.Name,
			Description: agentCfg.Desc,
			Creator: func(ctx context.Context) (runtimeport.Agent, error) {
				mc := config.Get().ResolveModel(agentCfg.Model)
				var model custom.Model
				if mc != nil {
					model = custom.Model{
						Provider: mc.Provider,
						Name:     mc.Model,
						APIKey:   mc.APIKey,
						BaseURL:  mc.BaseURL,
					}
				}
				return custom.NewAgent(ctx, custom.Config{
					Name:         agentCfg.Name,
					Description:  agentCfg.Desc,
					SystemPrompt: agentCfg.SystemPrompt,
					Model:        model,
					ToolNames:    agentCfg.Tools,
				})
			},
		})
	}
}

// GetRegistry 获取智能体注册表
func GetRegistry() []AgentInfo {
	initRegistry()
	registryMu.RLock()
	defer registryMu.RUnlock()
	result := make([]AgentInfo, len(Registry))
	copy(result, Registry)
	return result
}

// GetAgentByName 根据名字获取智能体
func GetAgentByName(name string) *AgentInfo {
	initRegistry()
	registryMu.RLock()
	defer registryMu.RUnlock()
	for i := range Registry {
		if Registry[i].Name == name {
			return &Registry[i]
		}
		for _, alias := range Registry[i].Aliases {
			if alias == name {
				return &Registry[i]
			}
		}
	}
	return nil
}

// GetTeamAgents 获取团队模式的智能体列表
func GetTeamAgents(ctx context.Context) ([]runtimeport.Agent, error) {
	initRegistry()
	registryMu.RLock()
	defer registryMu.RUnlock()

	var subAgents []runtimeport.Agent
	for _, agentInfo := range Registry {
		agent, err := agentInfo.Creator(ctx)
		if err != nil {
			return nil, err
		}
		subAgents = append(subAgents, agent)
	}

	return subAgents, nil
}

// ReloadRegistry 重新构建智能体注册表（配置变更后调用）
func ReloadRegistry() {
	registryMu.Lock()
	defer registryMu.Unlock()
	buildRegistry()
}
