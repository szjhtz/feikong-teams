package testmodel

import (
	"context"
	"fmt"
	"sync"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type GenerateResult struct {
	Message *schema.Message
	Err     error
}

type StreamResult struct {
	Chunks []*schema.Message
	Err    error
}

type Call struct {
	Input []*schema.Message
	Tools []*schema.ToolInfo
	Opts  []model.Option
}

type Model struct {
	state *state
	tools []*schema.ToolInfo
}

type state struct {
	mu             sync.Mutex
	generateQueue  []GenerateResult
	streamQueue    []StreamResult
	generateCalls  []Call
	streamCalls    []Call
	withToolsCalls [][]*schema.ToolInfo
}

var _ model.ToolCallingChatModel = (*Model)(nil)

func New(responses ...*schema.Message) *Model {
	m := &Model{state: &state{}}
	for _, resp := range responses {
		m.EnqueueGenerate(resp, nil)
	}
	return m
}

func (m *Model) EnqueueGenerate(message *schema.Message, err error) *Model {
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	m.state.generateQueue = append(m.state.generateQueue, GenerateResult{Message: message, Err: err})
	return m
}

func (m *Model) EnqueueStream(chunks ...*schema.Message) *Model {
	m.EnqueueStreamResult(chunks, nil)
	return m
}

func (m *Model) EnqueueStreamResult(chunks []*schema.Message, err error) *Model {
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	m.state.streamQueue = append(m.state.streamQueue, StreamResult{Chunks: chunks, Err: err})
	return m
}

func (m *Model) Generate(_ context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	m.state.mu.Lock()
	defer m.state.mu.Unlock()

	m.state.generateCalls = append(m.state.generateCalls, Call{
		Input: copyMessages(input),
		Tools: copyTools(m.tools),
		Opts:  copyOptions(opts),
	})
	if len(m.state.generateQueue) == 0 {
		return nil, fmt.Errorf("testmodel: no queued generate response")
	}

	resp := m.state.generateQueue[0]
	m.state.generateQueue = m.state.generateQueue[1:]
	return resp.Message, resp.Err
}

func (m *Model) Stream(_ context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	m.state.mu.Lock()
	defer m.state.mu.Unlock()

	m.state.streamCalls = append(m.state.streamCalls, Call{
		Input: copyMessages(input),
		Tools: copyTools(m.tools),
		Opts:  copyOptions(opts),
	})
	if len(m.state.streamQueue) == 0 {
		return nil, fmt.Errorf("testmodel: no queued stream response")
	}

	resp := m.state.streamQueue[0]
	m.state.streamQueue = m.state.streamQueue[1:]
	if resp.Err != nil {
		return nil, resp.Err
	}
	return schema.StreamReaderFromArray(resp.Chunks), nil
}

func (m *Model) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	copied := copyTools(tools)
	m.state.withToolsCalls = append(m.state.withToolsCalls, copied)
	return &Model{state: m.state, tools: copied}, nil
}

func (m *Model) GenerateCalls() []Call {
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	return copyCalls(m.state.generateCalls)
}

func (m *Model) StreamCalls() []Call {
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	return copyCalls(m.state.streamCalls)
}

func (m *Model) WithToolsCalls() [][]*schema.ToolInfo {
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	calls := make([][]*schema.ToolInfo, len(m.state.withToolsCalls))
	for i, call := range m.state.withToolsCalls {
		calls[i] = copyTools(call)
	}
	return calls
}

func copyMessages(in []*schema.Message) []*schema.Message {
	out := make([]*schema.Message, len(in))
	copy(out, in)
	return out
}

func copyTools(in []*schema.ToolInfo) []*schema.ToolInfo {
	out := make([]*schema.ToolInfo, len(in))
	copy(out, in)
	return out
}

func copyOptions(in []model.Option) []model.Option {
	out := make([]model.Option, len(in))
	copy(out, in)
	return out
}

func copyCalls(in []Call) []Call {
	out := make([]Call, len(in))
	for i, call := range in {
		out[i] = Call{
			Input: copyMessages(call.Input),
			Tools: copyTools(call.Tools),
			Opts:  copyOptions(call.Opts),
		}
	}
	return out
}
