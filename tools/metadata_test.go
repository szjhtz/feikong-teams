package tools

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
)

type stubTool struct {
	info *schema.ToolInfo
}

func (t stubTool) Info(context.Context) (*schema.ToolInfo, error) {
	return t.info, nil
}

func TestGitToolClassification(t *testing.T) {
	readOnly := []string{"git_status", "git_log", "git_diff"}
	for _, name := range readOnly {
		if !IsReadOnlyTool(name) {
			t.Fatalf("%s should be read-only", name)
		}
		if IsDestructiveTool(name) {
			t.Fatalf("%s should not be destructive", name)
		}
	}

	destructive := []string{
		"git_init", "git_add", "git_commit", "git_checkout", "git_reset",
		"git_remove", "git_branch", "git_tag", "git_remote", "git_config", "git_clean",
	}
	for _, name := range destructive {
		if IsReadOnlyTool(name) {
			t.Fatalf("%s should not be read-only", name)
		}
		if !IsDestructiveTool(name) {
			t.Fatalf("%s should be destructive", name)
		}
	}
}

func TestClassifyToolSetsMetadata(t *testing.T) {
	readTool := stubTool{info: &schema.ToolInfo{Name: "git_status"}}
	ClassifyTool(readTool)
	if readTool.info.Extra[metaReadOnly] != true {
		t.Fatalf("expected read-only metadata for git_status")
	}
	if readTool.info.Extra[metaDestructive] == true {
		t.Fatalf("did not expect destructive metadata for git_status")
	}

	writeTool := stubTool{info: &schema.ToolInfo{Name: "git_clean"}}
	ClassifyTool(writeTool)
	if writeTool.info.Extra[metaReadOnly] == true {
		t.Fatalf("did not expect read-only metadata for git_clean")
	}
	if writeTool.info.Extra[metaDestructive] != true {
		t.Fatalf("expected destructive metadata for git_clean")
	}
}
