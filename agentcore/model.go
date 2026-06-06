package agentcore

import (
	"context"
	"io"
)

type ChatModel interface {
	RuntimeModel() any
}

type ModelOption struct {
	Runtime any
}

type ModelCall struct {
	Input []Message
	Tools []ToolInfo
	Opts  []ModelOption
}

type NativeChatModel interface {
	ChatModel
	Generate(ctx context.Context, input []Message, opts ...ModelOption) (Message, error)
	Stream(ctx context.Context, input []Message, opts ...ModelOption) (MessageStream, error)
	WithTools(tools []ToolInfo) (ChatModel, error)
}

type MessageStream interface {
	Recv() (Message, error)
	Close()
}

type sliceMessageStream struct {
	messages []Message
	index    int
}

func NewMessageStream(messages []Message) MessageStream {
	copied := make([]Message, len(messages))
	copy(copied, messages)
	return &sliceMessageStream{messages: copied}
}

func (s *sliceMessageStream) Recv() (Message, error) {
	if s.index >= len(s.messages) {
		return Message{}, io.EOF
	}
	msg := s.messages[s.index]
	s.index++
	return msg, nil
}

func (s *sliceMessageStream) Close() {}

type runtimeChatModel struct {
	runtime any
}

func WrapRuntimeChatModel(runtime any) ChatModel {
	return &runtimeChatModel{runtime: runtime}
}

func (m *runtimeChatModel) RuntimeModel() any {
	if m == nil {
		return nil
	}
	return m.runtime
}
