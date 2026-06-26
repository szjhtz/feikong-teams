package testmodel

import (
	"context"
	domainmessage "fkteams/internal/domain/message"
	runtimeport "fkteams/internal/ports/runtime"
	"fmt"
	"sync"
)

type GenerateResult struct {
	Message domainmessage.Message
	Err     error
}

type StreamResult struct {
	Chunks []domainmessage.Message
	Err    error
}

type Call struct {
	Input []domainmessage.Message
	Tools []runtimeport.ToolInfo
}

type Model struct {
	state *state
	tools []runtimeport.ToolInfo
}

type state struct {
	mu             sync.Mutex
	generateQueue  []GenerateResult
	streamQueue    []StreamResult
	generateCalls  []Call
	streamCalls    []Call
	withToolsCalls [][]runtimeport.ToolInfo
}

var _ runtimeport.ChatModel = (*Model)(nil)

func New(responses ...domainmessage.Message) *Model {
	m := &Model{state: &state{}}
	for _, resp := range responses {
		m.EnqueueGenerate(resp, nil)
	}
	return m
}

func AssistantMessage(content string) domainmessage.Message {
	return domainmessage.Message{Role: domainmessage.RoleAssistant, Content: content}
}

func UserMessage(content string) domainmessage.Message {
	return domainmessage.Message{Role: domainmessage.RoleUser, Content: content}
}

func (m *Model) EnqueueGenerate(message domainmessage.Message, err error) *Model {
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	m.state.generateQueue = append(m.state.generateQueue, GenerateResult{Message: message, Err: err})
	return m
}

func (m *Model) EnqueueStream(chunks ...domainmessage.Message) *Model {
	m.EnqueueStreamResult(chunks, nil)
	return m
}

func (m *Model) EnqueueStreamResult(chunks []domainmessage.Message, err error) *Model {
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	copied := copyMessages(chunks)
	m.state.streamQueue = append(m.state.streamQueue, StreamResult{Chunks: copied, Err: err})
	return m
}

func (m *Model) Generate(_ context.Context, input []domainmessage.Message) (domainmessage.Message, error) {
	m.state.mu.Lock()
	defer m.state.mu.Unlock()

	m.state.generateCalls = append(m.state.generateCalls, Call{
		Input: copyMessages(input),
		Tools: copyTools(m.tools),
	})
	if len(m.state.generateQueue) == 0 {
		return domainmessage.Message{}, fmt.Errorf("testmodel: no queued generate response")
	}

	resp := m.state.generateQueue[0]
	m.state.generateQueue = m.state.generateQueue[1:]
	return resp.Message, resp.Err
}

func (m *Model) Stream(_ context.Context, input []domainmessage.Message) (runtimeport.MessageStream, error) {
	m.state.mu.Lock()
	defer m.state.mu.Unlock()

	m.state.streamCalls = append(m.state.streamCalls, Call{
		Input: copyMessages(input),
		Tools: copyTools(m.tools),
	})
	if len(m.state.streamQueue) == 0 {
		return nil, fmt.Errorf("testmodel: no queued stream response")
	}

	resp := m.state.streamQueue[0]
	m.state.streamQueue = m.state.streamQueue[1:]
	if resp.Err != nil {
		return nil, resp.Err
	}
	return runtimeport.NewMessageStream(resp.Chunks), nil
}

func (m *Model) WithTools(tools []runtimeport.ToolInfo) (runtimeport.ChatModel, error) {
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

func (m *Model) WithToolsCalls() [][]runtimeport.ToolInfo {
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	calls := make([][]runtimeport.ToolInfo, len(m.state.withToolsCalls))
	for i, call := range m.state.withToolsCalls {
		calls[i] = copyTools(call)
	}
	return calls
}

func copyMessages(in []domainmessage.Message) []domainmessage.Message {
	out := make([]domainmessage.Message, len(in))
	copy(out, in)
	return out
}

func copyTools(in []runtimeport.ToolInfo) []runtimeport.ToolInfo {
	out := make([]runtimeport.ToolInfo, len(in))
	copy(out, in)
	return out
}

func copyCalls(in []Call) []Call {
	out := make([]Call, len(in))
	for i, call := range in {
		out[i] = Call{
			Input: copyMessages(call.Input),
			Tools: copyTools(call.Tools),
		}
	}
	return out
}
