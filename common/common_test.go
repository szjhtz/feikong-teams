package common

import (
	"context"
	"errors"
	"fkteams/agentcore"
	"fkteams/fkenv"
	"net"
	"path/filepath"
	"strings"
	"testing"
)

func TestSessionIDContext(t *testing.T) {
	ctx := WithSessionID(context.Background(), "session-1")
	got, ok := SessionIDFromCtx(ctx)
	if !ok || got != "session-1" {
		t.Fatalf("SessionIDFromCtx = %q/%v, want session-1/true", got, ok)
	}
	if _, ok := SessionIDFromCtx(context.Background()); ok {
		t.Fatal("empty context should not have session id")
	}
}

func TestMaxIterationsReadsEnvironment(t *testing.T) {
	t.Setenv(fkenv.MaxIterations, "7")
	if got := MaxIterations(); got != 7 {
		t.Fatalf("MaxIterations = %d, want 7", got)
	}

	t.Setenv(fkenv.MaxIterations, "0")
	if got := MaxIterations(); got != 1<<31-1 {
		t.Fatalf("MaxIterations zero = %d, want unlimited sentinel", got)
	}

	t.Setenv(fkenv.MaxIterations, "bad")
	if got := MaxIterations(); got != defaultMaxIterations {
		t.Fatalf("MaxIterations invalid = %d, want default", got)
	}
}

func TestDirectoryHelpersUseAppDir(t *testing.T) {
	appDir := t.TempDir()
	t.Setenv(fkenv.AppDir, appDir)

	if AppDir() != appDir {
		t.Fatalf("AppDir = %q, want %q", AppDir(), appDir)
	}
	for _, got := range []string{SessionsDir(), WorkspaceDir(), SchedulerDir(), ShareDir()} {
		if !strings.HasPrefix(got, appDir+string(filepath.Separator)) {
			t.Fatalf("derived dir %q should be under app dir %q", got, appDir)
		}
	}
}

func TestGenerateSessionIDLooksLikeUUID(t *testing.T) {
	id := GenerateSessionID()
	if len(id) != 36 || strings.Count(id, "-") != 4 {
		t.Fatalf("session id = %q, want UUID-like string", id)
	}
}

func TestIsRetryAble(t *testing.T) {
	if IsRetryAble(context.Background(), nil) {
		t.Fatal("nil error should not be retryable")
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if IsRetryAble(cancelled, errors.New("status code: 503")) {
		t.Fatal("cancelled context should not retry")
	}
	if !IsRetryAble(context.Background(), timeoutError{}) {
		t.Fatal("net.Error should be retryable")
	}
	for _, msg := range []string{"status code: 429", "connection reset by peer", "stream error: INTERNAL_ERROR", "EOF"} {
		if !IsRetryAble(context.Background(), errors.New(msg)) {
			t.Fatalf("%q should be retryable", msg)
		}
	}
	if IsRetryAble(context.Background(), errors.New("bad request")) {
		t.Fatal("ordinary error should not be retryable")
	}
}

func TestNewModelRetryConfig(t *testing.T) {
	cfg := NewModelRetryConfig()
	if cfg.MaxRetries != MaxRetries {
		t.Fatalf("MaxRetries = %d, want %d", cfg.MaxRetries, MaxRetries)
	}
	if cfg.ShouldRetry(context.Background(), nil) != nil {
		t.Fatal("nil retry context should not retry")
	}
	decision := cfg.ShouldRetry(context.Background(), &agentcore.RetryContext{Err: errors.New("status code: 500")})
	if decision == nil || !decision.Retry || decision.RejectReason == "" {
		t.Fatalf("retry decision = %#v", decision)
	}
	if cfg.ShouldRetry(context.Background(), &agentcore.RetryContext{Err: errors.New("bad request")}) != nil {
		t.Fatal("non-retryable error should return nil decision")
	}
}

type timeoutError struct{}

func (timeoutError) Error() string   { return "timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

var _ net.Error = timeoutError{}
