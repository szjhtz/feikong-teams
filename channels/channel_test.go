package channels

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeChannel struct {
	name      string
	running   bool
	startErr  error
	sent      []sentMessage
	handler   MessageHandler
	started   int
	stopped   int
	stopCalls int
}

type sentMessage struct {
	chatID string
	msg    Message
}

func (c *fakeChannel) Name() string { return c.name }

func (c *fakeChannel) Start(context.Context) error {
	c.started++
	if c.startErr != nil {
		return c.startErr
	}
	c.running = true
	return nil
}

func (c *fakeChannel) Stop(context.Context) error {
	c.stopped++
	c.stopCalls++
	c.running = false
	return nil
}

func (c *fakeChannel) Send(_ context.Context, chatID string, msg Message) error {
	c.sent = append(c.sent, sentMessage{chatID: chatID, msg: msg})
	return nil
}

func (c *fakeChannel) IsRunning() bool { return c.running }

func TestAttachmentTypeName(t *testing.T) {
	tests := []struct {
		typ  MessageType
		want string
	}{
		{MsgImage, "图片"},
		{MsgAudio, "语音"},
		{MsgVideo, "视频"},
		{MsgFile, "文件"},
		{MsgText, "附件"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := (Attachment{Type: tt.typ}).TypeName(); got != tt.want {
				t.Fatalf("TypeName = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestManagerRegisterStartSendStop(t *testing.T) {
	factoryName := "fake_manager_test"
	var created *fakeChannel
	var handlerSeen MessageHandler
	RegisterFactory(factoryName, func(cfg ChannelConfig, handler MessageHandler) (Channel, error) {
		handlerSeen = handler
		created = &fakeChannel{name: factoryName, handler: handler}
		return created, nil
	})

	var handled bool
	manager := NewManager(func(context.Context, string, string, Message, bool) {
		handled = true
	})
	if err := manager.Register(factoryName, ChannelConfig{Enabled: false}); err != nil {
		t.Fatalf("Register disabled returned error: %v", err)
	}
	if _, ok := manager.Get(factoryName); ok {
		t.Fatal("disabled channel should not be registered")
	}

	if err := manager.Register(factoryName, ChannelConfig{Enabled: true}); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if handlerSeen == nil {
		t.Fatal("factory did not receive manager handler")
	}
	handlerSeen(context.Background(), "chat", "sender", Message{Content: "hi"}, false)
	if !handled {
		t.Fatal("manager handler was not invoked")
	}

	if err := manager.StartAll(context.Background()); err != nil {
		t.Fatalf("StartAll returned error: %v", err)
	}
	if !created.IsRunning() || created.started != 1 {
		t.Fatalf("channel running=%v started=%d, want running once", created.IsRunning(), created.started)
	}

	if err := manager.SendText(context.Background(), factoryName, "chat-1", "hello"); err != nil {
		t.Fatalf("SendText returned error: %v", err)
	}
	if len(created.sent) != 1 || created.sent[0].chatID != "chat-1" || created.sent[0].msg.Content != "hello" {
		t.Fatalf("sent messages = %#v", created.sent)
	}

	manager.StopAll(context.Background())
	if created.IsRunning() || created.stopped != 1 {
		t.Fatalf("channel running=%v stopped=%d, want stopped once", created.IsRunning(), created.stopped)
	}
}

func TestManagerErrors(t *testing.T) {
	manager := NewManager(nil)
	if err := manager.Register("missing_factory", ChannelConfig{Enabled: true}); err == nil || !strings.Contains(err.Error(), "unknown channel") {
		t.Fatalf("Register missing factory error = %v, want unknown channel", err)
	}
	if err := manager.SendText(context.Background(), "missing", "chat", "hello"); err == nil || !strings.Contains(err.Error(), "channel not found") {
		t.Fatalf("SendText missing channel error = %v, want channel not found", err)
	}

	factoryName := "fake_start_error_test"
	startErr := errors.New("boom")
	RegisterFactory(factoryName, func(ChannelConfig, MessageHandler) (Channel, error) {
		return &fakeChannel{name: factoryName, startErr: startErr}, nil
	})
	if err := manager.Register(factoryName, ChannelConfig{Enabled: true}); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if err := manager.StartAll(context.Background()); err == nil || !strings.Contains(err.Error(), "start channel") {
		t.Fatalf("StartAll error = %v, want start channel context", err)
	}
}
