package eventlog

import (
	"encoding/json"
	"errors"
	domainsession "fkteams/internal/domain/session"
	"fkteams/internal/runtime/atomicfile"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// SessionMetadata 保留历史存储包的兼容名称。
type SessionMetadata = domainsession.Metadata

// SaveMetadata 保存会话元数据到指定目录
func SaveMetadata(sessionDir string, meta *SessionMetadata) error {
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return fmt.Errorf("create session dir: %w", err)
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}
	return atomicfile.WriteFile(filepath.Join(sessionDir, "metadata.json"), data, 0644)
}

// LoadMetadata 从指定目录加载会话元数据
func LoadMetadata(sessionDir string) (*SessionMetadata, error) {
	data, err := os.ReadFile(filepath.Join(sessionDir, "metadata.json"))
	if err != nil {
		return nil, err
	}
	var meta SessionMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("unmarshal metadata: %w", err)
	}
	return &meta, nil
}

// SessionHistoryManager 按会话 ID 管理独立的 HistoryRecorder，支持并发安全
type SessionHistoryManager struct {
	mu       sync.RWMutex
	sessions map[string]*HistoryRecorder
}

func NewSessionHistoryManager() *SessionHistoryManager {
	return &SessionHistoryManager{
		sessions: make(map[string]*HistoryRecorder),
	}
}

// GetOrCreate 获取或创建会话的 HistoryRecorder，不存在时尝试从 transcript 加载
func (m *SessionHistoryManager) GetOrCreate(sessionID, historyDir string) *HistoryRecorder {
	if !domainsession.ValidID(sessionID) {
		recorder := NewHistoryRecorder()
		recorder.persistErr = fmt.Errorf("invalid session ID")
		return recorder
	}
	m.mu.RLock()
	if recorder, exists := m.sessions[sessionID]; exists {
		m.mu.RUnlock()
		return recorder
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	if recorder, exists := m.sessions[sessionID]; exists {
		return recorder
	}

	sessionDir := filepath.Join(historyDir, sessionID)
	recorder := NewHistoryRecorder()
	recorder.SetSessionDir(sessionDir)
	filePath := filepath.Join(sessionDir, TranscriptFileName)
	if err := recorder.LoadFromFile(filePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		recorder.mu.Lock()
		recorder.persistErr = fmt.Errorf("load existing transcript: %w", err)
		recorder.mu.Unlock()
	}

	m.sessions[sessionID] = recorder
	return recorder
}

func (m *SessionHistoryManager) Get(sessionID string) *HistoryRecorder {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[sessionID]
}

func (m *SessionHistoryManager) Remove(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
}

// Clear 清空指定会话的历史（不移除 Recorder 实例）
func (m *SessionHistoryManager) Clear(sessionID string) {
	m.mu.RLock()
	recorder, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if exists {
		recorder.Clear()
	}
}

// LoadForSession 从 transcript 加载历史并替换指定会话的 Recorder
func (m *SessionHistoryManager) LoadForSession(sessionID, filePath string) (*HistoryRecorder, error) {
	if !domainsession.ValidID(sessionID) {
		return nil, fmt.Errorf("invalid session ID")
	}
	recorder := NewHistoryRecorder()
	recorder.SetSessionDir(filepath.Dir(filePath))
	if err := recorder.LoadFromFile(filePath); err != nil {
		return nil, fmt.Errorf("load session %s: %w", sessionID, err)
	}

	m.mu.Lock()
	m.sessions[sessionID] = recorder
	m.mu.Unlock()

	return recorder, nil
}

func (m *SessionHistoryManager) SaveSession(sessionID, filePath string) error {
	if !domainsession.ValidID(sessionID) {
		return fmt.Errorf("invalid session ID")
	}
	m.mu.RLock()
	recorder, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return recorder.SaveToFile(filePath)
}
