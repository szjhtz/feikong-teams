package agentcore

import (
	"errors"
	"io"
	"testing"
)

func TestNewMessageStreamCopiesAndReadsMessages(t *testing.T) {
	messages := []Message{
		{Role: RoleAssistant, Content: "one"},
		{Role: RoleAssistant, Content: "two"},
	}
	stream := NewMessageStream(messages)
	messages[0].Content = "changed"

	msg, err := stream.Recv()
	if err != nil {
		t.Fatalf("first recv: %v", err)
	}
	if msg.Content != "one" {
		t.Fatalf("stream should use copied messages, got %q", msg.Content)
	}
	msg, err = stream.Recv()
	if err != nil {
		t.Fatalf("second recv: %v", err)
	}
	if msg.Content != "two" {
		t.Fatalf("second message = %q", msg.Content)
	}
	if _, err := stream.Recv(); !errors.Is(err, io.EOF) {
		t.Fatalf("final recv error = %v, want EOF", err)
	}
	stream.Close()
}
