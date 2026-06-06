package eino

import (
	"context"
	"fkteams/agentcore"
	"fmt"
	"io"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

func AdaptChatModelForRunner(m agentcore.ChatModel) (model.ToolCallingChatModel, error) {
	if m == nil {
		return nil, fmt.Errorf("model is nil")
	}
	runtimeModel := m.RuntimeModel()
	if runtimeModel != nil {
		chatModel, ok := runtimeModel.(model.ToolCallingChatModel)
		if !ok {
			return nil, fmt.Errorf("unsupported runtime model: %T", runtimeModel)
		}
		return chatModel, nil
	}
	nativeModel, ok := m.(agentcore.NativeChatModel)
	if !ok {
		return nil, fmt.Errorf("unsupported native model: %T", m)
	}
	return &nativeChatModelAdapter{inner: nativeModel}, nil
}

type nativeChatModelAdapter struct {
	inner agentcore.NativeChatModel
}

func (m *nativeChatModelAdapter) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	msg, err := m.inner.Generate(ctx, adaptMessagesFromRunner(input), adaptModelOptions(opts)...)
	if err != nil {
		return nil, err
	}
	return adaptMessageForRunner(msg), nil
}

func (m *nativeChatModelAdapter) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	stream, err := m.inner.Stream(ctx, adaptMessagesFromRunner(input), adaptModelOptions(opts)...)
	if err != nil {
		return nil, err
	}
	reader, writer := schema.Pipe[*schema.Message](1)
	go func() {
		defer writer.Close()
		defer stream.Close()
		for {
			msg, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					return
				}
				writer.Send(nil, err)
				return
			}
			if writer.Send(adaptMessageForRunner(msg), nil) {
				return
			}
		}
	}()
	return reader, nil
}

func (m *nativeChatModelAdapter) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	coreTools := make([]agentcore.ToolInfo, 0, len(tools))
	for _, t := range tools {
		if t == nil {
			continue
		}
		coreTools = append(coreTools, agentcore.ToolInfo{Name: t.Name, Desc: t.Desc, Extra: t.Extra})
	}
	next, err := m.inner.WithTools(coreTools)
	if err != nil {
		return nil, err
	}
	runnerModel, err := AdaptChatModelForRunner(next)
	if err != nil {
		return nil, err
	}
	return runnerModel, nil
}

func adaptMessageForRunner(msg agentcore.Message) *schema.Message {
	messages := adaptMessagesForRunner([]agentcore.Message{msg})
	if len(messages) == 0 || messages[0] == nil {
		return &schema.Message{}
	}
	return messages[0]
}

func adaptMessagesFromRunner(messages []*schema.Message) []agentcore.Message {
	result := make([]agentcore.Message, 0, len(messages))
	for _, msg := range messages {
		result = append(result, adaptMessageFromRunner(msg))
	}
	return result
}

func adaptModelOptions(opts []model.Option) []agentcore.ModelOption {
	result := make([]agentcore.ModelOption, 0, len(opts))
	for _, opt := range opts {
		result = append(result, agentcore.ModelOption{Runtime: opt})
	}
	return result
}
