package storage

import (
	"context"

	domainhistory "fkteams/internal/domain/history"
)

// SessionMessageReader 读取指定会话的结构化历史消息。
type SessionMessageReader interface {
	LoadSessionMessages(ctx context.Context, sessionID string) ([]domainhistory.AgentMessage, error)
}
