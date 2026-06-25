package event

import (
	"fkteams/internal/domain/message"
	"time"
)

type Type string

const (
	TypeAgentStart   Type = "agent_start"
	TypeAgentEnd     Type = "agent_end"
	TypeTurnStart    Type = "turn_start"
	TypeTurnEnd      Type = "turn_end"
	TypeMessageStart Type = "message_start"
	TypeMessageDelta Type = "message_delta"
	TypeMessageEnd   Type = "message_end"
	TypeToolStart    Type = "tool_start"
	TypeToolUpdate   Type = "tool_update"
	TypeToolEnd      Type = "tool_end"
	TypeAction       Type = "action"
	TypeUsage        Type = "usage"
	TypeError        Type = "error"
	TypeMemberUpdate Type = "member_update"
)

type DeltaKind string

const (
	DeltaOutput     DeltaKind = "output"
	DeltaReasoning  DeltaKind = "reasoning"
	DeltaToolArgs   DeltaKind = "tool_args"
	DeltaToolResult DeltaKind = "tool_result"
)

type ActionType string

const (
	ActionTransfer             ActionType = "transfer"
	ActionInterrupted          ActionType = "interrupted"
	ActionExit                 ActionType = "exit"
	ActionAskQuestions         ActionType = "ask_questions"
	ActionAskResponse          ActionType = "ask_response"
	ActionApprovalRequired     ActionType = "approval_required"
	ActionApprovalDecision     ActionType = "approval_decision"
	ActionContextCompressStart ActionType = "context_compress_start"
	ActionContextCompress      ActionType = "context_compress"
)

type NotifyType string

const (
	NotifyProcessingStart  NotifyType = "processing_start"
	NotifyProcessingEnd    NotifyType = "processing_end"
	NotifyUserMessage      NotifyType = "user_message"
	NotifyQueueUpdated     NotifyType = "queue_updated"
	NotifyCancelled        NotifyType = "cancelled"
	NotifyError            NotifyType = "error"
	NotifyAskQuestions     NotifyType = "ask_questions"
	NotifyApprovalRequired NotifyType = "approval_required"
	NotifyConnected        NotifyType = "connected"
	NotifyPong             NotifyType = "pong"
	NotifyInvalidAPIKey    NotifyType = "invalid_api_key"
)

type Event struct {
	EventID          string             `json:"event_id,omitempty"`
	Sequence         int64              `json:"sequence,omitempty"`
	CreatedAt        time.Time          `json:"created_at,omitempty"`
	Type             Type               `json:"type"`
	RunID            string             `json:"run_id,omitempty"`
	TurnID           string             `json:"turn_id,omitempty"`
	MessageID        string             `json:"message_id,omitempty"`
	ToolCallID       string             `json:"tool_call_id,omitempty"`
	ToolCallRef      string             `json:"tool_call_ref,omitempty"`
	ParentToolCallID string             `json:"parent_tool_call_id,omitempty"`
	ParentToolName   string             `json:"parent_tool_name,omitempty"`
	AgentName        string             `json:"agent_name,omitempty"`
	RunPath          string             `json:"run_path,omitempty"`
	Role             message.Role       `json:"role,omitempty"`
	DeltaKind        DeltaKind          `json:"delta_kind,omitempty"`
	Content          string             `json:"content,omitempty"`
	Detail           string             `json:"detail,omitempty"`
	ReasoningContent string             `json:"reasoning_content,omitempty"`
	Message          *message.Message   `json:"message,omitempty"`
	ToolCall         *message.ToolCall  `json:"tool_call,omitempty"`
	ToolCalls        []message.ToolCall `json:"tool_calls,omitempty"`
	ToolCallRefs     map[int]string     `json:"tool_call_refs,omitempty"`
	ToolName         string             `json:"tool_name,omitempty"`
	ToolArgs         string             `json:"tool_args,omitempty"`
	ToolResult       string             `json:"tool_result,omitempty"`
	ToolCallIndex    *int               `json:"tool_call_index,omitempty"`
	MemberCallID     string             `json:"member_call_id,omitempty"`
	MemberToolName   string             `json:"member_tool_name,omitempty"`
	MemberName       string             `json:"member_name,omitempty"`
	MemberOrder      *int               `json:"member_order,omitempty"`
	ActionType       ActionType         `json:"action_type,omitempty"`
	Error            string             `json:"error,omitempty"`
	PromptTokens     int                `json:"prompt_tokens,omitempty"`
	CompletionTokens int                `json:"completion_tokens,omitempty"`
	TotalTokens      int                `json:"total_tokens,omitempty"`
}
