package eventlog

import (
	appchat "fkteams/internal/app/chat"
	"fkteams/internal/domain/message"
)

// ProjectContextMessages 返回从 offset 开始的对话上下文消息投影。
func (h *HistoryRecorder) ProjectContextMessages(offset int) []message.Message {
	messages := h.GetMessages()
	if offset < 0 {
		offset = 0
	}
	if offset > len(messages) {
		offset = len(messages)
	}
	return appchat.ProjectAgentMessages(messages[offset:])
}
