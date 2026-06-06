package chatutil

import (
	"fkteams/agentcore"
	"fkteams/eventlog"
	"strings"
	"testing"
)

func TestBuildTurnInputReturnsTurnInput(t *testing.T) {
	recorder := eventlog.NewHistoryRecorder()

	input := BuildTurnInput(recorder, "hello")

	if input.Message.DisplayText() != "hello" {
		t.Fatalf("input message = %q, want hello", input.Message.DisplayText())
	}
	messages := input.AllMessages()
	if len(messages) == 0 || messages[len(messages)-1].Content != "hello" {
		t.Fatalf("messages = %#v, want final user message", messages)
	}
}

func TestBuildMultimodalTurnInputReturnsDisplayText(t *testing.T) {
	recorder := eventlog.NewHistoryRecorder()
	parts := []agentcore.ContentPart{TextPart("describe this")}

	input := BuildMultimodalTurnInput(recorder, "describe this", parts)

	if input.Message.DisplayText() != "describe this" {
		t.Fatalf("input message = %q, want display text", input.Message.DisplayText())
	}
	messages := input.AllMessages()
	if len(messages) == 0 || len(messages[len(messages)-1].UserInputMultiContent) != 1 {
		t.Fatalf("messages = %#v, want multimodal user message", messages)
	}
}

func TestHistoryRecorderKeepsMultimodalUserInput(t *testing.T) {
	recorder := eventlog.NewHistoryRecorder()
	parts := []agentcore.ContentPart{
		TextPart("describe this"),
		ImageURLPart("https://example.com/a.png", "high"),
	}
	recorder.RecordUserMessage(agentcore.Message{
		Role:                  agentcore.RoleUser,
		UserInputMultiContent: parts,
	})

	input := BuildTurnInput(recorder, "continue")
	if len(input.Context) == 0 {
		t.Fatal("expected history context")
	}
	historyMessage := input.Context[0]
	if historyMessage.Role != agentcore.RoleUser {
		t.Fatalf("history role = %q, want user", historyMessage.Role)
	}
	if len(historyMessage.UserInputMultiContent) != 2 {
		t.Fatalf("history parts = %#v, want 2 parts", historyMessage.UserInputMultiContent)
	}
	if historyMessage.UserInputMultiContent[1].URL != "https://example.com/a.png" {
		t.Fatalf("image url = %q, want https://example.com/a.png", historyMessage.UserInputMultiContent[1].URL)
	}
}

func TestAgentMessageToSchemaMessagesIncludesCancellationNotice(t *testing.T) {
	msg := eventlog.AgentMessage{
		AgentName: "系统",
		Events: []eventlog.MessageEvent{
			{Type: eventlog.MsgTypeCancelled, Content: "任务已取消"},
		},
	}

	messages := agentMessageToCoreMessages(msg)
	if len(messages) != 1 {
		t.Fatalf("message count = %d, want 1", len(messages))
	}
	if messages[0].Role != agentcore.RoleAssistant {
		t.Fatalf("role = %q, want %q", messages[0].Role, agentcore.RoleAssistant)
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

	messages := agentMessageToCoreMessages(msg)
	if len(messages) != 1 {
		t.Fatalf("message count = %d, want 1", len(messages))
	}
	if messages[0].Role != agentcore.RoleAssistant {
		t.Fatalf("role = %q, want %q", messages[0].Role, agentcore.RoleAssistant)
	}
	if !strings.Contains(messages[0].Content, "[用户取消]") {
		t.Fatalf("content = %q, want cancellation marker", messages[0].Content)
	}
}
