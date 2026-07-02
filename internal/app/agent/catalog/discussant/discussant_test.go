package discussant

import (
	"fkteams/internal/app/config"
	"testing"
)

func TestInstructionForMember(t *testing.T) {
	if got := instructionForMember(config.TeamMember{}); got != discussantPrompt {
		t.Fatal("empty member prompt should use builtin prompt")
	}

	if got := instructionForMember(config.TeamMember{Prompt: "  自定义提示词  "}); got != "自定义提示词" {
		t.Fatalf("custom member prompt = %q", got)
	}
}
