package apperror

import (
	"errors"
	"testing"
)

func TestWrapPreservesCodeAndCause(t *testing.T) {
	cause := errors.New("disk failed")
	err := Wrap(CodeUnavailable, "storage unavailable", cause)
	if !IsCode(err, CodeUnavailable) {
		t.Fatalf("unexpected code: %s", CodeOf(err))
	}
	if !errors.Is(err, cause) {
		t.Fatal("wrapped cause was not preserved")
	}
	if got := PublicMessage(err); got != "storage unavailable" {
		t.Fatalf("unexpected public message: %q", got)
	}
}

func TestUnknownErrorIsInternal(t *testing.T) {
	err := errors.New("secret detail")
	if got := CodeOf(err); got != CodeInternal {
		t.Fatalf("unexpected code: %s", got)
	}
	if got := PublicMessage(err); got != "internal server error" {
		t.Fatalf("unexpected public message: %q", got)
	}
}
