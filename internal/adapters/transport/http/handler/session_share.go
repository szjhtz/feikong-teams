package handler

import (
	"encoding/json"
	"errors"
	"fkteams/internal/adapters/storage/file/history"
	"fkteams/internal/app/appdata"
	"fkteams/internal/runtime/atomicfile"
	"fkteams/internal/runtime/log"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const sessionShareFileName = "session_share.json"

// SessionShare 会话分享列表展示信息
type SessionShare struct {
	ID               string `json:"id"`
	SessionID        string `json:"session_id"`
	Title            string `json:"title"`
	HasPassword      bool   `json:"has_password"`
	AllowToolDetails bool   `json:"allow_tool_details"`
	MessageCount     int    `json:"message_count"`
	ExpiresAt        int64  `json:"expires_at"`
	CreatedAt        int64  `json:"created_at"`
	LastAccessedAt   int64  `json:"last_accessed_at,omitempty"`
}

type sessionShareEntry struct {
	SessionID        string    `json:"session_id"`
	Title            string    `json:"title"`
	PasswordHash     string    `json:"password_hash,omitempty"`
	AllowToolDetails bool      `json:"allow_tool_details"`
	MessageCount     int       `json:"message_count"`
	ExpiresAt        time.Time `json:"expires_at,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	LastAccessedAt   time.Time `json:"last_accessed_at,omitempty"`
}

type sessionShareFileEntry struct {
	SessionID        string `json:"session_id"`
	Title            string `json:"title"`
	PasswordHash     string `json:"password_hash,omitempty"`
	AllowToolDetails bool   `json:"allow_tool_details"`
	MessageCount     int    `json:"message_count"`
	ExpiresAt        int64  `json:"expires_at"`
	CreatedAt        int64  `json:"created_at"`
	LastAccessedAt   int64  `json:"last_accessed_at,omitempty"`
}

// SessionShareStore 保存单个 HTTP runtime 的会话分享状态。
type SessionShareStore struct {
	sync.RWMutex
	filePath string
	m        map[string]*sessionShareEntry
	loadErr  error
}

// NewSessionShareStore 创建会话分享存储，并从持久化文件加载现有分享。
func NewSessionShareStore(filePath string) *SessionShareStore {
	if filePath == "" {
		filePath = sessionSharesFilePath()
	}
	store := &SessionShareStore{
		filePath: filePath,
		m:        make(map[string]*sessionShareEntry),
	}
	store.loadErr = store.Load()
	return store
}

func sessionSharesFilePath() string {
	return filepath.Join(appdata.ShareDir(), sessionShareFileName)
}

// Load 从持久化文件加载未过期的会话分享。
func (s *SessionShareStore) Load() error {
	if s == nil {
		return nil
	}
	entries, err := readSessionShareEntries(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("load session shares: %w", err)
	}
	loaded := make(map[string]*sessionShareEntry, len(entries))
	now := time.Now()
	for id, e := range entries {
		var expiresAt time.Time
		if e.ExpiresAt > 0 {
			expiresAt = time.Unix(e.ExpiresAt, 0)
			if now.After(expiresAt) {
				continue
			}
		}
		var lastAccessedAt time.Time
		if e.LastAccessedAt > 0 {
			lastAccessedAt = time.Unix(e.LastAccessedAt, 0)
		}
		loaded[id] = &sessionShareEntry{
			SessionID:        e.SessionID,
			Title:            e.Title,
			PasswordHash:     e.PasswordHash,
			AllowToolDetails: e.AllowToolDetails,
			MessageCount:     e.MessageCount,
			ExpiresAt:        expiresAt,
			CreatedAt:        time.Unix(e.CreatedAt, 0),
			LastAccessedAt:   lastAccessedAt,
		}
	}
	s.Lock()
	s.m = loaded
	s.Unlock()
	return nil
}

func (s *SessionShareStore) LoadError() error {
	if s == nil {
		return nil
	}
	return s.loadErr
}

func readSessionShareEntries(filePath string) (map[string]*sessionShareFileEntry, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var entries map[string]*sessionShareFileEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// Save 将会话分享持久化。
func (s *SessionShareStore) Save() error {
	if s == nil {
		return nil
	}
	return s.SaveTo(s.filePath)
}

// SaveTo 将会话分享写入指定文件。
func (s *SessionShareStore) SaveTo(filePath string) error {
	if s == nil {
		return nil
	}
	s.RLock()
	defer s.RUnlock()
	return s.saveLockedTo(filePath)
}

func (s *SessionShareStore) saveLockedTo(filePath string) error {
	entries := make(map[string]*sessionShareFileEntry, len(s.m))
	for id, e := range s.m {
		entries[id] = &sessionShareFileEntry{
			SessionID:        e.SessionID,
			Title:            e.Title,
			PasswordHash:     e.PasswordHash,
			AllowToolDetails: e.AllowToolDetails,
			MessageCount:     e.MessageCount,
			ExpiresAt:        expiresAtUnix(e.ExpiresAt),
			CreatedAt:        e.CreatedAt.Unix(),
			LastAccessedAt:   expiresAtUnix(e.LastAccessedAt),
		}
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session shares: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("create share dir: %w", err)
	}
	if err := atomicfile.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("write session shares: %w", err)
	}
	return nil
}

func (s *SessionShareStore) Put(id string, entry *sessionShareEntry) error {
	s.Lock()
	defer s.Unlock()
	previous, existed := s.m[id]
	s.m[id] = entry
	if err := s.saveLockedTo(s.filePath); err != nil {
		if existed {
			s.m[id] = previous
		} else {
			delete(s.m, id)
		}
		return err
	}
	return nil
}

func (s *SessionShareStore) Delete(id string) (bool, error) {
	s.Lock()
	defer s.Unlock()
	previous, existed := s.m[id]
	if !existed {
		return false, nil
	}
	delete(s.m, id)
	if err := s.saveLockedTo(s.filePath); err != nil {
		s.m[id] = previous
		return true, err
	}
	return true, nil
}

func (s *SessionShareStore) DeleteMany(ids []string) error {
	s.Lock()
	defer s.Unlock()
	previous := make(map[string]*sessionShareEntry, len(ids))
	for _, id := range ids {
		if entry, ok := s.m[id]; ok {
			previous[id] = entry
			delete(s.m, id)
		}
	}
	if len(previous) == 0 {
		return nil
	}
	if err := s.saveLockedTo(s.filePath); err != nil {
		for id, entry := range previous {
			s.m[id] = entry
		}
		return err
	}
	return nil
}

func (s *SessionShareStore) Touch(id string, at time.Time) error {
	s.Lock()
	defer s.Unlock()
	entry := s.m[id]
	if entry == nil {
		return nil
	}
	previous := entry.LastAccessedAt
	entry.LastAccessedAt = at
	if err := s.saveLockedTo(s.filePath); err != nil {
		entry.LastAccessedAt = previous
		return err
	}
	return nil
}

func sessionShareResponse(id string, entry *sessionShareEntry) SessionShare {
	if entry == nil {
		return SessionShare{ID: id}
	}
	return SessionShare{
		ID:               id,
		SessionID:        entry.SessionID,
		Title:            entry.Title,
		HasPassword:      entry.PasswordHash != "",
		AllowToolDetails: entry.AllowToolDetails,
		MessageCount:     entry.MessageCount,
		ExpiresAt:        expiresAtUnix(entry.ExpiresAt),
		CreatedAt:        entry.CreatedAt.Unix(),
		LastAccessedAt:   expiresAtUnix(entry.LastAccessedAt),
	}
}

func sessionShareMessages(historyDir, sessionID string, allowToolDetails bool) ([]eventlog.AgentMessage, error) {
	if !validateSessionID(sessionID) {
		return nil, errors.New("invalid session ID")
	}
	recorder := eventlog.NewHistoryRecorder()
	sessionDir := sessionDirPath(historyDir, sessionID)
	recorder.SetSessionDir(sessionDir)
	transcriptFile := filepath.Join(sessionDir, eventlog.TranscriptFileName)
	if err := recorder.LoadFromFile(transcriptFile); err != nil {
		return nil, err
	}
	messages := recorder.GetMessages()
	if allowToolDetails {
		return messages, nil
	}
	for msgIndex := range messages {
		events := messages[msgIndex].Events
		for eventIndex := range events {
			if events[eventIndex].ToolCall != nil {
				events[eventIndex].ToolCall.Arguments = ""
				events[eventIndex].ToolCall.Result = ""
			}
			events[eventIndex].Detail = ""
		}
		messages[msgIndex].Events = events
	}
	return messages, nil
}

func sessionShareTranscript(historyDir, sessionID string, allowToolDetails bool) ([]eventlog.TranscriptEvent, error) {
	if !validateSessionID(sessionID) {
		return nil, errors.New("invalid session ID")
	}
	transcriptFile := filepath.Join(sessionDirPath(historyDir, sessionID), eventlog.TranscriptFileName)
	lines, err := eventlog.LoadTranscriptFromFile(transcriptFile)
	if err != nil {
		return nil, err
	}
	if allowToolDetails {
		return lines, nil
	}
	for index := range lines {
		lines[index].Args = ""
		lines[index].Result = ""
		lines[index].ResultRef = ""
		lines[index].Detail = ""
	}
	return lines, nil
}

func (s *SessionShareStore) entryByID(id string) (*sessionShareEntry, bool, bool) {
	s.RLock()
	entry, exists := s.m[id]
	s.RUnlock()
	if !exists {
		return nil, false, false
	}
	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		if _, err := s.Delete(id); err != nil {
			log.Printf("failed to persist expired session share cleanup: id=%s, err=%v", id, err)
		}
		return nil, false, true
	}
	return entry, true, false
}

// CreateSessionShareHandler 创建会话分享

func (rt *Runtime) CreateSessionShareHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			SessionID        string `json:"session_id" binding:"required"`
			Password         string `json:"password"`
			ExpiresIn        int64  `json:"expires_in"`
			AllowToolDetails bool   `json:"allow_tool_details"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, http.StatusBadRequest, "invalid request body")
			return
		}
		if !validateSessionID(req.SessionID) {
			Fail(c, http.StatusBadRequest, "invalid session ID")
			return
		}

		sessionDir := rt.sessionDirPath(req.SessionID)
		meta, metaErr := eventlog.LoadMetadata(sessionDir)
		messages, err := sessionShareMessages(rt.HistoryDir, req.SessionID, req.AllowToolDetails)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) || metaErr != nil {
				Fail(c, http.StatusNotFound, "session history not found")
				return
			}
			log.Printf("failed to load session share history: session=%s, err=%v", req.SessionID, err)
			Fail(c, http.StatusInternalServerError, "failed to read session history")
			return
		}
		if len(messages) == 0 {
			Fail(c, http.StatusBadRequest, "session has no shareable messages")
			return
		}

		expiresIn := req.ExpiresIn
		if expiresIn == 0 {
			expiresIn = 7 * 24 * 3600
		}
		const maxSessionShareExpiry = 90 * 24 * 3600
		if expiresIn > 0 && expiresIn > maxSessionShareExpiry {
			expiresIn = maxSessionShareExpiry
		}

		linkID, err := generateLinkID()
		if err != nil {
			Fail(c, http.StatusInternalServerError, "failed to create share")
			return
		}

		now := time.Now()
		var expiresAt time.Time
		if expiresIn >= 0 {
			expiresAt = now.Add(time.Duration(expiresIn) * time.Second)
		}
		title := req.SessionID
		if metaErr == nil && meta.Title != "" {
			title = meta.Title
		}
		entry := &sessionShareEntry{
			SessionID:        req.SessionID,
			Title:            title,
			AllowToolDetails: req.AllowToolDetails,
			MessageCount:     len(messages),
			ExpiresAt:        expiresAt,
			CreatedAt:        now,
		}
		if req.Password != "" {
			entry.PasswordHash = hashPassword(req.Password)
			if entry.PasswordHash == "" {
				Fail(c, http.StatusInternalServerError, "failed to process password")
				return
			}
		}

		store := rt.SessionShares
		if err := store.Put(linkID, entry); err != nil {
			log.Printf("failed to persist session share: id=%s, err=%v", linkID, err)
			Fail(c, http.StatusInternalServerError, "failed to save session share")
			return
		}

		OK(c, sessionShareResponse(linkID, entry))
	}
}

// ListSessionSharesHandler 列出会话分享

func (rt *Runtime) ListSessionSharesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		expired := make([]string, 0)
		shares := make([]SessionShare, 0)
		store := rt.SessionShares

		store.RLock()
		for id, entry := range store.m {
			if !entry.ExpiresAt.IsZero() && now.After(entry.ExpiresAt) {
				expired = append(expired, id)
				continue
			}
			shares = append(shares, sessionShareResponse(id, entry))
		}
		store.RUnlock()

		if len(expired) > 0 {
			if err := store.DeleteMany(expired); err != nil {
				log.Printf("failed to persist expired session share cleanup: err=%v", err)
			}
		}

		sort.Slice(shares, func(i, j int) bool {
			return shares[i].CreatedAt > shares[j].CreatedAt
		})
		OK(c, shares)
	}
}

// DeleteSessionShareHandler 删除会话分享

func (rt *Runtime) DeleteSessionShareHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		shareID := c.Param("shareID")
		if shareID == "" {
			Fail(c, http.StatusBadRequest, "missing share ID")
			return
		}

		store := rt.SessionShares
		exists, err := store.Delete(shareID)
		if !exists {
			Fail(c, http.StatusNotFound, "share not found")
			return
		}
		if err != nil {
			log.Printf("failed to persist session share deletion: id=%s, err=%v", shareID, err)
			Fail(c, http.StatusInternalServerError, "failed to delete session share")
			return
		}
		OK(c, gin.H{"message": "share deleted"})
	}
}

// GetPublicSessionShareInfoHandler 返回公开分享基础信息

func (rt *Runtime) GetPublicSessionShareInfoHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		shareID := c.Param("shareID")
		entry, exists, expired := rt.SessionShares.entryByID(shareID)
		if expired {
			Fail(c, http.StatusGone, "share expired")
			return
		}
		if !exists {
			Fail(c, http.StatusNotFound, "share not found")
			return
		}
		OK(c, gin.H{
			"id":                 shareID,
			"title":              entry.Title,
			"has_password":       entry.PasswordHash != "",
			"message_count":      entry.MessageCount,
			"expires_at":         expiresAtUnix(entry.ExpiresAt),
			"created_at":         entry.CreatedAt.Unix(),
			"allow_tool_details": entry.AllowToolDetails,
		})
	}
}

// AccessPublicSessionShareHandler 访问公开分享内容

func (rt *Runtime) AccessPublicSessionShareHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		shareID := c.Param("shareID")
		store := rt.SessionShares
		entry, exists, expired := store.entryByID(shareID)
		if expired {
			Fail(c, http.StatusGone, "share expired")
			return
		}
		if !exists {
			Fail(c, http.StatusNotFound, "share not found")
			return
		}

		var req struct {
			Password string `json:"password"`
		}
		if c.Request.Body != nil {
			_ = c.ShouldBindJSON(&req)
		}
		if entry.PasswordHash != "" {
			if req.Password == "" {
				c.JSON(http.StatusUnauthorized, Response{
					Code:    1,
					Message: "password required",
					Data:    gin.H{"require_password": true},
				})
				return
			}
			attemptKey := "session-share:" + shareID + ":" + c.ClientIP()
			if allowed, retryAfter := publicShareAttempts.Allow(attemptKey, time.Now()); !allowed {
				rateLimitExceeded(c, retryAfter)
				return
			}
			if !verifyPassword(req.Password, entry.PasswordHash) {
				c.JSON(http.StatusUnauthorized, Response{
					Code:    1,
					Message: "invalid password",
					Data:    gin.H{"require_password": true},
				})
				return
			}
			publicShareAttempts.Reset(attemptKey)
		}

		transcript, err := sessionShareTranscript(rt.HistoryDir, entry.SessionID, entry.AllowToolDetails)
		if err != nil {
			log.Printf("failed to load public session share: share=%s, session=%s, err=%v", shareID, entry.SessionID, err)
			Fail(c, http.StatusGone, "shared session unavailable")
			return
		}

		if err := store.Touch(shareID, time.Now()); err != nil {
			log.Printf("failed to persist session share access time: id=%s, err=%v", shareID, err)
		}

		OK(c, gin.H{
			"id":                 shareID,
			"title":              entry.Title,
			"events":             rt.transcriptToChatEvents(entry.SessionID, transcript),
			"message_count":      entry.MessageCount,
			"expires_at":         expiresAtUnix(entry.ExpiresAt),
			"created_at":         entry.CreatedAt.Unix(),
			"allow_tool_details": entry.AllowToolDetails,
		})
	}
}
