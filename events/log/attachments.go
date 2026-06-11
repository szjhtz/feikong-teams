package eventlog

import (
	"fmt"
	"strings"
	"time"

	"fkteams/agentcore"
)

type AttachmentRef struct {
	ID           string                `json:"id"`
	MessageIndex int                   `json:"message_index"`
	EventIndex   int                   `json:"event_index"`
	PartIndex    int                   `json:"part_index"`
	AgentName    string                `json:"agent_name"`
	MessageText  string                `json:"message_text,omitempty"`
	StartTime    time.Time             `json:"start_time,omitempty"`
	Part         agentcore.ContentPart `json:"part"`
}

func AttachmentID(messageIndex, eventIndex, partIndex int) string {
	return fmt.Sprintf("history:%06d:%02d:%02d", messageIndex, eventIndex, partIndex)
}

func ListAttachments(messages []AgentMessage) []AttachmentRef {
	var refs []AttachmentRef
	for msgIndex, msg := range messages {
		refs = append(refs, AttachmentsForMessage(msg, msgIndex)...)
	}
	return refs
}

func AttachmentsForMessage(msg AgentMessage, messageIndex int) []AttachmentRef {
	var refs []AttachmentRef
	messageText := strings.TrimSpace(msg.GetTextContent())
	for eventIndex, event := range msg.Events {
		for partIndex, part := range event.ContentParts {
			if !isAttachmentPart(part) {
				continue
			}
			refs = append(refs, AttachmentRef{
				ID:           AttachmentID(messageIndex, eventIndex, partIndex),
				MessageIndex: messageIndex,
				EventIndex:   eventIndex,
				PartIndex:    partIndex,
				AgentName:    msg.AgentName,
				MessageText:  messageText,
				StartTime:    msg.StartTime,
				Part:         part,
			})
		}
	}
	return refs
}

func FindAttachment(messages []AgentMessage, id string) (AttachmentRef, bool) {
	for _, ref := range ListAttachments(messages) {
		if ref.ID == id {
			return ref, true
		}
	}
	return AttachmentRef{}, false
}

func isAttachmentPart(part agentcore.ContentPart) bool {
	return part.Type != "" && part.Type != agentcore.ContentPartText
}
