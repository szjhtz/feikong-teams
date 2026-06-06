package agentcore

import (
	"context"
	"fmt"
)

type InterruptRuntime interface {
	Interrupt(ctx context.Context, info any) error
	GetInterruptState(ctx context.Context) (bool, bool, any)
	GetResumeContext(ctx context.Context, out any) (bool, bool)
}

var interruptRuntime InterruptRuntime

func RegisterInterruptRuntime(runtime InterruptRuntime) {
	interruptRuntime = runtime
}

func RequestInterrupt(ctx context.Context, info any) error {
	if interruptRuntime == nil {
		return fmt.Errorf("interrupt runtime is not registered")
	}
	return interruptRuntime.Interrupt(ctx, info)
}

func GetInterruptState(ctx context.Context) (bool, bool, any) {
	if interruptRuntime == nil {
		return false, false, nil
	}
	return interruptRuntime.GetInterruptState(ctx)
}

func GetResumeContext[T any](ctx context.Context) (bool, bool, T) {
	var value T
	if interruptRuntime == nil {
		return false, false, value
	}
	isTarget, hasData := interruptRuntime.GetResumeContext(ctx, &value)
	return isTarget, hasData, value
}
