package eventlog

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

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

func transcriptIDPrefix(typ TranscriptEventType) string {
	switch typ {
	case TranscriptUserMessage, TranscriptAssistantMessage:
		return "msg"
	case TranscriptToolCallStart, TranscriptToolCallEnd:
		return "tool"
	case TranscriptAskRequested, TranscriptAskAnswered:
		return "ask"
	case TranscriptError:
		return "err"
	case TranscriptSystemNotice:
		return "notice"
	case TranscriptCancelled:
		return "cancel"
	default:
		return "rec"
	}
}

func transcriptTurnID(sessionID string, turn int) string {
	if turn <= 0 {
		turn = 1
	}
	if sessionID == "" {
		return fmt.Sprintf("history:turn:%d", turn)
	}
	return fmt.Sprintf("%s:history:turn:%d", sessionID, turn)
}

func transcriptTurnFromRuntimeID(turnID string) int {
	turnID = strings.TrimSpace(turnID)
	if turnID == "" {
		return 0
	}
	index := strings.LastIndex(turnID, ":turn:")
	if index < 0 {
		return 0
	}
	n, err := strconv.Atoi(turnID[index+len(":turn:"):])
	if err != nil || n < 0 {
		return 0
	}
	return n
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
		event.ID = newPrefixedID(transcriptIDPrefix(event.Type))
	}
	if event.At.IsZero() {
		event.At = time.Now()
	}
	if subagent == nil {
		_ = appendJSONL(transcriptPath(h.sessionDir), event)
		return
	}
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

type toolResultPayload struct {
	Result        string
	ResultRef     string
	Summary       string
	Truncated     bool
	OriginalChars int
}

func (h *HistoryRecorder) toolResultPayload(toolName, content string) toolResultPayload {
	payload := toolResultPayload{
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

func usageRecordFromEvent(event Event) *UsageRecord {
	promptTokens := event.PromptTokens
	completionTokens := event.CompletionTokens
	totalTokens := event.TotalTokens
	if event.Usage != nil {
		if promptTokens == 0 {
			promptTokens = event.Usage.PromptTokens
		}
		if completionTokens == 0 {
			completionTokens = event.Usage.CompletionTokens
		}
		if totalTokens == 0 {
			totalTokens = event.Usage.TotalTokens
		}
	}
	if promptTokens == 0 && completionTokens == 0 && totalTokens == 0 {
		return nil
	}
	return &UsageRecord{PromptTokens: promptTokens, CompletionTokens: completionTokens, TotalTokens: totalTokens}
}
