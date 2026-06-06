package handler

import (
	"context"
	"encoding/json"
	"fkteams/agentcore"
	"fkteams/engine"
	"fkteams/events"
	"fkteams/events/log"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ChatRequest HTTP 聊天请求
type ChatRequest struct {
	SessionID string        `json:"session_id"`
	Message   string        `json:"message"`
	Mode      string        `json:"mode"`
	AgentName string        `json:"agent_name"`
	Stream    bool          `json:"stream"`
	Contents  []ContentPart `json:"contents"`
}

// ChatHandler HTTP POST 聊天处理器，支持普通 JSON 响应和 SSE 流式响应
func ChatHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ChatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
			return
		}

		if req.Message == "" && len(req.Contents) == 0 {
			Fail(c, http.StatusBadRequest, "message or contents is required")
			return
		}

		sessionID := req.SessionID
		if sessionID == "" {
			sessionID = uuid.New().String()
		}
		mode := req.Mode
		if mode == "" {
			mode = "team"
		}

		ctx := c.Request.Context()
		r, err := resolveRunner(ctx, mode, req.AgentName)
		if err != nil {
			log.Printf("failed to resolve runner: mode=%s, agent=%s, err=%v", mode, req.AgentName, err)
			status := http.StatusInternalServerError
			if req.AgentName != "" {
				status = http.StatusBadRequest
			}
			Fail(c, status, err.Error())
			return
		}

		recorder := eventlog.GlobalSessionManager.GetOrCreate(sessionID, historyDir)
		turnInput, userDisplayText := buildChatInput(recorder, req.Message, req.Contents)

		if req.Stream {
			handleStreamChat(c, ctx, r, recorder, turnInput, sessionID, userDisplayText)
		} else {
			handleSyncChat(c, ctx, r, recorder, turnInput, sessionID, userDisplayText)
		}
	}
}

// handleStreamChat SSE 流式聊天响应
func handleStreamChat(c *gin.Context, ctx context.Context, r agentcore.Runner, recorder *eventlog.HistoryRecorder, turnInput engine.TurnInput, sessionID, userDisplayText string) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	taskCtx, taskCancel := context.WithCancel(ctx)
	defer taskCancel()

	_, runErr := engine.NewSession(r, sessionID).
		WithInput(turnInput).
		OnEvent(func(event events.Event) error {
			recorder.RecordEvent(event)
			data, _ := json.Marshal(convertEventToMap(event))
			_, err := fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			c.Writer.Flush()
			return err
		}).
		WithHistory(recorder).
		OnFinish(func(ctx context.Context, _ *agentcore.RunResult, err error) {
			if err != nil {
				if isConnectionClosed(ctx, err) {
					log.Printf("connection closed, stopping: session=%s", sessionID)
					saveHistory(recorder, chatHistoryPath(sessionID), sessionID)
					return
				}
				log.Printf("error processing event: %v", err)
				finishErrorChat(recorder, sessionID, userDisplayText, err)
				data, _ := json.Marshal(map[string]string{"type": string(events.NotifyError), "error": err.Error()})
				fmt.Fprintf(c.Writer, "data: %s\n\n", data)
				c.Writer.Flush()
				return
			}
			finishChat(recorder, sessionID, userDisplayText)
			data, _ := json.Marshal(map[string]string{"type": string(events.NotifyProcessingEnd), "message": "处理完成"})
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			c.Writer.Flush()
		}).
		Run(taskCtx)
	if runErr != nil && !isConnectionClosed(taskCtx, runErr) {
		log.Printf("stream chat failed: session=%s, err=%v", sessionID, runErr)
	}
}

// handleSyncChat 同步聊天响应（收集完整结果后返回）
func handleSyncChat(c *gin.Context, ctx context.Context, r agentcore.Runner, recorder *eventlog.HistoryRecorder, turnInput engine.TurnInput, sessionID, userDisplayText string) {
	taskCtx, taskCancel := context.WithCancel(ctx)
	defer taskCancel()

	var collectedEvents []events.Event

	_, runErr := engine.NewSession(r, sessionID).
		WithInput(turnInput).
		OnEvent(func(event events.Event) error {
			recorder.RecordEvent(event)
			collectedEvents = append(collectedEvents, event)
			return nil
		}).
		WithHistory(recorder).
		OnFinish(func(ctx context.Context, _ *agentcore.RunResult, err error) {
			if err != nil {
				log.Printf("error processing event: %v", err)
				finishErrorChat(recorder, sessionID, userDisplayText, err)
				return
			}
			finishChat(recorder, sessionID, userDisplayText)
		}).
		Run(taskCtx)
	if runErr != nil {
		Fail(c, http.StatusInternalServerError, runErr.Error())
		return
	}

	var content strings.Builder
	for _, e := range collectedEvents {
		if e.Content != "" {
			content.WriteString(e.Content)
		}
	}

	OK(c, gin.H{
		"session_id": sessionID,
		"content":    content.String(),
		"events":     collectedEvents,
	})
}
