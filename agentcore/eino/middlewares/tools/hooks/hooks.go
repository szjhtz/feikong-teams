// Package hooks 将工具执行边界暴露给项目 HookBus。
package hooks

import (
	"context"
	"fmt"

	"fkteams/agentcore"
	einoruntime "fkteams/agentcore/eino"
	projecthooks "fkteams/hooks"

	"github.com/cloudwego/eino/compose"
)

// New 创建工具 hook 中间件。
func New() agentcore.ToolMiddleware {
	return einoruntime.WrapToolMiddleware("hooks", compose.ToolMiddleware{
		Invokable: func(next compose.InvokableToolEndpoint) compose.InvokableToolEndpoint {
			return func(ctx context.Context, input *compose.ToolInput) (*compose.ToolOutput, error) {
				if err := invokeBeforeTool(ctx, input); err != nil {
					return nil, err
				}
				output, err := next(ctx, input)
				result := ""
				if output != nil {
					result = output.Result
				}
				if hookErr := invokeAfterTool(ctx, input, result, err); hookErr != nil && err == nil {
					err = hookErr
				}
				return output, err
			}
		},
		Streamable: func(next compose.StreamableToolEndpoint) compose.StreamableToolEndpoint {
			return func(ctx context.Context, input *compose.ToolInput) (*compose.StreamToolOutput, error) {
				if err := invokeBeforeTool(ctx, input); err != nil {
					return nil, err
				}
				output, err := next(ctx, input)
				if hookErr := invokeAfterTool(ctx, input, "<stream>", err); hookErr != nil && err == nil {
					err = hookErr
				}
				return output, err
			}
		},
	})
}

func invokeBeforeTool(ctx context.Context, input *compose.ToolInput) error {
	if input == nil {
		return nil
	}
	result, err := projecthooks.FromContext(ctx).Invoke(ctx, projecthooks.Invocation{
		HookPoint: projecthooks.HookBeforeToolCall,
		Payload: projecthooks.BeforeToolCallPayload{
			ToolName: input.Name,
			Args:     input.Arguments,
			Meta: map[string]any{
				"call_id": input.CallID,
			},
		},
	})
	if err != nil {
		return err
	}
	if result.Action == projecthooks.ActionReject {
		if result.Message != "" {
			return fmt.Errorf("tool call rejected by hook: %s", result.Message)
		}
		return fmt.Errorf("tool call rejected by hook")
	}
	if result.Action == projecthooks.ActionSkip {
		return fmt.Errorf("tool call skipped by hook")
	}
	if payload, ok := result.Payload.(projecthooks.BeforeToolCallPayload); ok {
		input.Arguments = payload.Args
	}
	return nil
}

func invokeAfterTool(ctx context.Context, input *compose.ToolInput, output string, toolErr error) error {
	if input == nil {
		return nil
	}
	_, err := projecthooks.FromContext(ctx).Invoke(ctx, projecthooks.Invocation{
		HookPoint: projecthooks.HookAfterToolCall,
		Payload: projecthooks.AfterToolCallPayload{
			ToolName: input.Name,
			Args:     input.Arguments,
			Result:   output,
			Error:    toolErr,
			Meta: map[string]any{
				"call_id": input.CallID,
			},
		},
	})
	return err
}
