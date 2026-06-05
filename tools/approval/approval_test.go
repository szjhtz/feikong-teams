package approval

import (
	"errors"
	"testing"
)

func TestOperationInfoFormatsReusablePrompt(t *testing.T) {
	op := Operation{
		Title:  "Git operation requires approval",
		Target: "/tmp/repo",
		Details: []OperationDetail{
			{Name: "Operation", Value: "commit"},
			{Name: "Secret", Value: ""},
		},
	}

	got := op.Info()
	want := "Git operation requires approval\n  Target: /tmp/repo\n  Operation: commit"
	if got != want {
		t.Fatalf("unexpected operation info:\n%s", got)
	}
}

func TestOperationInfoUsesDefaultTitle(t *testing.T) {
	op := Operation{}
	if got := op.Info(); got != "Operation requires approval" {
		t.Fatalf("unexpected default info: %q", got)
	}
}

func TestRejectedMessage(t *testing.T) {
	got, ok := RejectedMessage(ErrRejected, "custom rejected")
	if !ok {
		t.Fatal("expected rejection to be detected")
	}
	if got != "custom rejected" {
		t.Fatalf("unexpected rejected message: %q", got)
	}

	if _, ok := RejectedMessage(errors.New("other"), "custom rejected"); ok {
		t.Fatal("unexpected rejection match")
	}
}
