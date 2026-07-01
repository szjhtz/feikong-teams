package registry

import (
	"context"
	runtimeport "fkteams/internal/ports/runtime"
	"sort"
	"testing"
)

func TestRegisterAndUseRuntime(t *testing.T) {
	registry := NewRegistry(DefaultRuntimeName)
	engine := testEngine{}
	if err := registry.Register("test-runtime", engine); err != nil {
		t.Fatalf("register runtime: %v", err)
	}
	if err := registry.Use("test-runtime"); err != nil {
		t.Fatalf("use runtime: %v", err)
	}
	if registry.DefaultName() != "test-runtime" {
		t.Fatalf("default runtime = %q, want test-runtime", registry.DefaultName())
	}
	got, err := registry.Runtime()
	if err != nil {
		t.Fatalf("default engine: %v", err)
	}
	if got != engine {
		t.Fatal("Engine did not return registered runtime")
	}
}

func TestEngineByNameRequiresExplicitRegistration(t *testing.T) {
	registry := NewRegistry(DefaultRuntimeName)
	if _, err := registry.RuntimeByName("missing-runtime"); err == nil {
		t.Fatal("expected missing runtime error")
	}
}

func TestUseUnknownRuntimeReturnsError(t *testing.T) {
	registry := NewRegistry(DefaultRuntimeName)
	if err := registry.Use("missing-runtime"); err == nil {
		t.Fatal("expected error for missing runtime")
	}
}

func TestRegisteredNamesAreSorted(t *testing.T) {
	registry := NewRegistry(DefaultRuntimeName)
	if err := registry.Register("z-runtime", testEngine{}); err != nil {
		t.Fatalf("register z runtime: %v", err)
	}
	if err := registry.Register("a-runtime", testEngine{}); err != nil {
		t.Fatalf("register a runtime: %v", err)
	}

	names := registry.RegisteredNames()
	if !sort.StringsAreSorted(names) {
		t.Fatalf("registered names are not sorted: %v", names)
	}
}

type testEngine struct{}

func (testEngine) NewChatModelAgent(context.Context, *runtimeport.ChatAgentConfig) (runtimeport.Agent, error) {
	return nil, nil
}

func (testEngine) NewLoopAgent(context.Context, *runtimeport.LoopAgentConfig) (runtimeport.Agent, error) {
	return nil, nil
}

func (testEngine) NewDeepAgent(context.Context, *runtimeport.DeepAgentConfig) (runtimeport.Agent, error) {
	return nil, nil
}

func (testEngine) NewRunner(context.Context, runtimeport.RunnerConfig) (runtimeport.Runner, error) {
	return nil, nil
}

func (testEngine) NewAgentTools(context.Context, []runtimeport.Agent, runtimeport.AgentToolConfig) ([]runtimeport.Tool, error) {
	return nil, nil
}
