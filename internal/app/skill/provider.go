package skill

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// SkillResult 统一的技能搜索结果
type SkillResult struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	DescZh      string `json:"description_zh,omitempty"`
	Owner       string `json:"owner"`
	Homepage    string `json:"homepage"`
	Version     string `json:"version"`
	Downloads   int    `json:"downloads"`
	Stars       int    `json:"stars"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Skills []SkillResult `json:"skills"`
	Total  int           `json:"total"`
}

// Provider 技能后端接口
type Provider interface {
	Name() string
	Search(ctx context.Context, keyword string, page, pageSize int, sortBy, order string) (*SearchResponse, error)
	Download(ctx context.Context, slug, version string) (io.ReadCloser, error)
}

// ProviderRegistry 保存技能市场后端实例。
type ProviderRegistry struct {
	providers []Provider
}

// NewProviderRegistry 创建技能市场后端注册表。
func NewProviderRegistry(providers ...Provider) *ProviderRegistry {
	registry := &ProviderRegistry{}
	for _, provider := range providers {
		if provider != nil {
			registry.providers = append(registry.providers, provider)
		}
	}
	return registry
}

// NewDefaultProviderRegistry 创建内置技能市场后端注册表。
func NewDefaultProviderRegistry() *ProviderRegistry {
	return NewProviderRegistry(NewSkillHubProvider("https://lightmake.site/api/skills"))
}

// Providers 返回所有后端。
func (r *ProviderRegistry) Providers() []Provider {
	if r == nil || len(r.providers) == 0 {
		return nil
	}
	providers := make([]Provider, len(r.providers))
	copy(providers, r.providers)
	return providers
}

// DefaultProvider 返回默认后端。
func (r *ProviderRegistry) DefaultProvider() Provider {
	if r == nil || len(r.providers) == 0 {
		return nil
	}
	return r.providers[0]
}

// ProviderByName 按名称查找后端（不区分大小写）。
func (r *ProviderRegistry) ProviderByName(name string) Provider {
	if r == nil {
		return nil
	}
	for _, p := range r.providers {
		if strings.EqualFold(p.Name(), name) {
			return p
		}
	}
	return nil
}

// ProvidersByNames 按名称列表查找后端，返回匹配的后端列表。
// 如果 names 为空，返回所有后端。
func (r *ProviderRegistry) ProvidersByNames(names []string) ([]Provider, error) {
	if len(names) == 0 {
		return r.Providers(), nil
	}
	var result []Provider
	for _, name := range names {
		p := r.ProviderByName(name)
		if p == nil {
			return nil, fmt.Errorf("skill provider not found: %s", name)
		}
		result = append(result, p)
	}
	return result, nil
}

// Names 返回所有后端名称。
func (r *ProviderRegistry) Names() []string {
	if r == nil || len(r.providers) == 0 {
		return nil
	}
	names := make([]string, len(r.providers))
	for i, p := range r.providers {
		names[i] = p.Name()
	}
	return names
}
