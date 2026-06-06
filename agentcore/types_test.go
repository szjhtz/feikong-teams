package agentcore

import "testing"

func TestRunOptionsWithDefaults(t *testing.T) {
	opts := RunOptions{CheckpointID: "checkpoint-1"}.WithDefaults("default-run")

	if opts.RunID != "checkpoint-1" {
		t.Fatalf("run id = %q, want checkpoint-1", opts.RunID)
	}
	if opts.Sink == nil {
		t.Fatal("sink was not defaulted")
	}
	if err := opts.Sink(Event{}); err != nil {
		t.Fatalf("default sink returned error: %v", err)
	}
}

func TestRunOptionsWithDefaultsUsesFallbackRunID(t *testing.T) {
	opts := RunOptions{}.WithDefaults("default-run")

	if opts.RunID != "default-run" {
		t.Fatalf("run id = %q, want default-run", opts.RunID)
	}
}
