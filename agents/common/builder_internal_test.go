package common

import (
	"context"
	"testing"

	"fkteams/internal/testmodel"
)

func TestAgentBuilderBuildDoesNotMutateResolvedTools(t *testing.T) {
	ctx := context.Background()
	builder := NewAgentBuilder("builder_tools_test", "builder tools test agent").
		WithModel(testmodel.New()).
		WithInstruction("test").
		WithToolNames("ask")

	if _, err := builder.Build(ctx); err != nil {
		t.Fatalf("first build: %v", err)
	}
	if len(builder.tools) != 0 {
		t.Fatalf("expected builder tools to remain empty after first build, got %d", len(builder.tools))
	}

	if _, err := builder.Build(ctx); err != nil {
		t.Fatalf("second build: %v", err)
	}
	if len(builder.tools) != 0 {
		t.Fatalf("expected builder tools to remain empty after second build, got %d", len(builder.tools))
	}
}
