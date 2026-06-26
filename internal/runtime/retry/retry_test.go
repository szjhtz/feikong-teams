package retry

import (
	"context"
	"errors"
	"net"
	"testing"

	"fkteams/internal/ports/runtime"
	"fkteams/internal/runtime/env"
)

func TestMaxIterationsReadsEnvironment(t *testing.T) {
	t.Setenv(env.MaxIterations, "7")
	if got := MaxIterations(); got != 7 {
		t.Fatalf("MaxIterations = %d, want 7", got)
	}

	t.Setenv(env.MaxIterations, "0")
	if got := MaxIterations(); got != 1<<31-1 {
		t.Fatalf("MaxIterations zero = %d, want unlimited sentinel", got)
	}

	t.Setenv(env.MaxIterations, "bad")
	if got := MaxIterations(); got != defaultMaxIterations {
		t.Fatalf("MaxIterations invalid = %d, want default", got)
	}
}

func TestIsRetryable(t *testing.T) {
	if IsRetryable(context.Background(), nil) {
		t.Fatal("nil error should not be retryable")
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if IsRetryable(cancelled, errors.New("status code: 503")) {
		t.Fatal("cancelled context should not retry")
	}
	if !IsRetryable(context.Background(), timeoutError{}) {
		t.Fatal("net.Error should be retryable")
	}
	for _, msg := range []string{"status code: 429", "connection reset by peer", "stream error: INTERNAL_ERROR", "EOF"} {
		if !IsRetryable(context.Background(), errors.New(msg)) {
			t.Fatalf("%q should be retryable", msg)
		}
	}
	if IsRetryable(context.Background(), errors.New("bad request")) {
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
	decision := cfg.ShouldRetry(context.Background(), &runtime.RetryContext{Err: errors.New("status code: 500")})
	if decision == nil || !decision.Retry || decision.RejectReason == "" {
		t.Fatalf("retry decision = %#v", decision)
	}
	if cfg.ShouldRetry(context.Background(), &runtime.RetryContext{Err: errors.New("bad request")}) != nil {
		t.Fatal("non-retryable error should return nil decision")
	}
}

type timeoutError struct{}

func (timeoutError) Error() string   { return "timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

var _ net.Error = timeoutError{}
