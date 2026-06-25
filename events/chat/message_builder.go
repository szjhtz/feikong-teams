// Package chat 保留旧输入构建入口，实际实现位于 internal/app/chat。
package chat

import (
	"fkteams/agentcore"
	"fkteams/appstate"
	"fkteams/engine"
	appchat "fkteams/internal/app/chat"

	eventlog "fkteams/events/log"
)

func BuildTurnInput(recorder *eventlog.HistoryRecorder, userInput string) engine.TurnInput {
	return appchat.BuildTurnInput(recorder, userInput)
}

func BuildTurnInputWithMemory(recorder *eventlog.HistoryRecorder, userInput string, manager appstate.MemoryManager) engine.TurnInput {
	return appchat.BuildTurnInputWithMemory(recorder, userInput, manager)
}

func BuildMultimodalTurnInput(recorder *eventlog.HistoryRecorder, textContent string, parts []agentcore.ContentPart) engine.TurnInput {
	return appchat.BuildMultimodalTurnInput(recorder, textContent, parts)
}

func BuildMultimodalTurnInputWithMemory(recorder *eventlog.HistoryRecorder, textContent string, parts []agentcore.ContentPart, manager appstate.MemoryManager) engine.TurnInput {
	return appchat.BuildMultimodalTurnInputWithMemory(recorder, textContent, parts, manager)
}

var TextPart = appchat.TextPart
var ImageURLPart = appchat.ImageURLPart
var ImageBase64Part = appchat.ImageBase64Part
var AudioURLPart = appchat.AudioURLPart
var VideoURLPart = appchat.VideoURLPart
var FileURLPart = appchat.FileURLPart
var ExtractTextFromParts = appchat.ExtractTextFromParts
