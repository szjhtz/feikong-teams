package chatutil

import (
	"fkteams/eventlog"
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestBuildTurnInputReturnsTurnInput(t *testing.T) {
	recorder := eventlog.NewHistoryRecorder()

	input := BuildTurnInput(recorder, "hello")

	if input.UserInput != "hello" {
		t.Fatalf("user input = %q, want hello", input.UserInput)
	}
	if len(input.Messages) == 0 || input.Messages[len(input.Messages)-1].Content != "hello" {
		t.Fatalf("messages = %#v, want final user message", input.Messages)
	}
}

func TestBuildMultimodalTurnInputReturnsDisplayText(t *testing.T) {
	recorder := eventlog.NewHistoryRecorder()
	parts := []schema.MessageInputPart{TextPart("describe this")}

	input := BuildMultimodalTurnInput(recorder, "describe this", parts)

	if input.UserInput != "describe this" {
		t.Fatalf("user input = %q, want display text", input.UserInput)
	}
	if len(input.Messages) == 0 || len(input.Messages[len(input.Messages)-1].UserInputMultiContent) != 1 {
		t.Fatalf("messages = %#v, want multimodal user message", input.Messages)
	}
}

func TestAgentMessageToSchemaMessagesIncludesCancellationNotice(t *testing.T) {
	msg := eventlog.AgentMessage{
		AgentName: "系统",
		Events: []eventlog.MessageEvent{
			{Type: eventlog.MsgTypeCancelled, Content: "任务已取消"},
		},
	}

	messages := agentMessageToSchemaMessages(msg)
	if len(messages) != 1 {
		t.Fatalf("message count = %d, want 1", len(messages))
	}
	if messages[0].Role != schema.Assistant {
		t.Fatalf("role = %q, want %q", messages[0].Role, schema.Assistant)
	}
	if !strings.Contains(messages[0].Content, "用户刚才取消了上一轮任务") {
		t.Fatalf("content = %q, want cancellation notice", messages[0].Content)
	}
}

func TestAgentMessageToSchemaMessagesMarksCancelledAssistantOutput(t *testing.T) {
	msg := eventlog.AgentMessage{
		AgentName: "assistant",
		Events: []eventlog.MessageEvent{
			{Type: eventlog.MsgTypeText, Content: "处理中"},
			{Type: eventlog.MsgTypeCancelled, Content: "任务已取消"},
		},
	}

	messages := agentMessageToSchemaMessages(msg)
	if len(messages) != 1 {
		t.Fatalf("message count = %d, want 1", len(messages))
	}
	if messages[0].Role != schema.Assistant {
		t.Fatalf("role = %q, want %q", messages[0].Role, schema.Assistant)
	}
	if !strings.Contains(messages[0].Content, "[用户取消]") {
		t.Fatalf("content = %q, want cancellation marker", messages[0].Content)
	}
}
