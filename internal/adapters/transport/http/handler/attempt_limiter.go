package handler

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type attemptWindow struct {
	count     int
	startedAt time.Time
	lastSeen  time.Time
}

type attemptLimiter struct {
	mu         sync.Mutex
	entries    map[string]attemptWindow
	limit      int
	window     time.Duration
	maxEntries int
}

func newAttemptLimiter(limit int, window time.Duration, maxEntries int) *attemptLimiter {
	return &attemptLimiter{
		entries:    make(map[string]attemptWindow),
		limit:      limit,
		window:     window,
		maxEntries: maxEntries,
	}
}

func (l *attemptLimiter) Allow(key string, now time.Time) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, exists := l.entries[key]
	if !exists || now.Sub(entry.startedAt) >= l.window {
		if !exists && len(l.entries) >= l.maxEntries {
			l.evictOldestLocked(now)
		}
		l.entries[key] = attemptWindow{count: 1, startedAt: now, lastSeen: now}
		return true, 0
	}
	entry.lastSeen = now
	if entry.count >= l.limit {
		l.entries[key] = entry
		return false, entry.startedAt.Add(l.window).Sub(now)
	}
	entry.count++
	l.entries[key] = entry
	return true, 0
}

func (l *attemptLimiter) Reset(key string) {
	l.mu.Lock()
	delete(l.entries, key)
	l.mu.Unlock()
}

func (l *attemptLimiter) evictOldestLocked(now time.Time) {
	var oldestKey string
	var oldestTime time.Time
	for key, entry := range l.entries {
		if now.Sub(entry.lastSeen) >= l.window {
			delete(l.entries, key)
			continue
		}
		if oldestKey == "" || entry.lastSeen.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.lastSeen
		}
	}
	if len(l.entries) >= l.maxEntries && oldestKey != "" {
		delete(l.entries, oldestKey)
	}
}

func rateLimitExceeded(c *gin.Context, retryAfter time.Duration) {
	seconds := int(retryAfter.Round(time.Second) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	c.Header("Retry-After", fmt.Sprintf("%d", seconds))
	Fail(c, http.StatusTooManyRequests, "too many authentication attempts")
}

var (
	loginAttempts       = newAttemptLimiter(8, 5*time.Minute, 10000)
	publicShareAttempts = newAttemptLimiter(8, 5*time.Minute, 20000)
)
