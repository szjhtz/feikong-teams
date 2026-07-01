package standalone

import (
	"context"
	"io"
	"reflect"
	"testing"

	domainevent "fkteams/internal/domain/event"
	domainmessage "fkteams/internal/domain/message"
	runtimeport "fkteams/internal/ports/runtime"
)

func TestRunTextBuildsBareAgentAndReturnsText(t *testing.T) {
	runtime := &fakeRuntime{deltas: []string{"标题"}}
	service := NewService(Dependencies{
		AgentRuntime:  runtime,
		RunnerRuntime: runtime,
	})

	title, err := service.RunText(context.Background(), Request{
		Name:        "session_title",
		Description: "title generator",
		Instruction: "return a short title",
		Model:       fakeModel{},
		Input:       "帮我重构 runtime",
	})
	if err != nil {
		t.Fatalf("run text: %v", err)
	}
	if title != "标题" {
		t.Fatalf("title = %q, want 标题", title)
	}
	if runtime.agentConfig == nil {
		t.Fatal("agent was not built")
	}
	if runtime.agentConfig.Name != "session_title" || runtime.agentConfig.Instruction != "return a short title" {
		t.Fatalf("agent config = %#v", runtime.agentConfig)
	}
	if len(runtime.agentConfig.Middlewares) != 0 {
		t.Fatalf("bare agent middlewares = %d, want 0", len(runtime.agentConfig.Middlewares))
	}
	if len(runtime.agentConfig.ToolMiddlewares) != 0 {
		t.Fatalf("bare agent tool middlewares = %d, want 0", len(runtime.agentConfig.ToolMiddlewares))
	}
	if runtime.runnerConfig.EnableStreaming {
		t.Fatal("RunText enabled streaming")
	}
	if runtime.runner.input.Message.Content != "帮我重构 runtime" {
		t.Fatalf("runner input = %#v", runtime.runner.input)
	}
}

func TestStreamTextEmitsDeltasAndReturnsFinalText(t *testing.T) {
	runtime := &fakeRuntime{deltas: []string{"会话", "标题"}}
	service := NewService(Dependencies{
		AgentRuntime:  runtime,
		RunnerRuntime: runtime,
	})

	var deltas []string
	var events []domainevent.Event
	title, err := service.StreamText(context.Background(), Request{
		Name:  "session_title",
		Model: fakeModel{},
		Input: "hello",
		EventSink: func(event domainevent.Event) error {
			events = append(events, event)
			return nil
		},
	}, func(delta string) error {
		deltas = append(deltas, delta)
		return nil
	})
	if err != nil {
		t.Fatalf("stream text: %v", err)
	}
	if title != "会话标题" {
		t.Fatalf("title = %q, want 会话标题", title)
	}
	if !runtime.runnerConfig.EnableStreaming {
		t.Fatal("StreamText did not enable streaming")
	}
	if !reflect.DeepEqual(deltas, []string{"会话", "标题"}) {
		t.Fatalf("deltas = %#v", deltas)
	}
	if len(events) != 3 {
		t.Fatalf("events = %d, want 3", len(events))
	}
}

func TestRunTextRejectsMissingRequiredFields(t *testing.T) {
	runtime := &fakeRuntime{}
	service := NewService(Dependencies{AgentRuntime: runtime, RunnerRuntime: runtime})

	if _, err := service.RunText(context.Background(), Request{Name: "x", Input: "hello"}); err == nil {
		t.Fatal("expected missing model error")
	}
	if _, err := service.RunText(context.Background(), Request{Name: "x", Model: fakeModel{}}); err == nil {
		t.Fatal("expected missing input error")
	}
	if _, err := NewService(Dependencies{}).RunText(context.Background(), Request{Name: "x", Model: fakeModel{}, Input: "hello"}); err == nil {
		t.Fatal("expected missing runtime error")
	}
}

type fakeRuntime struct {
	deltas       []string
	agentConfig  *runtimeport.ChatAgentConfig
	runnerConfig runtimeport.RunnerConfig
	runner       *fakeStandaloneRunner
}

func (r *fakeRuntime) NewChatModelAgent(_ context.Context, cfg *runtimeport.ChatAgentConfig) (runtimeport.Agent, error) {
	copied := *cfg
	r.agentConfig = &copied
	return fakeAgent{name: cfg.Name, description: cfg.Description}, nil
}

func (r *fakeRuntime) NewLoopAgent(context.Context, *runtimeport.LoopAgentConfig) (runtimeport.Agent, error) {
	return nil, nil
}

func (r *fakeRuntime) NewDeepAgent(context.Context, *runtimeport.DeepAgentConfig) (runtimeport.Agent, error) {
	return nil, nil
}

func (r *fakeRuntime) NewRunner(_ context.Context, cfg runtimeport.RunnerConfig) (runtimeport.Runner, error) {
	r.runnerConfig = cfg
	r.runner = &fakeStandaloneRunner{deltas: append([]string(nil), r.deltas...)}
	return r.runner, nil
}

type fakeAgent struct {
	name        string
	description string
}

func (a fakeAgent) Name() string {
	return a.name
}

func (a fakeAgent) Description() string {
	return a.description
}

type fakeStandaloneRunner struct {
	deltas []string
	input  domainmessage.TurnInput
	opts   runtimeport.RunOptions
}

func (r *fakeStandaloneRunner) Run(_ context.Context, input domainmessage.TurnInput, opts runtimeport.RunOptions) (*runtimeport.RunResult, error) {
	r.input = input
	r.opts = opts
	for _, delta := range r.deltas {
		if err := opts.Sink(domainevent.Event{Type: domainevent.TypeAssistantText, Content: delta}); err != nil {
			return nil, err
		}
	}
	completed := domainevent.Event{Type: domainevent.TypeAssistantCompleted, Content: joinDeltas(r.deltas)}
	if err := opts.Sink(completed); err != nil {
		return nil, err
	}
	return &runtimeport.RunResult{LastEvent: completed}, nil
}

type fakeModel struct{}

func (fakeModel) Generate(context.Context, []domainmessage.Message) (domainmessage.Message, error) {
	return domainmessage.Message{}, nil
}

func (fakeModel) Stream(context.Context, []domainmessage.Message) (runtimeport.MessageStream, error) {
	return emptyStream{}, nil
}

func (fakeModel) WithTools([]runtimeport.ToolInfo) (runtimeport.ChatModel, error) {
	return fakeModel{}, nil
}

type emptyStream struct{}

func (emptyStream) Recv() (domainmessage.Message, error) {
	return domainmessage.Message{}, io.EOF
}

func (emptyStream) Close() {}

func joinDeltas(deltas []string) string {
	var result string
	for _, delta := range deltas {
		result += delta
	}
	return result
}
