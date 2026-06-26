package runtime

import (
	"context"
	"fkteams/internal/domain/message"
	"io"
)

type ChatModel interface {
	Generate(ctx context.Context, input []message.Message) (message.Message, error)
	Stream(ctx context.Context, input []message.Message) (MessageStream, error)
	WithTools(tools []ToolInfo) (ChatModel, error)
}

type ModelCall struct {
	Input []message.Message
	Tools []ToolInfo
}

type MessageStream interface {
	Recv() (message.Message, error)
	Close()
}

type sliceMessageStream struct {
	messages []message.Message
	index    int
}

func NewMessageStream(messages []message.Message) MessageStream {
	copied := make([]message.Message, len(messages))
	copy(copied, messages)
	return &sliceMessageStream{messages: copied}
}

func (s *sliceMessageStream) Recv() (message.Message, error) {
	if s.index >= len(s.messages) {
		return message.Message{}, io.EOF
	}
	msg := s.messages[s.index]
	s.index++
	return msg, nil
}

func (s *sliceMessageStream) Close() {}
