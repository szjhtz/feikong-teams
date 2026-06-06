// Package chatutil 提供 CLI 和 Web 共享的聊天工具函数
package chatutil

import (
	"fkteams/agentcore"
	"fkteams/engine"
	"fkteams/eventlog"
	"fkteams/fkenv"
	"fkteams/g"
	"fkteams/log"
	"fkteams/memory"
	"fmt"
	"strings"
)

// BuildTurnInput 构建一轮输入（长期记忆 + 对话历史 + 用户输入）
func BuildTurnInput(recorder *eventlog.HistoryRecorder, userInput string) engine.TurnInput {
	var contextMessages []agentcore.Message

	// 注入长期记忆
	if g.MemoryManager != nil {
		memories := g.MemoryManager.Search(userInput, 5)
		if memCtx := memory.BuildMemoryContext(memories); memCtx != "" {
			contextMessages = append(contextMessages, agentcore.Message{Role: agentcore.RoleSystem, Content: memCtx})
		}
	}

	// 对话历史
	contextMessages = append(contextMessages, buildHistoryMessages(recorder)...)
	message := agentcore.Message{Role: agentcore.RoleUser, Content: userInput}

	if debugContextEnabled() {
		logMessages("BuildTurnInput", append(contextMessages, message))
	}
	return engine.TurnInput{
		Context: contextMessages,
		Message: message,
	}
}

// BuildMultimodalTurnInput 构建一轮多模态输入（长期记忆 + 对话历史 + 多模态内容）
func BuildMultimodalTurnInput(recorder *eventlog.HistoryRecorder, textContent string, parts []agentcore.ContentPart) engine.TurnInput {
	var contextMessages []agentcore.Message

	// 注入长期记忆（使用文本部分进行搜索）
	if g.MemoryManager != nil {
		memories := g.MemoryManager.Search(textContent, 5)
		if memCtx := memory.BuildMemoryContext(memories); memCtx != "" {
			contextMessages = append(contextMessages, agentcore.Message{Role: agentcore.RoleSystem, Content: memCtx})
		}
	}

	// 对话历史
	contextMessages = append(contextMessages, buildHistoryMessages(recorder)...)
	message := agentcore.Message{
		Role:                  agentcore.RoleUser,
		UserInputMultiContent: parts,
	}

	if debugContextEnabled() {
		logMessages("BuildMultimodalTurnInput", append(contextMessages, message))
	}
	return engine.TurnInput{
		Context: contextMessages,
		Message: message,
	}
}

// TextPart 创建文本内容部分
func TextPart(text string) agentcore.ContentPart {
	return agentcore.ContentPart{
		Type: agentcore.ContentPartText,
		Text: text,
	}
}

// ImageURLPart 创建图片 URL 内容部分
func ImageURLPart(url string, detail ...string) agentcore.ContentPart {
	d := "auto"
	if len(detail) > 0 {
		d = detail[0]
	}
	return agentcore.ContentPart{
		Type:   agentcore.ContentPartImageURL,
		URL:    url,
		Detail: d,
	}
}

// ImageBase64Part 创建 Base64 编码图片内容部分
func ImageBase64Part(base64Data, mimeType string) agentcore.ContentPart {
	return agentcore.ContentPart{
		Type:       agentcore.ContentPartImageURL,
		Base64Data: base64Data,
		MIMEType:   mimeType,
	}
}

// AudioURLPart 创建音频 URL 内容部分
func AudioURLPart(url string) agentcore.ContentPart {
	return agentcore.ContentPart{
		Type: agentcore.ContentPartAudioURL,
		URL:  url,
	}
}

// VideoURLPart 创建视频 URL 内容部分
func VideoURLPart(url string) agentcore.ContentPart {
	return agentcore.ContentPart{
		Type: agentcore.ContentPartVideoURL,
		URL:  url,
	}
}

// FileURLPart 创建文件 URL 内容部分
func FileURLPart(url string) agentcore.ContentPart {
	return agentcore.ContentPart{
		Type: agentcore.ContentPartFileURL,
		URL:  url,
	}
}

// ExtractTextFromParts 从多模态内容中提取纯文本
func ExtractTextFromParts(parts []agentcore.ContentPart) string {
	var texts []string
	for _, p := range parts {
		if p.Type == agentcore.ContentPartText && p.Text != "" {
			texts = append(texts, p.Text)
		}
	}
	return strings.Join(texts, " ")
}

// buildHistoryMessages 构建结构化历史消息列表
func buildHistoryMessages(recorder *eventlog.HistoryRecorder) []agentcore.Message {
	agentMessages := recorder.GetMessages()
	summaryText, summarizedCount := recorder.GetSummary()

	var messages []agentcore.Message

	if summaryText != "" && summarizedCount > 0 {
		messages = append(messages, agentcore.Message{Role: agentcore.RoleSystem, Content: "## 对话历史摘要\n" + summaryText + "\n\n以上对话均已处理完毕，请仅回答用户当前的最新问题。"})

		// 摘要未覆盖的最近记录
		for _, msg := range agentMessages[summarizedCount:] {
			messages = append(messages, agentMessageToCoreMessages(msg)...)
		}
	} else if len(agentMessages) > 0 {
		for _, msg := range agentMessages {
			messages = append(messages, agentMessageToCoreMessages(msg)...)
		}
	}

	return messages
}

// debugContextEnabled 检查是否启用上下文调试日志
func debugContextEnabled() bool {
	return fkenv.Get(fkenv.DebugContext) == "1"
}

// logMessages 打印消息列表摘要
func logMessages(tag string, msgs []agentcore.Message) {
	totalChars := 0
	for _, m := range msgs {
		totalChars += len(m.Content)
		if m.ReasoningContent != "" {
			totalChars += len(m.ReasoningContent)
		}
	}
	log.Debugf("[%s] 共 %d 条消息, 约 %d 字符", tag, len(msgs), totalChars)
	for i, m := range msgs {
		role := string(m.Role)
		preview := truncatePreview(m.Content, 120)
		extra := ""

		// 工具调用：拆分为独立的 tool_call / tool_result 展示
		if len(m.ToolCalls) > 0 {
			for _, tc := range m.ToolCalls {
				name := tc.Function.Name
				if tc.Function.Arguments != "" {
					name += "(" + tc.Function.Arguments + ")"
				}
				log.Debugf("  [%d] %-10s | %s", i+1, "tool_call", truncatePreview(name, 160))
			}
			continue
		}
		if m.Role == agentcore.RoleTool {
			log.Debugf("  [%d] %-10s | %s%s", i+1, "tool_result", preview, extra)
			continue
		}

		if m.ReasoningContent != "" {
			extra += fmt.Sprintf(" reasoning=%dchars", len([]rune(m.ReasoningContent)))
		}
		if len(m.UserInputMultiContent) > 0 {
			extra += fmt.Sprintf(" multimodal_parts=%d", len(m.UserInputMultiContent))
		}
		if m.Name != "" {
			extra += fmt.Sprintf(" name=%s", m.Name)
		}
		log.Debugf("  [%d] %-10s | %s%s", i+1, role, preview, extra)
	}
}

func truncatePreview(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "..."
}

// agentMessageToCoreMessages 将 AgentMessage 转为结构化消息列表。
// 用户消息 → UserMessage；Agent 消息 → 文本 AssistantMessage + 工具调用拆分为 ToolCall/ToolMessage 对。
func agentMessageToCoreMessages(msg eventlog.AgentMessage) []agentcore.Message {
	if msg.AgentName == "用户" {
		var text strings.Builder
		var parts []agentcore.ContentPart
		for _, event := range msg.Events {
			if event.Type != eventlog.MsgTypeText {
				continue
			}
			text.WriteString(event.Content)
			parts = append(parts, event.ContentParts...)
		}
		message := agentcore.Message{Role: agentcore.RoleUser, Content: text.String()}
		if len(parts) > 0 {
			message.Content = ""
			message.UserInputMultiContent = parts
		}
		return []agentcore.Message{message}
	}

	var messages []agentcore.Message
	var textBuf strings.Builder
	var reasoningBuf strings.Builder

	flushText := func() {
		content := strings.TrimSpace(textBuf.String())
		reasoning := strings.TrimSpace(reasoningBuf.String())
		textBuf.Reset()
		reasoningBuf.Reset()
		if content == "" && reasoning == "" {
			return
		}
		m := agentcore.Message{Role: agentcore.RoleAssistant, Content: content}
		m.Name = msg.AgentName
		if reasoning != "" {
			m.ReasoningContent = reasoning
		}
		messages = append(messages, m)
	}

	for _, event := range msg.Events {
		switch event.Type {
		case eventlog.MsgTypeText:
			textBuf.WriteString(event.Content)

		case eventlog.MsgTypeReasoning:
			reasoningBuf.WriteString(event.Content)

		case eventlog.MsgTypeToolCall:
			tc := event.ToolCall
			if tc == nil {
				continue
			}
			flushText()
			// AssistantMessage 携带 ToolCall
			messages = append(messages, agentcore.Message{Role: agentcore.RoleAssistant, ToolCalls: []agentcore.ToolCall{{
				ID:   tc.ID,
				Type: "function",
				Function: agentcore.FunctionCall{
					Name:      tc.Name,
					Arguments: tc.Arguments,
				},
			}}})
			// ToolMessage 携带结果
			messages = append(messages, agentcore.Message{Role: agentcore.RoleTool, Content: tc.Result, ToolCallID: tc.ID, ToolName: tc.Name})

		case eventlog.MsgTypeAction:
			if event.Action != nil && (event.Action.ActionType != "" || event.Action.Content != "") {
				fmt.Fprintf(&textBuf, "[%s] %s\n", event.Action.ActionType, event.Action.Content)
			}

		case eventlog.MsgTypeError:
			fmt.Fprintf(&textBuf, "[错误] %s\n", event.Content)

		case eventlog.MsgTypeCancelled:
			fmt.Fprintf(&textBuf, "[用户取消] %s\n", cancellationNotice(event.Content))
		}
	}

	flushText()
	return messages
}

func cancellationNotice(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		content = "任务已取消"
	}
	return content + "。用户刚才取消了上一轮任务；继续对话时不要把上一轮未完成的执行当作已经完成。"
}
