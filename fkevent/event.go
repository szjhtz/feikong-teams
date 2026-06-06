// Package fkevent provides the engine-neutral event dispatch layer.
package fkevent

import (
	"context"
	"fkteams/agentcore"
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

type callbackKey struct{}
type nonInteractiveKey struct{}

var globalEventSequence int64

// WithCallback binds an event callback to context.
func WithCallback(ctx context.Context, cb func(Event) error) context.Context {
	return context.WithValue(ctx, callbackKey{}, cb)
}

func getCallback(ctx context.Context) func(Event) error {
	if cb, ok := ctx.Value(callbackKey{}).(func(Event) error); ok {
		return cb
	}
	return nil
}

// WithNonInteractive marks a context as non-interactive.
func WithNonInteractive(ctx context.Context) context.Context {
	return context.WithValue(ctx, nonInteractiveKey{}, true)
}

// IsNonInteractive reports whether a context is marked non-interactive.
func IsNonInteractive(ctx context.Context) bool {
	v, _ := ctx.Value(nonInteractiveKey{}).(bool)
	return v
}

// NormalizeEvent fills common metadata for an event.
func NormalizeEvent(event Event) Event {
	if event.Sequence == 0 {
		event.Sequence = atomic.AddInt64(&globalEventSequence, 1)
	}
	if event.EventID == "" {
		event.EventID = fmt.Sprintf("evt_%d", event.Sequence)
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	if event.Delta == "" {
		event.Delta = event.Content
	}
	if event.RunID == "" {
		event.RunID = event.SpanID
	}
	if event.MemberCallID != "" {
		event.IsMemberEvent = true
	}
	return event
}

// DispatchEvent normalizes and sends an event to the context callback.
func DispatchEvent(ctx context.Context, event Event) error {
	event = NormalizeEvent(event)
	if cb := getCallback(ctx); cb != nil {
		return cb(event)
	}
	return nil
}

// Dispatch is a convenience EventSink adapter.
func Dispatch(ctx context.Context) agentcore.EventSink {
	return func(event agentcore.Event) error {
		return DispatchEvent(ctx, event)
	}
}

func IsInternalToolName(name string) bool {
	return name == "continue_output"
}

func IsInternalContinueContent(content string) bool {
	return strings.Contains(content, "Your previous text output was truncated") ||
		strings.Contains(content, "Your previous tool call was truncated")
}
