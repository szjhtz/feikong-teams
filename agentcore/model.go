package agentcore

import (
	domainmessage "fkteams/internal/domain/message"
	runtimeport "fkteams/internal/ports/runtime"
)

type ChatModel = runtimeport.ChatModel
type ModelCall = runtimeport.ModelCall
type MessageStream = runtimeport.MessageStream

func NewMessageStream(messages []Message) MessageStream {
	return runtimeport.NewMessageStream([]domainmessage.Message(messages))
}
