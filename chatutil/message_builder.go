// Package chatutil 提供 CLI 和 Web 共享的聊天工具函数
package chatutil

import (
	"fkteams/fkenv"
	"fkteams/fkevent"
	"fkteams/g"
	"fkteams/log"
	"fkteams/memory"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

// BuildInputMessages 构建输入消息列表（长期记忆 + 对话历史 + 用户输入）
func BuildInputMessages(recorder *fkevent.HistoryRecorder, userInput string) []adk.Message {
	var inputMessages []adk.Message

	// 注入长期记忆
	if g.MemoryManager != nil {
		memories := g.MemoryManager.Search(userInput, 5)
		if memCtx := memory.BuildMemoryContext(memories); memCtx != "" {
			inputMessages = append(inputMessages, schema.SystemMessage(memCtx))
		}
	}

	// 对话历史
	inputMessages = append(inputMessages, buildHistoryMessages(recorder)...)

	// 添加用户输入
	inputMessages = append(inputMessages, schema.UserMessage(userInput))

	if debugContextEnabled() {
		logMessages("BuildInputMessages", inputMessages)
	}
	return inputMessages
}

// BuildMultimodalInputMessages 构建多模态输入消息列表（长期记忆 + 对话历史 + 多模态内容）
func BuildMultimodalInputMessages(recorder *fkevent.HistoryRecorder, textContent string, parts []schema.MessageInputPart) []adk.Message {
	var inputMessages []adk.Message

	// 注入长期记忆（使用文本部分进行搜索）
	if g.MemoryManager != nil {
		memories := g.MemoryManager.Search(textContent, 5)
		if memCtx := memory.BuildMemoryContext(memories); memCtx != "" {
			inputMessages = append(inputMessages, schema.SystemMessage(memCtx))
		}
	}

	// 对话历史
	inputMessages = append(inputMessages, buildHistoryMessages(recorder)...)

	// 添加多模态用户输入
	inputMessages = append(inputMessages, &schema.Message{
		Role:                  schema.User,
		UserInputMultiContent: parts,
	})

	if debugContextEnabled() {
		logMessages("BuildMultimodalInputMessages", inputMessages)
	}
	return inputMessages
}

// TextPart 创建文本内容部分
func TextPart(text string) schema.MessageInputPart {
	return schema.MessageInputPart{
		Type: schema.ChatMessagePartTypeText,
		Text: text,
	}
}

// ImageURLPart 创建图片 URL 内容部分
func ImageURLPart(url string, detail ...schema.ImageURLDetail) schema.MessageInputPart {
	d := schema.ImageURLDetailAuto
	if len(detail) > 0 {
		d = detail[0]
	}
	return schema.MessageInputPart{
		Type: schema.ChatMessagePartTypeImageURL,
		Image: &schema.MessageInputImage{
			MessagePartCommon: schema.MessagePartCommon{
				URL: &url,
			},
			Detail: d,
		},
	}
}

// ImageBase64Part 创建 Base64 编码图片内容部分
func ImageBase64Part(base64Data, mimeType string) schema.MessageInputPart {
	return schema.MessageInputPart{
		Type: schema.ChatMessagePartTypeImageURL,
		Image: &schema.MessageInputImage{
			MessagePartCommon: schema.MessagePartCommon{
				Base64Data: &base64Data,
				MIMEType:   mimeType,
			},
		},
	}
}

// AudioURLPart 创建音频 URL 内容部分
func AudioURLPart(url string) schema.MessageInputPart {
	return schema.MessageInputPart{
		Type: schema.ChatMessagePartTypeAudioURL,
		Audio: &schema.MessageInputAudio{
			MessagePartCommon: schema.MessagePartCommon{
				URL: &url,
			},
		},
	}
}

// VideoURLPart 创建视频 URL 内容部分
func VideoURLPart(url string) schema.MessageInputPart {
	return schema.MessageInputPart{
		Type: schema.ChatMessagePartTypeVideoURL,
		Video: &schema.MessageInputVideo{
			MessagePartCommon: schema.MessagePartCommon{
				URL: &url,
			},
		},
	}
}

// FileURLPart 创建文件 URL 内容部分
func FileURLPart(url string) schema.MessageInputPart {
	return schema.MessageInputPart{
		Type: schema.ChatMessagePartTypeFileURL,
		File: &schema.MessageInputFile{
			MessagePartCommon: schema.MessagePartCommon{
				URL: &url,
			},
		},
	}
}

// ExtractTextFromParts 从多模态内容中提取纯文本
func ExtractTextFromParts(parts []schema.MessageInputPart) string {
	var texts []string
	for _, p := range parts {
		if p.Type == schema.ChatMessagePartTypeText && p.Text != "" {
			texts = append(texts, p.Text)
		}
	}
	return strings.Join(texts, " ")
}

// buildHistoryMessages 构建结构化历史消息列表
func buildHistoryMessages(recorder *fkevent.HistoryRecorder) []adk.Message {
	agentMessages := recorder.GetMessages()
	summaryText, summarizedCount := recorder.GetSummary()

	var messages []adk.Message

	if summaryText != "" && summarizedCount > 0 {
		messages = append(messages, schema.SystemMessage(
			"## 对话历史摘要\n"+summaryText+"\n\n以上对话均已处理完毕，请仅回答用户当前的最新问题。",
		))

		// 摘要未覆盖的最近记录
		for _, msg := range agentMessages[summarizedCount:] {
			messages = append(messages, agentMessageToSchemaMessage(msg))
		}
	} else if len(agentMessages) > 0 {
		for _, msg := range agentMessages {
			messages = append(messages, agentMessageToSchemaMessage(msg))
		}
	}

	return messages
}

// debugContextEnabled 检查是否启用上下文调试日志
func debugContextEnabled() bool {
	return fkenv.Get(fkenv.DebugContext) == "1"
}

// logMessages 打印消息列表摘要
func logMessages(tag string, msgs []adk.Message) {
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
		label := m.Name
		preview := truncatePreview(m.Content, 120)
		extra := ""
		if len(m.ToolCalls) > 0 {
			extra += fmt.Sprintf(" tool_calls=%d", len(m.ToolCalls))
		}
		if m.ReasoningContent != "" {
			extra += fmt.Sprintf(" reasoning=%dchars", len([]rune(m.ReasoningContent)))
		}
		if len(m.UserInputMultiContent) > 0 {
			extra += fmt.Sprintf(" multimodal_parts=%d", len(m.UserInputMultiContent))
		}
		if label != "" {
			extra += fmt.Sprintf(" name=%s", label)
		}
		log.Debugf("  [%d] %-9s | %s%s", i+1, role, preview, extra)
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

// agentMessageToSchemaMessage 将 AgentMessage 转为对应角色的 Message
func agentMessageToSchemaMessage(msg fkevent.AgentMessage) *schema.Message {
	content := msg.GetContentWithToolsFiltered()

	if msg.AgentName == "用户" {
		return schema.UserMessage(content)
	}

	m := schema.AssistantMessage(content, nil)
	m.Name = msg.AgentName

	if reasoning := msg.GetReasoningContent(); reasoning != "" {
		m.ReasoningContent = reasoning
	}

	return m
}
