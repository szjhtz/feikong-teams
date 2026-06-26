package registry

import (
	"context"
	runtimeport "fkteams/internal/ports/runtime"
	"sort"
	"testing"
)

func TestRegisterAndUseRuntime(t *testing.T) {
	original := DefaultName()
	t.Cleanup(func() {
		registry.Lock()
		registry.defaultName = original
		registry.Unlock()
	})

	engine := testEngine{}
	if err := Register("test-runtime", engine); err != nil {
		t.Fatalf("register runtime: %v", err)
	}
	if err := Use("test-runtime"); err != nil {
		t.Fatalf("use runtime: %v", err)
	}
	if DefaultName() != "test-runtime" {
		t.Fatalf("default runtime = %q, want test-runtime", DefaultName())
	}
	got, err := Engine()
	if err != nil {
		t.Fatalf("default engine: %v", err)
	}
	if got != engine {
		t.Fatal("Engine did not return registered runtime")
	}
}

func TestEngineByNameRequiresExplicitRegistration(t *testing.T) {
	if _, err := EngineByName("missing-runtime"); err == nil {
		t.Fatal("expected missing runtime error")
	}
}

func TestUseUnknownRuntimeReturnsError(t *testing.T) {
	if err := Use("missing-runtime"); err == nil {
		t.Fatal("expected error for missing runtime")
	}
}

func TestRegisteredNamesAreSorted(t *testing.T) {
	if err := Register("z-runtime", testEngine{}); err != nil {
		t.Fatalf("register z runtime: %v", err)
	}
	if err := Register("a-runtime", testEngine{}); err != nil {
		t.Fatalf("register a runtime: %v", err)
	}

	names := RegisteredNames()
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
