package memory

import "context"

// LLMClient 是长期记忆提取需要的最小模型能力。
type LLMClient interface {
	Complete(ctx context.Context, prompt string) (string, error)
}
