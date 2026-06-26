package runtime

import (
	"context"
	"encoding/gob"
	"fmt"
	"sync"
)

type interruptMetadataContextKey struct{}
type interruptRuntimeContextKey struct{}

// InterruptRuntime 封装底层 runtime 的 HITL 中断能力。
type InterruptRuntime interface {
	Interrupt(ctx context.Context, info any) error
	GetInterruptState(ctx context.Context) (bool, bool, any)
	GetResumeContext(ctx context.Context, out any) (bool, bool)
}

// InterruptMetadata 描述成员智能体触发中断时的上游身份。
type InterruptMetadata struct {
	MemberCallID   string
	MemberToolName string
	MemberName     string
	MemberOrder    *int
}

// InterruptPayload 是跨 runtime 传递中断信息和成员身份的稳定载荷。
type InterruptPayload struct {
	Info     any
	Metadata InterruptMetadata
}

var interruptRuntimeRegistry = struct {
	sync.RWMutex
	runtime InterruptRuntime
}{}

func init() {
	gob.Register(InterruptPayload{})
}

// RegisterInterruptRuntime 注册进程默认中断 runtime，并返回恢复函数。
func RegisterInterruptRuntime(runtime InterruptRuntime) func() {
	interruptRuntimeRegistry.Lock()
	previous := interruptRuntimeRegistry.runtime
	interruptRuntimeRegistry.runtime = runtime
	interruptRuntimeRegistry.Unlock()
	return func() {
		interruptRuntimeRegistry.Lock()
		interruptRuntimeRegistry.runtime = previous
		interruptRuntimeRegistry.Unlock()
	}
}

// WithInterruptRuntime 为当前 context 覆盖默认中断 runtime。
func WithInterruptRuntime(ctx context.Context, runtime InterruptRuntime) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if runtime == nil {
		return ctx
	}
	return context.WithValue(ctx, interruptRuntimeContextKey{}, runtime)
}

// WithInterruptMetadata 将成员中断身份写入 context。
func WithInterruptMetadata(ctx context.Context, metadata InterruptMetadata) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, interruptMetadataContextKey{}, metadata)
}

// InterruptMetadataFromContext 从 context 读取成员中断身份。
func InterruptMetadataFromContext(ctx context.Context) (InterruptMetadata, bool) {
	if ctx == nil {
		return InterruptMetadata{}, false
	}
	metadata, ok := ctx.Value(interruptMetadataContextKey{}).(InterruptMetadata)
	return metadata, ok
}

// RequestInterrupt 触发一次 HITL 中断。
func RequestInterrupt(ctx context.Context, info any) error {
	runtime := interruptRuntimeFromContext(ctx)
	if runtime == nil {
		return fmt.Errorf("interrupt runtime is not registered")
	}
	return runtime.Interrupt(ctx, info)
}

// GetInterruptState 查询当前调用是否处于中断恢复流程。
func GetInterruptState(ctx context.Context) (bool, bool, any) {
	runtime := interruptRuntimeFromContext(ctx)
	if runtime == nil {
		return false, false, nil
	}
	return runtime.GetInterruptState(ctx)
}

// GetResumeContext 读取当前中断恢复数据。
func GetResumeContext[T any](ctx context.Context) (bool, bool, T) {
	var value T
	runtime := interruptRuntimeFromContext(ctx)
	if runtime == nil {
		return false, false, value
	}
	isTarget, hasData := runtime.GetResumeContext(ctx, &value)
	return isTarget, hasData, value
}

func interruptRuntimeFromContext(ctx context.Context) InterruptRuntime {
	if ctx != nil {
		if runtime, ok := ctx.Value(interruptRuntimeContextKey{}).(InterruptRuntime); ok && runtime != nil {
			return runtime
		}
	}
	interruptRuntimeRegistry.RLock()
	defer interruptRuntimeRegistry.RUnlock()
	return interruptRuntimeRegistry.runtime
}
