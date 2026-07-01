package eventlog

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"fkteams/internal/domain/message"
	"fkteams/internal/runtime/atomicfile"

	"github.com/google/uuid"
)

const (
	toolResultsDirName = "tool-results"
	subagentsDirName   = "subagents"

	longToolResultChars = 4000
	resultSummaryChars  = 1200
)

func newPrefixedID(prefix string) string {
	return prefix + "_" + uuid.NewString()
}

func transcriptPath(sessionDir string) string {
	return filepath.Join(sessionDir, TranscriptFileName)
}

func subagentTranscriptPath(sessionDir, agentRunID string) string {
	return filepath.Join(sessionDir, subagentsDirName, filepath.Base(agentRunID)+".jsonl")
}

func (h *HistoryRecorder) appendTranscriptEvent(event TranscriptEvent, subagent *subagentRun) {
	if h == nil || h.sessionDir == "" {
		return
	}
	if event.ID == "" {
		event.ID = newPrefixedID("evt")
	}
	if event.TS.IsZero() {
		event.TS = time.Now()
	}
	if subagent == nil {
		h.nextSeq++
		event.Seq = h.nextSeq
		_ = appendJSONL(transcriptPath(h.sessionDir), event)
		return
	}
	subagent.Seq++
	event.Seq = subagent.Seq
	event.AgentRunID = subagent.AgentRunID
	event.ParentToolCallID = subagent.ParentToolCallID
	if event.Agent == "" {
		event.Agent = subagent.AgentName
	}
	_ = appendJSONL(subagent.TranscriptPath, event)
}

func appendJSONL(filePath string, value any) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open jsonl: %w", err)
	}
	defer file.Close()
	if err := json.NewEncoder(file).Encode(value); err != nil {
		return fmt.Errorf("append jsonl: %w", err)
	}
	return nil
}

func loadTranscript(filePath string) ([]TranscriptEvent, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("read transcript: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var events []TranscriptEvent
	for line := 1; ; line++ {
		var event TranscriptEvent
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("decode transcript record %d: %w", line, err)
		}
		events = append(events, event)
	}
	return events, nil
}

func shouldExternalizeToolResult(toolName, content string) bool {
	if utf8.RuneCountInString(content) > longToolResultChars {
		return true
	}
	for _, prefix := range NoisyToolPrefixes {
		if strings.HasPrefix(toolName, prefix) && utf8.RuneCountInString(content) > resultSummaryChars {
			return true
		}
	}
	return false
}

func summarizeToolResult(content string) string {
	content = strings.TrimSpace(content)
	runes := []rune(content)
	if len(runes) <= resultSummaryChars {
		return content
	}
	return string(runes[:resultSummaryChars]) + "\n...(truncated)..."
}

func (h *HistoryRecorder) toolResultPayload(toolName, content string) TranscriptPayload {
	payload := TranscriptPayload{
		Result: content,
	}
	if h == nil || h.sessionDir == "" || content == "" || !shouldExternalizeToolResult(toolName, content) {
		return payload
	}
	id := newPrefixedID("result")
	rel := filepath.Join(toolResultsDirName, id+".json")
	artifact := ToolResultArtifact{
		ID:            id,
		ToolName:      toolName,
		Content:       content,
		Summary:       summarizeToolResult(content),
		OriginalChars: utf8.RuneCountInString(content),
		CreatedAt:     time.Now(),
	}
	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return payload
	}
	if err := atomicfile.WriteFile(filepath.Join(h.sessionDir, rel), data, 0644); err != nil {
		return payload
	}
	payload.Result = ""
	payload.ResultRef = rel
	payload.Summary = artifact.Summary
	payload.Truncated = true
	payload.OriginalChars = artifact.OriginalChars
	return payload
}

func transcriptRoleFromEvent(event Event) message.Role {
	if event.Role != "" {
		return event.Role
	}
	return message.RoleAssistant
}
