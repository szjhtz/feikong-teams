package toolmeta

import (
	"context"
	"strings"
	"sync"
	"unicode"
)

const AgentToolPrefix = "ask_fkagent_"

const (
	ToolKindAgent = "agent"
	ToolKindTool  = "tool"
)

type registryKey struct{}

// ToolDisplay 描述工具在入口层展示时需要的稳定元信息。
type ToolDisplay struct {
	Name        string
	DisplayName string
	Kind        string
	Target      string
}

// Resolver 只暴露展示解析能力，用于历史与入口展示层。
type Resolver interface {
	FormatToolDisplay(name string) ToolDisplay
}

// Registry 保存单个应用实例内的工具展示元信息。
type Registry struct {
	agentTools sync.Map
}

// NewRegistry 创建空展示元信息注册表。
func NewRegistry() *Registry {
	return &Registry{}
}

// WithRegistry 将工具展示注册表绑定到 context。
func WithRegistry(ctx context.Context, registry *Registry) context.Context {
	if registry == nil {
		return ctx
	}
	return context.WithValue(ctx, registryKey{}, registry)
}

// RegistryFromContext 从 context 获取工具展示注册表。
func RegistryFromContext(ctx context.Context) (*Registry, bool) {
	if ctx == nil {
		return nil, false
	}
	registry, ok := ctx.Value(registryKey{}).(*Registry)
	return registry, ok && registry != nil
}

// ResolverFromContext 从 context 获取展示解析器；未注入时返回无状态 fallback。
func ResolverFromContext(ctx context.Context) Resolver {
	if registry, ok := RegistryFromContext(ctx); ok {
		return registry
	}
	return fallbackResolver{}
}

// RegisterAgentToolDisplay 注册成员智能体工具的展示名。
func (r *Registry) RegisterAgentToolDisplay(toolName, displayName string) {
	if r == nil {
		return
	}
	if toolName == "" {
		return
	}
	target := displayName
	if target == "" {
		target = titleIdentifier(strings.TrimPrefix(toolName, AgentToolPrefix))
	}
	r.agentTools.Store(toolName, ToolDisplay{
		Name:        toolName,
		DisplayName: "指派给 " + target,
		Kind:        ToolKindAgent,
		Target:      target,
	})
}

// FormatToolDisplay 解析工具展示信息。
func (r *Registry) FormatToolDisplay(name string) ToolDisplay {
	if r != nil {
		if value, ok := r.agentTools.Load(name); ok {
			return value.(ToolDisplay)
		}
	}
	return FallbackDisplay(name)
}

// FallbackDisplay 返回无注册表时的确定性展示信息。
func FallbackDisplay(name string) ToolDisplay {
	display := ToolDisplay{
		Name:        name,
		DisplayName: name,
		Kind:        ToolKindTool,
	}
	return display
}

// FormatToolDisplay 返回无状态 fallback 展示信息。
func FormatToolDisplay(name string) ToolDisplay {
	return FallbackDisplay(name)
}

type fallbackResolver struct{}

func (fallbackResolver) FormatToolDisplay(name string) ToolDisplay {
	return FallbackDisplay(name)
}

func titleIdentifier(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-'
	})
	for i, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(strings.ToLower(part))
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}
	return strings.Join(parts, " ")
}
