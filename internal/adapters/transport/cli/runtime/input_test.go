package runtime

import (
	"strings"
	"testing"
)

func TestReadLimitedInputEnforcesExactBoundary(t *testing.T) {
	data, err := readLimitedInput(strings.NewReader("1234"), 4)
	if err != nil || string(data) != "1234" {
		t.Fatalf("data = %q, err = %v", data, err)
	}
	if _, err := readLimitedInput(strings.NewReader("12345"), 4); err == nil {
		t.Fatal("readLimitedInput accepted oversized input")
	}
}
