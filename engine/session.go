package engine

import (
	"context"
	"fkteams/fkevent"

	"github.com/cloudwego/eino/adk"
)

type EventHandler func(fkevent.Event) error
type StartHandler func(context.Context)
type FinishHandler func(context.Context, *adk.AgentEvent, error)

// Session 提供面向一次会话执行的易用接口。
type Session struct {
	engine *core
	cfg    runConfig
}

func NewSession(runner *adk.Runner, checkpointID string) *Session {
	return &Session{engine: newEngine(runner, checkpointID)}
}

func (s *Session) WithInput(input TurnInput) *Session {
	s.cfg.Messages = input.Messages
	s.cfg.UserInput = input.UserInput
	return s
}

func (s *Session) WithMessages(messages []adk.Message) *Session {
	s.cfg.Messages = messages
	return s
}

func (s *Session) OnEvent(handler EventHandler) *Session {
	s.cfg.EventCallback = handler
	return s
}

func (s *Session) WithHistory(history HistorySink) *Session {
	s.cfg.Recorder = history
	return s
}

func (s *Session) OnStart(handler StartHandler) *Session {
	s.cfg.OnStart = handler
	return s
}

func (s *Session) OnInterrupt(handler InterruptHandler) *Session {
	s.cfg.OnInterrupt = handler
	return s
}

func (s *Session) NonInteractive() *Session {
	s.cfg.NonInteractive = true
	return s
}

func (s *Session) WithContext(hook ContextHook) *Session {
	if hook != nil {
		s.cfg.ContextHooks = append(s.cfg.ContextHooks, hook)
	}
	return s
}

func (s *Session) OnFinish(handler FinishHandler) *Session {
	s.cfg.OnFinish = handler
	return s
}

func (s *Session) Run(ctx context.Context) (*adk.AgentEvent, error) {
	return s.engine.run(ctx, s.cfg)
}
