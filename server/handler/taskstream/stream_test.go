package taskstream

import (
	"testing"
	"time"
)

func newTestStream() *Stream {
	return NewManager().Register(StreamConfig{
		SessionID: "test-session",
		Cancel:    func() {},
	})
}

func TestSubmitInterruptRequiresPendingKind(t *testing.T) {
	s := newTestStream()

	if err := s.SubmitInterrupt(InterruptApproval, 1); err == nil {
		t.Fatal("expected submit without pending request to fail")
	}

	s.BeginInterrupt(InterruptApproval)
	if err := s.SubmitInterrupt(InterruptAsk, "answer"); err == nil {
		t.Fatal("expected submit with wrong interrupt kind to fail")
	}

	if err := s.SubmitInterrupt(InterruptApproval, 1); err != nil {
		t.Fatalf("expected approval submit to succeed: %v", err)
	}
	if err := s.SubmitInterrupt(InterruptApproval, 2); err == nil {
		t.Fatal("expected duplicate submit to fail")
	}

	got := <-s.InterruptCh()
	if got != 1 {
		t.Fatalf("expected first decision to be delivered, got %v", got)
	}

	s.CompleteInterrupt(InterruptApproval)
	if err := s.SubmitInterrupt(InterruptApproval, 1); err == nil {
		t.Fatal("expected submit after completion to fail")
	}
}

func TestBeginInterruptDrainsStaleDecision(t *testing.T) {
	s := newTestStream()
	s.interruptCh <- "stale"

	s.BeginInterrupt(InterruptAsk)
	if err := s.SubmitInterrupt(InterruptAsk, "fresh"); err != nil {
		t.Fatalf("expected ask submit to succeed: %v", err)
	}

	got := <-s.InterruptCh()
	if got != "fresh" {
		t.Fatalf("expected stale decision to be drained, got %v", got)
	}
}

func TestUnsubscribeWithZeroGraceDoesNotCancelTask(t *testing.T) {
	cancelled := make(chan struct{}, 1)
	s := NewManager().Register(StreamConfig{
		SessionID:   "test-session",
		Cancel:      func() { cancelled <- struct{}{} },
		GracePeriod: 0,
	})

	ok, epoch := s.Subscribe(FuncSubscriber(func(any) error { return nil }))
	if !ok {
		t.Fatal("expected subscribe to succeed")
	}

	s.Unsubscribe(epoch)

	select {
	case <-cancelled:
		t.Fatal("expected unsubscribe to detach without cancelling task")
	case <-time.After(20 * time.Millisecond):
	}
}
