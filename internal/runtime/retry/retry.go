package retry

import (
	"context"
	"errors"
	"net"
	"strconv"
	"strings"

	"fkteams/internal/ports/runtime"
	"fkteams/internal/runtime/env"
)

const (
	// MaxRetries 是模型调用最大重试次数。
	MaxRetries = 3

	defaultMaxIterations = 60
)

// MaxIterations 返回智能体最大迭代次数，支持 FEIKONG_MAX_ITERATIONS 环境变量覆盖。
func MaxIterations() int {
	if v := env.Get(env.MaxIterations); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			if n <= 0 {
				return 1<<31 - 1
			}
			return n
		}
	}
	return defaultMaxIterations
}

// IsRetryable 判断错误是否可重试。
func IsRetryable(ctx context.Context, err error) bool {
	if err == nil {
		return false
	}
	if ctx.Err() != nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	msg := err.Error()
	return strings.Contains(msg, "status code: 429") ||
		strings.Contains(msg, "status code: 500") ||
		strings.Contains(msg, "status code: 502") ||
		strings.Contains(msg, "status code: 503") ||
		strings.Contains(msg, "status code: 504") ||
		strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "stream error") ||
		strings.Contains(msg, "INTERNAL_ERROR") ||
		strings.Contains(msg, "EOF")
}

// NewModelRetryConfig 返回核心模型重试配置。
func NewModelRetryConfig() *runtime.ModelRetryConfig {
	return &runtime.ModelRetryConfig{
		MaxRetries: MaxRetries,
		ShouldRetry: func(ctx context.Context, retryCtx *runtime.RetryContext) *runtime.RetryDecision {
			if retryCtx == nil || retryCtx.Err == nil || !IsRetryable(ctx, retryCtx.Err) {
				return nil
			}
			return &runtime.RetryDecision{
				Retry:        true,
				RejectReason: retryCtx.Err.Error(),
			}
		},
	}
}
