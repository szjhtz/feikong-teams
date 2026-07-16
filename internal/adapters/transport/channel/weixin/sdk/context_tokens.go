package wechatbot

import (
	"container/list"
	"sync"
	"time"
)

const (
	maxContextTokens = 4_096
	contextTokenTTL  = 24 * time.Hour
)

type contextTokenEntry struct {
	userID   string
	token    string
	lastSeen time.Time
}

type contextTokenCache struct {
	mu         sync.Mutex
	entries    map[string]*list.Element
	order      *list.List
	maxEntries int
	ttl        time.Duration
}

func newContextTokenCache(maxEntries int, ttl time.Duration) *contextTokenCache {
	return &contextTokenCache{
		entries:    make(map[string]*list.Element),
		order:      list.New(),
		maxEntries: maxEntries,
		ttl:        ttl,
	}
}

func (c *contextTokenCache) Set(userID, token string) {
	if c == nil || userID == "" || token == "" {
		return
	}
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pruneExpiredLocked(now)
	if element := c.entries[userID]; element != nil {
		entry := element.Value.(*contextTokenEntry)
		entry.token = token
		entry.lastSeen = now
		c.order.MoveToFront(element)
		return
	}
	for len(c.entries) >= c.maxEntries {
		c.removeOldestLocked()
	}
	entry := &contextTokenEntry{userID: userID, token: token, lastSeen: now}
	c.entries[userID] = c.order.PushFront(entry)
}

func (c *contextTokenCache) Get(userID string) (string, bool) {
	if c == nil || userID == "" {
		return "", false
	}
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	element := c.entries[userID]
	if element == nil {
		return "", false
	}
	entry := element.Value.(*contextTokenEntry)
	if now.Sub(entry.lastSeen) >= c.ttl {
		delete(c.entries, userID)
		c.order.Remove(element)
		return "", false
	}
	entry.lastSeen = now
	c.order.MoveToFront(element)
	return entry.token, true
}

func (c *contextTokenCache) Reset() {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.entries = make(map[string]*list.Element)
	c.order.Init()
	c.mu.Unlock()
}

func (c *contextTokenCache) pruneExpiredLocked(now time.Time) {
	for element := c.order.Back(); element != nil; element = c.order.Back() {
		entry := element.Value.(*contextTokenEntry)
		if now.Sub(entry.lastSeen) < c.ttl {
			return
		}
		c.removeOldestLocked()
	}
}

func (c *contextTokenCache) removeOldestLocked() {
	element := c.order.Back()
	if element == nil {
		return
	}
	entry := element.Value.(*contextTokenEntry)
	delete(c.entries, entry.userID)
	c.order.Remove(element)
}
