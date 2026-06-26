package eventlog

import (
	domainhistory "fkteams/internal/domain/history"
)

func AttachmentID(messageIndex, eventIndex, partIndex int) string {
	return domainhistory.AttachmentID(messageIndex, eventIndex, partIndex)
}

func ListAttachments(messages []AgentMessage) []AttachmentRef {
	return domainhistory.ListAttachments(messages)
}

func AttachmentsForMessage(msg AgentMessage, messageIndex int) []AttachmentRef {
	return domainhistory.AttachmentsForMessage(msg, messageIndex)
}

func FindAttachment(messages []AgentMessage, id string) (AttachmentRef, bool) {
	return domainhistory.FindAttachment(messages, id)
}
