package eino

import (
	"fkteams/agentcore"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

func adaptMessagesForRunner(messages []agentcore.Message) []adk.Message {
	result := make([]adk.Message, 0, len(messages))
	for _, msg := range messages {
		m := &schema.Message{
			Role:                  adaptRoleForRunner(msg.Role),
			Content:               msg.Content,
			ReasoningContent:      msg.ReasoningContent,
			ToolCallID:            msg.ToolCallID,
			ToolName:              msg.ToolName,
			Name:                  msg.Name,
			UserInputMultiContent: adaptPartsForRunner(msg.UserInputMultiContent),
			MultiContent:          adaptChatPartsForRunner(msg.MultiContent),
		}
		if len(msg.ToolCalls) > 0 {
			m.ToolCalls = adaptToolCallsForRunner(msg.ToolCalls)
		}
		result = append(result, m)
	}
	return result
}

func adaptMessageFromRunner(msg *schema.Message) agentcore.Message {
	if msg == nil {
		return agentcore.Message{}
	}
	return agentcore.Message{
		Role:                  adaptRoleFromRunner(msg.Role),
		Content:               msg.Content,
		ReasoningContent:      msg.ReasoningContent,
		ToolCalls:             adaptToolCallsFromRunner(msg.ToolCalls),
		ToolCallID:            msg.ToolCallID,
		ToolName:              msg.ToolName,
		UserInputMultiContent: adaptPartsFromRunner(msg.UserInputMultiContent),
		MultiContent:          adaptChatPartsFromRunner(msg.MultiContent),
		Name:                  msg.Name,
	}
}

func adaptRoleForRunner(role agentcore.MessageRole) schema.RoleType {
	switch role {
	case agentcore.RoleSystem:
		return schema.System
	case agentcore.RoleUser:
		return schema.User
	case agentcore.RoleAssistant:
		return schema.Assistant
	case agentcore.RoleTool:
		return schema.Tool
	default:
		return schema.User
	}
}

func adaptRoleFromRunner(role schema.RoleType) agentcore.MessageRole {
	switch role {
	case schema.System:
		return agentcore.RoleSystem
	case schema.User:
		return agentcore.RoleUser
	case schema.Assistant:
		return agentcore.RoleAssistant
	case schema.Tool:
		return agentcore.RoleTool
	default:
		return agentcore.MessageRole(role)
	}
}

func adaptToolCallsForRunner(toolCalls []agentcore.ToolCall) []schema.ToolCall {
	result := make([]schema.ToolCall, 0, len(toolCalls))
	for _, tc := range toolCalls {
		result = append(result, schema.ToolCall{
			ID:    tc.ID,
			Index: tc.Index,
			Type:  tc.Type,
			Function: schema.FunctionCall{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		})
	}
	return result
}

func adaptToolCallsFromRunner(toolCalls []schema.ToolCall) []agentcore.ToolCall {
	result := make([]agentcore.ToolCall, 0, len(toolCalls))
	for _, tc := range toolCalls {
		result = append(result, adaptToolCallFromRunner(tc))
	}
	return result
}

func adaptToolCallFromRunner(tc schema.ToolCall) agentcore.ToolCall {
	return agentcore.ToolCall{
		ID:    tc.ID,
		Index: tc.Index,
		Type:  tc.Type,
		Function: agentcore.FunctionCall{
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		},
	}
}

func adaptPartsForRunner(parts []agentcore.ContentPart) []schema.MessageInputPart {
	result := make([]schema.MessageInputPart, 0, len(parts))
	for _, part := range parts {
		p := schema.MessageInputPart{Text: part.Text}
		switch part.Type {
		case agentcore.ContentPartText:
			p.Type = schema.ChatMessagePartTypeText
		case agentcore.ContentPartImageURL:
			p.Type = schema.ChatMessagePartTypeImageURL
			p.Image = &schema.MessageInputImage{
				MessagePartCommon: schema.MessagePartCommon{
					URL:        stringPtr(part.URL),
					Base64Data: stringPtr(part.Base64Data),
					MIMEType:   part.MIMEType,
				},
				Detail: schema.ImageURLDetail(part.Detail),
			}
		case agentcore.ContentPartAudioURL:
			p.Type = schema.ChatMessagePartTypeAudioURL
			p.Audio = &schema.MessageInputAudio{MessagePartCommon: schema.MessagePartCommon{URL: stringPtr(part.URL)}}
		case agentcore.ContentPartVideoURL:
			p.Type = schema.ChatMessagePartTypeVideoURL
			p.Video = &schema.MessageInputVideo{MessagePartCommon: schema.MessagePartCommon{URL: stringPtr(part.URL)}}
		case agentcore.ContentPartFileURL:
			p.Type = schema.ChatMessagePartTypeFileURL
			p.File = &schema.MessageInputFile{MessagePartCommon: schema.MessagePartCommon{URL: stringPtr(part.URL)}}
		}
		result = append(result, p)
	}
	return result
}

func adaptPartsFromRunner(parts []schema.MessageInputPart) []agentcore.ContentPart {
	result := make([]agentcore.ContentPart, 0, len(parts))
	for _, part := range parts {
		p := agentcore.ContentPart{Text: part.Text}
		switch part.Type {
		case schema.ChatMessagePartTypeText:
			p.Type = agentcore.ContentPartText
		case schema.ChatMessagePartTypeImageURL:
			p.Type = agentcore.ContentPartImageURL
			if part.Image != nil {
				if part.Image.URL != nil {
					p.URL = *part.Image.URL
				}
				if part.Image.Base64Data != nil {
					p.Base64Data = *part.Image.Base64Data
				}
				p.MIMEType = part.Image.MIMEType
				p.Detail = string(part.Image.Detail)
			}
		case schema.ChatMessagePartTypeAudioURL:
			p.Type = agentcore.ContentPartAudioURL
			if part.Audio != nil && part.Audio.URL != nil {
				p.URL = *part.Audio.URL
			}
		case schema.ChatMessagePartTypeVideoURL:
			p.Type = agentcore.ContentPartVideoURL
			if part.Video != nil && part.Video.URL != nil {
				p.URL = *part.Video.URL
			}
		case schema.ChatMessagePartTypeFileURL:
			p.Type = agentcore.ContentPartFileURL
			if part.File != nil && part.File.URL != nil {
				p.URL = *part.File.URL
			}
		}
		result = append(result, p)
	}
	return result
}

func adaptChatPartsForRunner(parts []agentcore.ContentPart) []schema.ChatMessagePart {
	result := make([]schema.ChatMessagePart, 0, len(parts))
	for _, part := range parts {
		p := schema.ChatMessagePart{Type: schema.ChatMessagePartType(part.Type), Text: part.Text}
		switch part.Type {
		case agentcore.ContentPartImageURL:
			p.ImageURL = &schema.ChatMessageImageURL{
				URL:      part.URL,
				Detail:   schema.ImageURLDetail(part.Detail),
				MIMEType: part.MIMEType,
			}
			if part.Base64Data != "" {
				p.ImageURL.URL = "data:" + part.MIMEType + ";base64," + part.Base64Data
			}
		case agentcore.ContentPartAudioURL:
			p.AudioURL = &schema.ChatMessageAudioURL{URL: part.URL, MIMEType: part.MIMEType}
		case agentcore.ContentPartVideoURL:
			p.VideoURL = &schema.ChatMessageVideoURL{URL: part.URL, MIMEType: part.MIMEType}
		case agentcore.ContentPartFileURL:
			p.FileURL = &schema.ChatMessageFileURL{URL: part.URL, MIMEType: part.MIMEType}
		}
		result = append(result, p)
	}
	return result
}

func adaptChatPartsFromRunner(parts []schema.ChatMessagePart) []agentcore.ContentPart {
	result := make([]agentcore.ContentPart, 0, len(parts))
	for _, part := range parts {
		p := agentcore.ContentPart{Type: agentcore.ContentPartType(part.Type), Text: part.Text}
		switch part.Type {
		case schema.ChatMessagePartTypeImageURL:
			p.Type = agentcore.ContentPartImageURL
			if part.ImageURL != nil {
				p.URL = part.ImageURL.URL
				p.MIMEType = part.ImageURL.MIMEType
				p.Detail = string(part.ImageURL.Detail)
			}
		case schema.ChatMessagePartTypeAudioURL:
			p.Type = agentcore.ContentPartAudioURL
			if part.AudioURL != nil {
				p.URL = part.AudioURL.URL
				p.MIMEType = part.AudioURL.MIMEType
			}
		case schema.ChatMessagePartTypeVideoURL:
			p.Type = agentcore.ContentPartVideoURL
			if part.VideoURL != nil {
				p.URL = part.VideoURL.URL
				p.MIMEType = part.VideoURL.MIMEType
			}
		case schema.ChatMessagePartTypeFileURL:
			p.Type = agentcore.ContentPartFileURL
			if part.FileURL != nil {
				p.URL = part.FileURL.URL
				p.MIMEType = part.FileURL.MIMEType
			}
		}
		result = append(result, p)
	}
	return result
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
