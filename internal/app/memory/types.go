package memory

import (
	domainmemory "fkteams/internal/domain/memory"
	memoryport "fkteams/internal/ports/memory"
)

type MemoryType = domainmemory.MemoryType

const (
	Preference = domainmemory.Preference
	Fact       = domainmemory.Fact
	Feedback   = domainmemory.Feedback
	Lesson     = domainmemory.Lesson
	Decision   = domainmemory.Decision
	Insight    = domainmemory.Insight
	Experience = domainmemory.Experience
)

var AllMemoryTypes = domainmemory.AllMemoryTypes

type TypeMeta = domainmemory.TypeMeta
type MemoryEntry = domainmemory.MemoryEntry
type Message = domainmemory.Message
type LLMClient = memoryport.LLMClient

// typeOrder 类型展示顺序，injector 和 markdown 共用。
var typeOrder = domainmemory.TypeOrder()
