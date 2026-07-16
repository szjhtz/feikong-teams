package qq

import (
	"context"
	"fmt"
	"time"

	"github.com/tencent-connect/botgo/dto"
	sessionmanager "github.com/tencent-connect/botgo/sessions/manager"
	"github.com/tencent-connect/botgo/websocket"
	"golang.org/x/oauth2"
)

const (
	qqReconnectMinDelay = time.Second
	qqReconnectMaxDelay = 30 * time.Second
	maxQQShards         = 64
)

type sessionRunner struct {
	newClient func(dto.Session) websocket.WebSocket
	minDelay  time.Duration
	maxDelay  time.Duration
}

func newSessionRunner() *sessionRunner {
	return &sessionRunner{
		newClient: func(session dto.Session) websocket.WebSocket {
			return websocket.ClientImpl.New(session)
		},
		minDelay: qqReconnectMinDelay,
		maxDelay: qqReconnectMaxDelay,
	}
}

func (r *sessionRunner) Run(ctx context.Context, info *dto.WebsocketAP, tokenSource oauth2.TokenSource, intent *dto.Intent) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if r == nil || r.newClient == nil {
		return fmt.Errorf("QQ websocket client factory is required")
	}
	if info == nil {
		return fmt.Errorf("QQ websocket session info is required")
	}
	if tokenSource == nil {
		return fmt.Errorf("QQ token source is required")
	}
	if intent == nil {
		return fmt.Errorf("QQ websocket intent is required")
	}
	if info.Shards == 0 {
		return fmt.Errorf("QQ websocket shard count must be positive")
	}
	if info.Shards > maxQQShards {
		return fmt.Errorf("QQ websocket shard count exceeds limit of %d", maxQQShards)
	}
	if err := sessionmanager.CheckSessionLimit(info); err != nil {
		return fmt.Errorf("check QQ websocket session limit: %w", err)
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	results := make(chan error, info.Shards)
	startInterval := sessionmanager.CalcInterval(info.SessionStartLimit.MaxConcurrency)
	for shardID := uint32(0); shardID < info.Shards; shardID++ {
		session := dto.Session{
			URL:         info.URL,
			TokenSource: tokenSource,
			Intent:      *intent,
			Shards: dto.ShardConfig{
				ShardID:    shardID,
				ShardCount: info.Shards,
			},
		}
		delay := time.Duration(shardID) * startInterval
		go func() {
			if !waitForSession(runCtx, delay) {
				results <- nil
				return
			}
			results <- r.runShard(runCtx, session)
		}()
	}

	var firstErr error
	for completed := uint32(0); completed < info.Shards; completed++ {
		if err := <-results; err != nil && firstErr == nil {
			firstErr = err
			cancel()
		}
	}
	return firstErr
}

func (r *sessionRunner) runShard(ctx context.Context, session dto.Session) error {
	delay := r.minDelay
	for {
		if ctx.Err() != nil {
			return nil
		}
		client := r.newClient(session)
		if client == nil {
			return fmt.Errorf("create QQ websocket client")
		}
		if err := client.Connect(); err != nil {
			// SDK 的未连接客户端无法安全关闭，避免重试持续遗留内部计时器。
			return fmt.Errorf("connect QQ websocket session: %w", err)
		}

		var authErr error
		if session.ID == "" {
			authErr = client.Identify()
		} else {
			authErr = client.Resume()
		}
		if authErr != nil {
			client.Close()
			if sessionmanager.CanNotIdentify(authErr) {
				return fmt.Errorf("authenticate QQ websocket session: %w", authErr)
			}
			if sessionmanager.CanNotResume(authErr) {
				session.ID = ""
				session.LastSeq = 0
			}
			if !waitForSession(ctx, delay) {
				return nil
			}
			delay = nextSessionDelay(delay, r.maxDelay)
			continue
		}

		delay = r.minDelay
		listenDone := make(chan error, 1)
		go func() {
			listenDone <- client.Listening()
		}()

		var listenErr error
		select {
		case <-ctx.Done():
			client.Close()
			<-listenDone
			return nil
		case listenErr = <-listenDone:
		}
		if current := client.Session(); current != nil {
			session = *current
		}
		if listenErr != nil && sessionmanager.CanNotIdentify(listenErr) {
			return fmt.Errorf("listen to QQ websocket session: %w", listenErr)
		}
		if listenErr != nil && sessionmanager.CanNotResume(listenErr) {
			session.ID = ""
			session.LastSeq = 0
		}
		if !waitForSession(ctx, delay) {
			return nil
		}
		delay = nextSessionDelay(delay, r.maxDelay)
	}
}

func waitForSession(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		return ctx.Err() == nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func nextSessionDelay(current, maximum time.Duration) time.Duration {
	if current <= 0 {
		return maximum
	}
	if current >= maximum/2 {
		return maximum
	}
	return current * 2
}
