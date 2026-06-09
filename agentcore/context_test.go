package agentcore

import (
	"context"
	"testing"
)

func TestSummaryPersistCallbackContext(t *testing.T) {
	called := false
	cb := SummaryPersistCallback(func(summary string) {
		called = summary == "done"
	})

	ctx := WithSummaryPersistCallback(context.Background(), cb)
	got, ok := SummaryPersistCallbackFromContext(ctx)
	if !ok {
		t.Fatal("expected summary callback in context")
	}
	got("done")
	if !called {
		t.Fatal("summary callback was not invoked")
	}

	if _, ok := SummaryPersistCallbackFromContext(context.Background()); ok {
		t.Fatal("plain context should not contain summary callback")
	}
}

func TestSteeringSourceContext(t *testing.T) {
	ctx := WithSteeringSource(context.Background(), nil)
	if _, ok := SteeringSourceFromContext(ctx); ok {
		t.Fatal("nil steering source should not be stored")
	}

	source := SteeringSource(func(context.Context) ([]Message, error) {
		return []Message{{Role: RoleUser, Content: "steer"}}, nil
	})
	ctx = WithSteeringSource(context.Background(), source)
	got, ok := SteeringSourceFromContext(ctx)
	if !ok {
		t.Fatal("expected steering source in context")
	}
	messages, err := got(context.Background())
	if err != nil {
		t.Fatalf("steering source returned error: %v", err)
	}
	if len(messages) != 1 || messages[0].Content != "steer" {
		t.Fatalf("steering messages = %#v", messages)
	}
}
