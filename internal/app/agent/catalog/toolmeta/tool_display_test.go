package toolmeta

import "testing"

func TestFormatToolDisplayDefaultsToToolKind(t *testing.T) {
	registry := NewRegistry()

	display := registry.FormatToolDisplay("file_read")
	if display.Name != "file_read" {
		t.Fatalf("name = %q, want file_read", display.Name)
	}
	if display.DisplayName != "file_read" {
		t.Fatalf("display name = %q, want file_read", display.DisplayName)
	}
	if display.Kind != ToolKindTool {
		t.Fatalf("kind = %q, want %q", display.Kind, ToolKindTool)
	}
	if display.Target != "" {
		t.Fatalf("target = %q, want empty", display.Target)
	}
}

func TestRegisterAgentToolDisplayUsesExplicitDisplayName(t *testing.T) {
	registry := NewRegistry()

	registry.RegisterAgentToolDisplay("ask_fkagent_coder", "代码助手")

	display := registry.FormatToolDisplay("ask_fkagent_coder")
	if display.Name != "ask_fkagent_coder" {
		t.Fatalf("name = %q, want ask_fkagent_coder", display.Name)
	}
	if display.DisplayName != "指派给 代码助手" {
		t.Fatalf("display name = %q, want 指派给 代码助手", display.DisplayName)
	}
	if display.Kind != ToolKindAgent {
		t.Fatalf("kind = %q, want %q", display.Kind, ToolKindAgent)
	}
	if display.Target != "代码助手" {
		t.Fatalf("target = %q, want 代码助手", display.Target)
	}
}

func TestRegisterAgentToolDisplayDerivesTargetFromToolName(t *testing.T) {
	registry := NewRegistry()

	registry.RegisterAgentToolDisplay("ask_fkagent_data-analyst", "")

	display := registry.FormatToolDisplay("ask_fkagent_data-analyst")
	if display.DisplayName != "指派给 Data Analyst" {
		t.Fatalf("display name = %q, want 指派给 Data Analyst", display.DisplayName)
	}
	if display.Target != "Data Analyst" {
		t.Fatalf("target = %q, want Data Analyst", display.Target)
	}
}

func TestRegisterAgentToolDisplayIgnoresEmptyName(t *testing.T) {
	registry := NewRegistry()

	registry.RegisterAgentToolDisplay("", "Nobody")

	display := registry.FormatToolDisplay("")
	if display.Kind != ToolKindTool {
		t.Fatalf("kind = %q, want %q", display.Kind, ToolKindTool)
	}
	if display.DisplayName != "" {
		t.Fatalf("display name = %q, want empty", display.DisplayName)
	}
}

func TestTitleIdentifier(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "underscore", in: "data_analyst", want: "Data Analyst"},
		{name: "hyphen", in: "deep-researcher", want: "Deep Researcher"},
		{name: "mixed case", in: "CoDeR", want: "Coder"},
		{name: "empty parts", in: "tasker__runner", want: "Tasker Runner"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := titleIdentifier(tt.in); got != tt.want {
				t.Fatalf("titleIdentifier(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
