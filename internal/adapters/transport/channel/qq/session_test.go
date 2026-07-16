package qq

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/websocket"
	"golang.org/x/oauth2"
)

func TestSessionRunnerStopsActiveClient(t *testing.T) {
	client := newFakeSessionClient()
	runner := &sessionRunner{
		newClient: func(session dto.Session) websocket.WebSocket {
			client.session = session
			return client
		},
		minDelay: time.Millisecond,
		maxDelay: 2 * time.Millisecond,
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- runner.Run(ctx, testSessionInfo(1), testTokenSource(), testIntent())
	}()

	select {
	case <-client.listening:
	case <-time.After(time.Second):
		t.Fatal("session client did not start listening")
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("session runner returned error while stopping: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("session runner did not stop")
	}
	if !client.closed.Load() {
		t.Fatal("active websocket client was not closed")
	}
}

func TestSessionRunnerValidatesShardCount(t *testing.T) {
	runner := newSessionRunner()
	info := testSessionInfo(maxQQShards + 1)
	info.SessionStartLimit.Remaining = info.Shards
	if err := runner.Run(context.Background(), info, testTokenSource(), testIntent()); err == nil {
		t.Fatal("expected excessive shard count to be rejected")
	}
}

type fakeSessionClient struct {
	session       dto.Session
	listening     chan struct{}
	listeningOnce sync.Once
	closeOnce     sync.Once
	closeCh       chan struct{}
	closed        atomic.Bool
}

func newFakeSessionClient() *fakeSessionClient {
	return &fakeSessionClient{
		listening: make(chan struct{}),
		closeCh:   make(chan struct{}),
	}
}

func (c *fakeSessionClient) New(session dto.Session) websocket.WebSocket {
	c.session = session
	return c
}

func (c *fakeSessionClient) Connect() error  { return nil }
func (c *fakeSessionClient) Identify() error { return nil }
func (c *fakeSessionClient) Resume() error   { return nil }
func (c *fakeSessionClient) Session() *dto.Session {
	return &c.session
}
func (c *fakeSessionClient) Listening() error {
	c.listeningOnce.Do(func() { close(c.listening) })
	<-c.closeCh
	return nil
}
func (c *fakeSessionClient) Write(*dto.WSPayload) error { return nil }
func (c *fakeSessionClient) Close() {
	c.closed.Store(true)
	c.closeOnce.Do(func() { close(c.closeCh) })
}

func testSessionInfo(shards uint32) *dto.WebsocketAP {
	return &dto.WebsocketAP{
		URL:    "wss://example.invalid",
		Shards: shards,
		SessionStartLimit: dto.SessionStartLimit{
			Remaining:      shards,
			MaxConcurrency: 1,
		},
	}
}

func testTokenSource() oauth2.TokenSource {
	return oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test", TokenType: "QQBot"})
}

func testIntent() *dto.Intent {
	intent := dto.IntentGroupMessages
	return &intent
}
