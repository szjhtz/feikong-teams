package agentcore

import (
	"context"
	"errors"
	"testing"
)

func TestInterruptRuntimeDelegation(t *testing.T) {
	oldRuntime := interruptRuntime
	t.Cleanup(func() {
		interruptRuntime = oldRuntime
	})

	RegisterInterruptRuntime(nil)
	if err := RequestInterrupt(context.Background(), "info"); err == nil {
		t.Fatal("expected missing runtime error")
	}
	interrupted, resumable, state := GetInterruptState(context.Background())
	if interrupted || resumable || state != nil {
		t.Fatalf("empty interrupt state = %v %v %#v", interrupted, resumable, state)
	}
	target, hasData, value := GetResumeContext[string](context.Background())
	if target || hasData || value != "" {
		t.Fatalf("empty resume context = %v %v %q", target, hasData, value)
	}

	runtime := &fakeInterruptRuntime{
		interruptErr: errors.New("stop"),
		state:        "state",
		resumeValue:  "resume",
	}
	RegisterInterruptRuntime(runtime)
	if err := RequestInterrupt(context.Background(), "payload"); !errors.Is(err, runtime.interruptErr) {
		t.Fatalf("interrupt error = %v", err)
	}
	if runtime.info != "payload" {
		t.Fatalf("runtime info = %#v", runtime.info)
	}
	interrupted, resumable, state = GetInterruptState(context.Background())
	if !interrupted || !resumable || state != "state" {
		t.Fatalf("runtime state = %v %v %#v", interrupted, resumable, state)
	}
	target, hasData, value = GetResumeContext[string](context.Background())
	if !target || !hasData || value != "resume" {
		t.Fatalf("resume context = %v %v %q", target, hasData, value)
	}
}

type fakeInterruptRuntime struct {
	info         any
	interruptErr error
	state        any
	resumeValue  string
}

func (r *fakeInterruptRuntime) Interrupt(_ context.Context, info any) error {
	r.info = info
	return r.interruptErr
}

func (r *fakeInterruptRuntime) GetInterruptState(context.Context) (bool, bool, any) {
	return true, true, r.state
}

func (r *fakeInterruptRuntime) GetResumeContext(_ context.Context, out any) (bool, bool) {
	if ptr, ok := out.(*string); ok {
		*ptr = r.resumeValue
	}
	return true, true
}
