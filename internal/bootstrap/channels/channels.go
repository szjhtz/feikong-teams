package channels

import (
	channel "fkteams/internal/adapters/transport/channel"
	"fkteams/internal/adapters/transport/channel/discord"
	"fkteams/internal/adapters/transport/channel/qq"
	"fkteams/internal/adapters/transport/channel/weixin"
)

// RegisterDefaults 注册内置消息通道工厂。
func RegisterDefaults() *channel.FactoryRegistry {
	registry := channel.NewFactoryRegistry()
	discord.Register(registry)
	qq.Register(registry)
	weixin.Register(registry)
	return registry
}
