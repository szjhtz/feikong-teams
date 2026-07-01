package eventlog

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	appchat "fkteams/internal/app/chat"
)

type ChatSessionStore struct {
	sessionsDir string
}

func NewChatSessionStore(sessionsDir string) *ChatSessionStore {
	return &ChatSessionStore{sessionsDir: sessionsDir}
}

func (s *ChatSessionStore) SaveHistory(_ context.Context, sessionID string, history appchat.SessionHistory) error {
	saver, ok := history.(interface{ SaveToFile(string) error })
	if !ok {
		return fmt.Errorf("history does not support file persistence")
	}
	return saver.SaveToFile(filepath.Join(s.sessionDir(sessionID), TranscriptFileName))
}

func (s *ChatSessionStore) UpdateMetadata(_ context.Context, update appchat.MetadataUpdate) error {
	sessionDir := s.sessionDir(update.SessionID)
	now := time.Now()
	meta, err := LoadMetadata(sessionDir)
	if err != nil {
		if !update.CreateIfMissing {
			return nil
		}
		meta = &SessionMetadata{
			ID:        update.SessionID,
			Title:     titleFromSource(update.TitleSource, update.DefaultTitle),
			Status:    update.Status,
			CreatedAt: now,
			UpdatedAt: now,
		}
	} else {
		meta.UpdatedAt = now
		if update.Status != "" {
			meta.Status = update.Status
		}
		if update.UpdateDefaultTitle && update.TitleSource != "" && isDefaultTitle(meta.Title) {
			meta.Title = truncateTitle(update.TitleSource)
		}
	}
	return SaveMetadata(sessionDir, meta)
}

func (s *ChatSessionStore) sessionDir(sessionID string) string {
	return filepath.Join(s.sessionsDir, filepath.Base(sessionID))
}

func titleFromSource(source, fallback string) string {
	if source != "" {
		return truncateTitle(source)
	}
	if fallback != "" {
		return fallback
	}
	return "未命名会话"
}

func isDefaultTitle(title string) bool {
	if title == "" || title == "未命名会话" {
		return true
	}
	_, err := time.Parse("2006-01-02 15:04:05", title)
	return err == nil
}

func truncateTitle(s string) string {
	const maxLen = 50
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
